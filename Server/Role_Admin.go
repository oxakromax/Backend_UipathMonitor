package Server

import (
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
)

func (h *Handler) UpdateUipathProcess(c echo.Context) error {
	// if this route is reached, is a manual solicitation to update every process in the database, and add new ones detected
	Orgs := new(ORM.Organizacion).GetAll(h.DB)
	var wg = new(sync.WaitGroup)
	errorChannel := make(chan error, 1000)
	var errorList []error

	for _, Org := range Orgs {
		wg.Add(1)
		go func(Org *ORM.Organizacion) {
			defer wg.Done()
			var MapFolderProcess = make(map[int][]*ORM.Proceso)
			for _, proceso := range Org.Procesos {
				MapFolderProcess[int(proceso.Folderid)] = append(MapFolderProcess[int(proceso.Folderid)], proceso)
			}
			var SubWg = new(sync.WaitGroup)
			for FolderID, procesos := range MapFolderProcess {
				SubWg.Add(1)
				go func(FolderID int, procesos []*ORM.Proceso) {
					defer SubWg.Done()
					// Check if procesos exist in uipath, if it updates the name, if not, add [DELETED] to the name and Alias (if alias are empty, copy the name)
					ProcessUipath := new(UipathAPI.ReleasesResponse)
					err := Org.GetFromApi(ProcessUipath, FolderID)
					if err != nil {
						errorChannel <- err
						return
					}
					for _, proceso := range procesos {
						// Check if process exists in uipath
						exists := false
						h.DB.Find(proceso)
						for _, Process := range ProcessUipath.Value {
							if Process.ID == int(proceso.UipathProcessID) {
								exists = true
								modified := false
								if proceso.Nombre != Process.Name {
									proceso.Nombre = Process.Name
									modified = true
								}
								if proceso.Foldername != Process.OrganizationUnitFullyQualifiedName {
									proceso.Foldername = Process.OrganizationUnitFullyQualifiedName
									modified = true
								}
								if modified {
									h.DB.Save(proceso)
								}
								break
							}
						}
						if !exists {
							if strings.Contains(proceso.Nombre, "[DELETED]") && strings.Contains(proceso.Alias, "[DELETED]") {
								continue
							}
							if !strings.Contains(proceso.Nombre, "[DELETED]") {
								proceso.Nombre = "[DELETED] " + proceso.Nombre
							}
							if proceso.Alias == "" {
								proceso.Alias = proceso.Nombre
							} else {
								if !strings.Contains(proceso.Alias, "[DELETED]") {
									proceso.Alias = "[DELETED] " + proceso.Alias
								}
							}
							h.DB.Save(proceso)
						}
					}
				}(FolderID, procesos)
			}
			FoldersResponse := new(UipathAPI.FoldersResponse)
			err := Org.GetFromApi(FoldersResponse)
			if err != nil {
				errorChannel <- err
				return
			}
			for _, FolderIter := range FoldersResponse.Value {
				SubWg.Add(1)
				Folders := FolderIter
				go func() {
					defer SubWg.Done()
					Processes := new(UipathAPI.ReleasesResponse)
					err = Org.GetFromApi(Processes, Folders.ID)
					if err != nil {
						errorChannel <- err
						return
					}
					for _, Process := range Processes.Value {
						ORMProcess := new(ORM.Proceso)
						h.DB.Where("folderid = ? AND uipath_process_iD = ?", Folders.ID, Process.ID).First(ORMProcess)
						if ORMProcess.ID == 0 {
							ORMProcess = &ORM.Proceso{
								Nombre:           Process.Name,
								UipathProcessID:  uint(Process.ID),
								Folderid:         uint(Folders.ID),
								Foldername:       Folders.DisplayName,
								OrganizacionID:   Org.ID,
								WarningTolerance: 999,
								ErrorTolerance:   999,
								FatalTolerance:   999,
							}
							h.DB.Create(ORMProcess)
						}
					}
				}()
			}
			SubWg.Wait()
		}(Org)
	}
	wg.Wait()
	close(errorChannel)
	for err := range errorChannel {
		errorList = append(errorList, err)
	}

	if len(errorList) == 0 {
		return c.JSON(http.StatusOK, "OK")
	} else {
		errorSummary := make(map[string]int)
		for _, err := range errorList {
			errorSummary[err.Error()]++
		}
		return c.JSON(http.StatusInternalServerError, errorSummary)
	}
}
