package Server

import (
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
)

// GetOrgs
func (h *Handler) GetOrgs(c echo.Context) error {
	var organizaciones []*ORM.Organizacion
	h.DB.Preload("Procesos").Preload("Clientes").Preload("Usuarios").Preload("Procesos.TicketsProcesos").Preload("Procesos.TicketsProcesos.Tipo").Find(&organizaciones)
	return c.JSON(200, organizaciones)
}

// PatchJobHistory
func (h *Handler) PatchJobHistory(c echo.Context) error {
	var wg sync.WaitGroup
	orgs := new(ORM.Organizacion).GetAll(h.DB)

	for _, org := range orgs {
		FoldersAndProcesses := groupProcessesByFolderID(org.Procesos)
		for folderID, processes := range FoldersAndProcesses {
			wg.Add(1)
			go h.handleFolder(org, folderID, processes, &wg)
		}
	}
	wg.Wait()
	return c.JSON(200, "OK")
}

func groupProcessesByFolderID(procesos []*ORM.Proceso) map[uint][]*ORM.Proceso {
	result := make(map[uint][]*ORM.Proceso)
	for _, process := range procesos {
		result[process.Folderid] = append(result[process.Folderid], process)
	}
	return result
}

func (h *Handler) handleFolder(org *ORM.Organizacion, folderID uint, processes []*ORM.Proceso, wg *sync.WaitGroup) {
	defer wg.Done()

	JobsHistoryResponse := new(UipathAPI.JobsResponse)
	LogsHistoryResponse := new(UipathAPI.LogResponse)
	_ = org.GetFromApi(LogsHistoryResponse, int(folderID))
	_ = org.GetFromApi(JobsHistoryResponse, int(folderID))

	existingJobs := h.getExistingJobs(JobsHistoryResponse)
	h.updateOrCreateJobs(existingJobs, JobsHistoryResponse, processes)
	h.createLogs(LogsHistoryResponse)
}

func (h *Handler) getExistingJobs(response *UipathAPI.JobsResponse) map[int]*ORM.JobHistory {
	JobIds := make([]int, 0)
	for _, value := range response.Value {
		JobIds = append(JobIds, value.ID)
	}
	var existingJobs []*ORM.JobHistory
	h.DB.Where("job_id IN ?", JobIds).Find(&existingJobs)

	jobMap := make(map[int]*ORM.JobHistory)
	for _, job := range existingJobs {
		jobMap[job.JobID] = job
	}
	return jobMap
}

func (h *Handler) updateOrCreateJobs(existingJobs map[int]*ORM.JobHistory, response *UipathAPI.JobsResponse, processes []*ORM.Proceso) {
	newJobs := make([]*ORM.JobHistory, 0)

	// Construimos un mapa para acceder a los procesos directamente
	nameToProcessMap := make(map[string]*ORM.Proceso)
	for _, process := range processes {
		nameToProcessMap[process.Nombre] = process
	}

	for _, value := range response.Value {
		if process, exists := nameToProcessMap[value.ReleaseName]; exists {
			hostMachineName := ""
			if value.HostMachineName != nil {
				hostMachineName = *value.HostMachineName
			}

			jobEntry := &ORM.JobHistory{
				Proceso:          process,
				ProcesoID:        process.ID,
				CreationTime:     value.CreationTime,
				StartTime:        value.StartTime,
				EndTime:          value.EndTime,
				HostMachineName:  hostMachineName,
				Source:           value.Source,
				State:            value.State,
				JobKey:           value.Key,
				JobID:            value.ID,
				MonitorException: false,
				Duration:         value.EndTime.Sub(value.StartTime),
			}

			if existingJob, found := existingJobs[value.ID]; found {
				if existingJob.State != jobEntry.State {
					h.DB.Model(existingJob).Updates(jobEntry)
					existingJob.MonitorException = false
					h.DB.Save(existingJob)
				}
			} else {
				newJobs = append(newJobs, jobEntry)
			}
		}
	}

	if len(newJobs) > 0 {
		h.DB.Create(&newJobs)
	}
}

func (h *Handler) createLogs(response *UipathAPI.LogResponse) {
	newLogs := make([]*ORM.LogJobHistory, 0)

	// Pre-fetch jobs to reduce DB lookups
	var allJobs []*ORM.JobHistory
	h.DB.Find(&allJobs)
	jobKeyMap := make(map[string]*ORM.JobHistory)
	for _, job := range allJobs {
		jobKeyMap[job.JobKey] = job
	}

	// Check logs without touching the DB
	existingLogsMap := make(map[string]bool)
	var existingLogs []ORM.LogJobHistory
	h.DB.Find(&existingLogs, "job_key IN (?)", keysFromResponse(response))
	for _, log := range existingLogs {
		existingLogsMap[log.RawMessage] = true
	}

	for _, value := range response.Value {
		if job, exists := jobKeyMap[value.JobKey]; exists {
			// Check if log already exists using the map
			if _, found := existingLogsMap[value.RawMessage]; !found {
				newEntry := &ORM.LogJobHistory{
					Level:           value.Level,
					WindowsIdentity: value.WindowsIdentity,
					ProcessName:     value.ProcessName,
					TimeStamp:       value.TimeStamp,
					Message:         value.Message,
					JobKey:          value.JobKey,
					RawMessage:      value.RawMessage,
					RobotName:       value.RobotName,
					HostMachineName: value.HostMachineName,
					MachineId:       value.MachineId,
					MachineKey:      value.MachineKey,
					JobID:           int(job.ID),
					Job:             job,
				}
				newLogs = append(newLogs, newEntry)
			}
		}
	}

	// Bulk insert the new logs
	if len(newLogs) > 0 {
		h.DB.Create(&newLogs)
	}
}

func keysFromResponse(response *UipathAPI.LogResponse) []string {
	keys := make([]string, len(response.Value))
	for i, value := range response.Value {
		keys[i] = value.JobKey
	}
	return keys
}

// UpdateExceptionJob
func (h *Handler) UpdateExceptionJob(c echo.Context) error {
	// query JobKey
	JobKey := c.QueryParam("JobKey")
	Job := new(ORM.JobHistory)
	h.DB.Where("job_key = ?", JobKey).First(&Job)
	if Job.ID == 0 {
		return c.JSON(404, "Not Found")
	}
	Job.MonitorException = true
	h.DB.Save(&Job)
	return c.JSON(200, "OK")
}
