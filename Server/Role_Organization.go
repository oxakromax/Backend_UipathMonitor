package Server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
)

func (h *Handler) CreateOrganization(c echo.Context) error {
	// Obtener la organización de la solicitud
	Organization := new(ORM.Organizacion)
	if err := c.Bind(Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Verificar si la organización ya existe en la base de datos
	checkOrganization := new(ORM.Organizacion)
	h.DB.Where("uipathname = ? and tenantname = ?", Organization.Uipathname, Organization.Tenantname).First(&checkOrganization)
	if checkOrganization.ID != 0 {
		return c.JSON(http.StatusConflict, "Organization already exists")
	}

	// Cifrar datos sensibles app_id y app_secret
	Organization.AppID, _ = functions.EncryptAES(h.DBKey, Organization.AppID)
	Organization.AppSecret, _ = functions.EncryptAES(h.DBKey, Organization.AppSecret)
	// Verificar que los datos son correctos
	err := Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath credentials")
	}
	// Guardar la organización en la base de datos
	h.DB.Create(&Organization)
	// Agregar a cada Administrador a la organización
	Admins := new(ORM.Usuario).GetAdmins(h.DB)
	for _, admin := range Admins {
		_ = h.DB.Model(&Organization).Association("Usuarios").Append(admin)
	}
	// Agregar al usuario que hace la solicitud a la organización
	User, err := h.GetUserJWT(c)
	if err != nil {
		return err
	}
	_ = h.DB.Model(&Organization).Association("Usuarios").Append(User)
	h.DB.Save(&Organization)
	go func() {
		time.Sleep(1 * time.Second)
		_ = h.UpdateUipathProcess(c)
	}() // Actualizar los procesos de la organización a través de la función UpdateUipathProcess
	return c.JSON(http.StatusOK, Organization)
}
func (h *Handler) GetOrganizations(c echo.Context) error {
	Organization := new(ORM.Organizacion)
	if c.QueryParam("id") != "" {
		// Obtener ID de la organización de la solicitud
		organizationID, err := strconv.Atoi(c.QueryParam("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid organization ID")
		}
		// Obtener la organización de la base de datos
		Organization.Get(h.DB, uint(organizationID))
		if Organization.ID == 0 {
			return c.JSON(http.StatusNotFound, "Organization not found")
		}
		for _, usuario := range Organization.Usuarios {
			usuario.Password = ""
		}
		return c.JSON(http.StatusOK, Organization)
	}
	// Obtener las organizaciones de la base de datos
	AllOrgs := Organization.GetAll(h.DB)
	if len(AllOrgs) == 0 {
		return c.JSON(http.StatusNotFound, "Organizations not found")
	}
	for _, org := range AllOrgs {
		for _, usuario := range org.Usuarios {
			usuario.Password = ""
		}
	}
	return c.JSON(http.StatusOK, AllOrgs)

}
func (h *Handler) UpdateOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(h.DB, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Actualizar los datos de la organización
	if err := c.Bind(&Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid data")
	}
	_, errDecryption1 := functions.DecryptAES(h.DBKey, Organization.AppID) // Verificar si los datos ya están cifrados
	_, errDecryption2 := functions.DecryptAES(h.DBKey, Organization.AppSecret)
	if errDecryption1 != nil || errDecryption2 != nil { // Si no estaban cifrados, significa que se actualizaron
		Organization.AppID, _ = functions.EncryptAES(h.DBKey, Organization.AppID) // Se encriptan primero
		Organization.AppSecret, _ = functions.EncryptAES(h.DBKey, Organization.AppSecret)
	}
	err = Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath Data")
	}

	// Guardar los datos actualizados de la organización en la base de datos
	h.DB.Updates(&Organization)
	return c.JSON(http.StatusOK, Organization)
}
func (h *Handler) DeleteOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(h.DB, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Eliminar la organización de la base de datos
	h.DB.Delete(&Organization)
	// Eliminar Procesos de la organización
	for _, proceso := range Organization.Procesos {
		h.DB.Delete(&proceso)
	}
	// Eliminar Clientes de la organización
	for _, cliente := range Organization.Clientes {
		h.DB.Delete(&cliente)
	}
	return c.JSON(http.StatusOK, "Organization deleted successfully")
}

func (h *Handler) CreateUpdateOrganizationClient(c echo.Context) error {
	Client := new(ORM.Cliente)
	if err := c.Bind(Client); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}
	// check if organization exists
	Organization := new(ORM.Organizacion)
	Organization.Get(h.DB, Client.OrganizacionID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check email doesn't exist
	NewClient := new(ORM.Cliente)
	h.DB.Where("email = ?", Client.Email).First(NewClient)
	if NewClient.ID != 0 && NewClient.ID != Client.ID {
		return c.JSON(http.StatusBadRequest, "Email already exists")
	}
	if int(Client.ID) == 0 {
		h.DB.Create(Client)
	} else {
		// check if client exists
		NewClient := new(ORM.Cliente)
		NewClient.Get(h.DB, Client.ID)
		if NewClient.ID == 0 {
			return c.JSON(http.StatusNotFound, "Client not found")
		}
		h.DB.Save(Client)
	}
	return c.JSON(http.StatusOK, Client)
}

func (h *Handler) DeleteOrganizationClient(c echo.Context) error {
	Client := new(ORM.Cliente)
	if err := c.Bind(Client); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}
	if Client.ID == 0 {
		return c.JSON(http.StatusBadRequest, "Invalid client ID")
	}
	// Check if organization exists
	Organization := new(ORM.Organizacion)
	Organization.Get(h.DB, Client.OrganizacionID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check if client exists
	Client.Get(h.DB, Client.ID)
	if Client.ID == 0 {
		return c.JSON(http.StatusNotFound, "Client not found")
	}
	h.DB.Delete(Client)
	return c.JSON(http.StatusOK, Client)
}

func (h *Handler) UpdateProcessAlias(c echo.Context) error {
	Process := new(ORM.Proceso)
	// Params id
	ProcessID, _ := strconv.Atoi(c.QueryParam("id"))
	Process.Get(h.DB, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(http.StatusNotFound, "Process not found")
	}
	// Params alias
	Process.Alias = c.QueryParam("alias")
	h.DB.Save(Process)
	return c.JSON(http.StatusOK, Process)
}

func (h *Handler) UpdateOrganizationUser(c echo.Context) error {
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
	Organization.Get(h.DB, Config.OrgID)
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Check if users exist
	for _, UserID := range Config.NewUsers {
		User := new(ORM.Usuario)
		User.Get(h.DB, UserID)
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
			err := h.DB.Model(Organization).Association("Usuarios").Append(User)
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
			err := h.DB.Model(Organization).Association("Usuarios").Delete(Organization.Usuarios[index])
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
		}
	}
	h.DB.Preload("Usuarios").Save(Organization)
	return c.JSON(http.StatusOK, "OK")
}
