package Routes

import (
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func (H *Structs.Handler) CreateOrganization(c echo.Context) error {
	// Obtener la organización de la solicitud
	Organization := new(ORM.Organizacion)
	if err := c.Bind(Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Verificar si la organización ya existe en la base de datos
	checkOrganization := new(ORM.Organizacion)
	H.Db.Where("uipathname = ? and tenantname = ?", Organization.Uipathname, Organization.Tenantname).First(&checkOrganization)
	if checkOrganization.ID != 0 {
		return c.JSON(http.StatusConflict, "Organization already exists")
	}

	// Cifrar datos sensibles app_id y app_secret
	Organization.AppID, _ = functions.EncryptAES(H.DbKey, Organization.AppID)
	Organization.AppSecret, _ = functions.EncryptAES(H.DbKey, Organization.AppSecret)
	// Verificar que los datos son correctos
	err := Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath credentials")
	}
	// Guardar la organización en la base de datos
	H.Db.Create(&Organization)
	// Agregar a cada Administrador a la organización
	Admins := new(ORM.Usuario).GetAdmins(H.Db)
	for _, admin := range Admins {
		_ = H.Db.Model(&Organization).Association("Usuarios").Append(admin)
	}
	JsonFolders := new(UipathAPI.FoldersResponse)
	err = Organization.GetFromApi(JsonFolders)
	if err != nil {
		for _, Folder := range JsonFolders.Value {
			IDFolder := Folder.Id
			JsonProcesses := new(UipathAPI.ReleasesResponse)
			err = Organization.GetFromApi(JsonProcesses, IDFolder)
			if err != nil {
				for _, Process := range JsonProcesses.Value {
					// Obtener el proceso de la base de datos
					ProcessDB := ORM.Proceso{
						Nombre:           Process.Name,
						Alias:            "",
						Folderid:         uint(IDFolder),
						Foldername:       Folder.DisplayName,
						OrganizacionID:   Organization.ID,
						WarningTolerance: 999, // 999 = no limit
						ErrorTolerance:   999, // 999 = no limit
						FatalTolerance:   999, // 999 = no limit
					}
					// Guardar el proceso en la base de datos
					H.Db.Create(&ProcessDB)
				}
			}

		}
	}
	return c.JSON(http.StatusOK, Organization)
}
func (H *Structs.Handler) GetOrganizations(c echo.Context) error {
	Organization := new(ORM.Organizacion)
	if c.QueryParam("id") != "" {
		// Obtener ID de la organización de la solicitud
		organizationID, err := strconv.Atoi(c.QueryParam("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid organization ID")
		}
		// Obtener la organización de la base de datos
		Organization.Get(H.Db, uint(organizationID))
		if Organization.ID == 0 {
			return c.JSON(http.StatusNotFound, "Organization not found")
		}
		for _, usuario := range Organization.Usuarios {
			usuario.Password = ""
		}
		return c.JSON(http.StatusOK, Organization)
	}
	// Obtener las organizaciones de la base de datos
	AllOrgs := Organization.GetAll(H.Db)
	if AllOrgs == nil || len(AllOrgs) == 0 {
		return c.JSON(http.StatusNotFound, "Organizations not found")
	}
	for _, org := range AllOrgs {
		for _, usuario := range org.Usuarios {
			usuario.Password = ""
		}
	}
	return c.JSON(http.StatusOK, AllOrgs)

}
func (H *Structs.Handler) UpdateOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Actualizar los datos de la organización
	if err := c.Bind(&Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid data")
	}
	_, errDecryption1 := functions.DecryptAES(H.DbKey, Organization.AppID) // Verificar si los datos ya están cifrados
	_, errDecryption2 := functions.DecryptAES(H.DbKey, Organization.AppSecret)
	if errDecryption1 != nil || errDecryption2 != nil { // Si no estaban cifrados, significa que se actualizaron
		Organization.AppID, _ = functions.EncryptAES(H.DbKey, Organization.AppID) // Se encriptan primero
		Organization.AppSecret, _ = functions.EncryptAES(H.DbKey, Organization.AppSecret)
	}
	err = Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath Data")
	}

	// Guardar los datos actualizados de la organización en la base de datos
	H.Db.Updates(&Organization)
	return c.JSON(http.StatusOK, Organization)
}
func (H *Structs.Handler) DeleteOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Eliminar la organización de la base de datos
	H.Db.Delete(&Organization)
	// Eliminar Procesos de la organización
	for _, proceso := range Organization.Procesos {
		H.Db.Delete(&proceso)
	}
	// Eliminar Clientes de la organización
	for _, cliente := range Organization.Clientes {
		H.Db.Delete(&cliente)
	}
	return c.JSON(http.StatusOK, "Organization deleted successfully")
}

func (H *Structs.Handler) CreateUpdateOrganizationClient(c echo.Context) error {
	Client := new(ORM.Cliente)
	if err := c.Bind(Client); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}
	// check if organization exists
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, Client.OrganizacionID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check email doesn't exist
	NewClient := new(ORM.Cliente)
	H.Db.Where("email = ?", Client.Email).First(NewClient)
	if NewClient.ID != 0 && NewClient.ID != Client.ID {
		return c.JSON(http.StatusBadRequest, "Email already exists")
	}
	if int(Client.ID) == 0 {
		H.Db.Create(Client)
	} else {
		// check if client exists
		NewClient := new(ORM.Cliente)
		NewClient.Get(H.Db, Client.ID)
		if NewClient.ID == 0 {
			return c.JSON(http.StatusNotFound, "Client not found")
		}
		H.Db.Save(Client)
	}
	return c.JSON(http.StatusOK, Client)
}

func (H *Structs.Handler) DeleteOrganizationClient(c echo.Context) error {
	Client := new(ORM.Cliente)
	if err := c.Bind(Client); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}
	if Client.ID == 0 {
		return c.JSON(http.StatusBadRequest, "Invalid client ID")
	}
	// Check if organization exists
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, Client.OrganizacionID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check if client exists
	Client.Get(H.Db, Client.ID)
	if Client.ID == 0 {
		return c.JSON(http.StatusNotFound, "Client not found")
	}
	H.Db.Delete(Client)
	return c.JSON(http.StatusOK, Client)
}

func (H *Structs.Handler) UpdateProcessAlias(c echo.Context) error {
	Process := new(ORM.Proceso)
	// Params id
	ProcessID, _ := strconv.Atoi(c.QueryParam("id"))
	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(http.StatusNotFound, "Process not found")
	}
	// Params alias
	Process.Alias = c.QueryParam("alias")
	H.Db.Save(Process)
	return c.JSON(http.StatusOK, Process)
}

func (H *Structs.Handler) UpdateOrganizationUser(c echo.Context) error {
	var Config = new(struct {
		OrgID       uint   `json:"org_id"`
		NewUsers    []uint `json:"new_users"`
		DeleteUsers []uint `json:"delete_users"`
	})
	if err := c.Bind(Config); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}
	// Check if organization exists
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, Config.OrgID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check if users exist
	for _, UserID := range Config.NewUsers {
		User := new(ORM.Usuario)
		User.Get(H.Db, UserID)
		if User.ID == 0 {
			return c.JSON(http.StatusNotFound, "User "+strconv.Itoa(int(UserID))+" not found")
		}
		// Check if user is already in organization
		isin := false
		for _, usuario := range Organization.Usuarios {
			if usuario.ID == User.ID {
				isin = true
			}
		}
		if !isin {
			err := H.Db.Model(Organization).Association("Usuarios").Append(User)
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
		}
	}
	// Check if users exist
	for _, UserID := range Config.DeleteUsers {
		// Check if user is already in organization
		isin := false
		index := 0
		for i, usuario := range Organization.Usuarios {
			if usuario.ID == UserID {
				isin = true
				index = i
				break
			}
		}
		if isin {
			err := H.Db.Model(Organization).Association("Usuarios").Delete(Organization.Usuarios[index])
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
		}
	}
	H.Db.Preload("Usuarios").Save(Organization)
	return c.JSON(http.StatusOK, "OK")
}

func (H *Structs.Handler) UpdateUipathProcess(c echo.Context) error {
	// if this route is reached, is a manual solicitation to update every process in the database, and add new ones detected
	Orgs := new(ORM.Organizacion).GetAll(H.Db)
	var wg = new(sync.WaitGroup)
	errorChannel := make(chan error, 1000)
	var errorList []error

	for _, Org := range Orgs {
		wg.Add(1)
		go func(Org *ORM.Organizacion) {
			var MapFolderProcess = make(map[int][]*ORM.Proceso)
			for _, proceso := range Org.Procesos {
				MapFolderProcess[int(proceso.Folderid)] = append(MapFolderProcess[int(proceso.Folderid)], proceso)
			}
			var SubWg = new(sync.WaitGroup)
			for FolderID, procesos := range MapFolderProcess {
				SubWg.Add(1)
				go func(FolderID int, procesos []*ORM.Proceso) {
					// Check if procesos exist in uipath, if it updates the name, if not, add [DELETED] to the name and Alias (if alias are empty, copy the name)
					ProcessUipath := new(UipathAPI.ReleasesResponse)
					err := Org.GetFromApi(ProcessUipath, FolderID)
					if err != nil {
						errorChannel <- err
						SubWg.Done()
						return
					}
					for _, proceso := range procesos {
						// Check if process exists in uipath
						exists := false
						H.Db.Find(proceso)
						for _, Process := range ProcessUipath.Value {
							if Process.Id == int(proceso.UipathProcessID) {
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
									H.Db.Save(proceso)
								}
								break
							}
						}
						if !exists {
							if strings.Contains(proceso.Nombre, "[DELETED]") {
								continue
							}
							proceso.Nombre = "[DELETED] " + proceso.Nombre
							if proceso.Alias == "" {
								proceso.Alias = proceso.Nombre
							} else {
								proceso.Alias = "[DELETED] " + proceso.Alias
							}
							H.Db.Save(proceso)
						}
					}
					SubWg.Done()
				}(FolderID, procesos)
			}
			FoldersResponse := new(UipathAPI.FoldersResponse)
			err := Org.GetFromApi(FoldersResponse)
			if err != nil {
				errorChannel <- err
				wg.Done()
				return
			}
			for _, FolderIter := range FoldersResponse.Value {
				SubWg.Add(1)
				Folders := FolderIter
				go func() {
					Processes := new(UipathAPI.ReleasesResponse)
					err = Org.GetFromApi(Processes, Folders.Id)
					if err != nil {
						errorChannel <- err
						SubWg.Done()
						return
					}
					for _, Process := range Processes.Value {
						ORMProcess := new(ORM.Proceso)
						H.Db.Where("folderid = ? AND uipath_process_iD = ?", Folders.Id, Process.Id).First(ORMProcess)
						if ORMProcess.ID == 0 {
							ORMProcess = &ORM.Proceso{
								Nombre:           Process.Name,
								UipathProcessID:  uint(Process.Id),
								Folderid:         uint(Folders.Id),
								Foldername:       Folders.DisplayName,
								OrganizacionID:   Org.ID,
								WarningTolerance: 999,
								ErrorTolerance:   999,
								FatalTolerance:   999,
							}
							H.Db.Create(ORMProcess)
						}
					}
					SubWg.Done()
				}()
			}
			SubWg.Wait()
			wg.Done()
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
