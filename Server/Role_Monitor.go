package Server

import (
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
	"gorm.io/gorm"
)

// GetOrgs
func (H *Handler) GetOrgs(c echo.Context) error {
	var organizaciones []*ORM.Organizacion
	H.Db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").Preload("Procesos.TicketsProcesos").Find(&organizaciones)
	return c.JSON(200, organizaciones)
}

// PatchJobHistory
func (H *Handler) PatchJobHistory(c echo.Context) error {
	var wg sync.WaitGroup
	var JobHistory = make([]*ORM.JobHistory, 0)
	H.Db.Find(&JobHistory) // Get all job history at once

	var organizaciones = new(ORM.Organizacion).GetAll(H.Db)

	for _, org := range organizaciones {
		FoldersAndProcesses := make(map[uint][]*ORM.Proceso)
		for _, process := range org.Procesos {
			FoldersAndProcesses[process.Folderid] = append(FoldersAndProcesses[process.Folderid], process)
		}
		for folderID, processes := range FoldersAndProcesses {
			folderID := folderID
			processes := processes
			wg.Add(1)
			OrgCopy := org
			go func() {
				defer wg.Done()
				logHistory := new(UipathAPI.JobsResponse)
				_ = OrgCopy.GetFromApi(logHistory, int(folderID))
				for _, value := range logHistory.Value {
					for _, process := range processes {
						if process.Nombre == value.ReleaseName {
							NewEntry := ORM.JobHistory{
								Model:           gorm.Model{},
								Proceso:         process,
								ProcesoID:       process.ID,
								CreationTime:    value.CreationTime,
								StartTime:       value.StartTime,
								EndTime:         value.EndTime,
								HostMachineName: value.HostMachineName,
								Source:          value.Source,
								State:           value.State,
								JobKey:          value.Key,
								JobID:           value.ID,
								Excepcion:       false,
							}
							NewEntry.Duration = NewEntry.EndTime.Sub(NewEntry.StartTime)
							Founded := false
							for _, job := range JobHistory {
								if job.JobID == NewEntry.JobID {
									if job.State != NewEntry.State {
										OriginalExcepcion := job.Excepcion
										H.Db.Model(&job).Updates(NewEntry)
										job.Excepcion = OriginalExcepcion
										H.Db.Save(&job)
									}
									Founded = true
									break
								}

							}
							if !Founded {
								H.Db.Create(&NewEntry)
							}
						}
					}
				}
			}()
		}
	}
	wg.Wait()
	return c.JSON(200, "OK")
}

// UpdateExceptionJob
func (H *Handler) UpdateExceptionJob(c echo.Context) error {
	// query JobKey
	JobKey := c.QueryParam("JobKey")
	Job := new(ORM.JobHistory)
	H.Db.Where("job_key = ?", JobKey).First(&Job)
	if Job.ID == 0 {
		return c.JSON(404, "Not Found")
	}
	Job.Excepcion = true
	H.Db.Save(&Job)
	return c.JSON(200, "OK")
}
