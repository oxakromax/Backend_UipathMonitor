package Server

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"gorm.io/gorm"
)

type Handler struct {
	Db                  *gorm.DB
	TokenKey            string
	UniversalRoutes     []string
	UserUniversalRoutes []string
	DbKey               string
	DbState             bool
}

func (H *Handler) RefreshDataBase(e *echo.Echo) {
	// Crear una lista de rutas a partir de las rutas definidas en Echo
	routes := new([]*ORM.Route)
	for _, r := range e.Routes() {
		// Crear un objeto Route a partir de la ruta en Echo
		route := ORM.Route{Route: r.Path, Method: r.Method}
		// Verificar si la ruta ya existe en la base de datos
		checkRoute := new(ORM.Route)
		H.Db.Where("route = ? AND method = ?", r.Path, r.Method).First(&checkRoute)
		if checkRoute.ID == 0 {
			// Si la ruta no existe, crearla y agregarla a la lista de rutas
			H.Db.Create(&route)
			*routes = append(*routes, &route)
		} else {
			// Si la ruta ya existe, agregarla a la lista de rutas
			*routes = append(*routes, checkRoute)
		}
	}
	// Guardar las rutas en la base de datos
	H.Db.Save(&routes)

	// Eliminar las rutas antiguas de la base de datos
	DbRoutes := new(ORM.Route).GetAll(H.Db)
	for _, route := range DbRoutes {
		found := false
		// Verificar si la ruta de la base de datos aún existe en las rutas definidas en Echo
		for _, newRoute := range *routes {
			if route.ID == newRoute.ID {
				found = true
				break
			}
		}
		if !found {
			// Si la ruta ya no existe en Echo, eliminar la relación entre la ruta y los roles, y luego eliminar la ruta
			H.Db.Exec("DELETE FROM roles_routes WHERE route_id = ?", route.ID)
			H.Db.Exec("DELETE FROM routes WHERE id = ?", route.ID)
		}
	}

	// Crear o actualizar el rol de administrador con las rutas definidas en Echo
	adminRole := ORM.Rol{
		Nombre:      "admin",
		Description: "El rol de administrador tiene acceso a todas las rutas del sistema. ",
	}
	checkAdminRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "admin").First(&checkAdminRole)
	if checkAdminRole.ID == 0 {
		// Si el rol de administrador no existe, crearlo y agregarle las rutas
		H.Db.Create(&adminRole)
	} else {
		// Si el rol de administrador ya existe, actualizar sus rutas
		adminRole = *checkAdminRole
	}
	adminRole.Rutas = *routes
	// Reemplazar las rutas del rol de administrador con las rutas actualizadas
	_ = H.Db.Model(&adminRole).Association("Rutas").Replace(adminRole.Rutas)
	AdminUser := ORM.Usuario{}
	AdminUser.GetByEmail(H.Db, "admin@admin.cl")
	if AdminUser.ID == 0 {
		// Si el usuario administrador no existe, crearlo
		Procesos := new(ORM.Proceso).GetAll(H.Db)
		Orgs := new(ORM.Organizacion).GetAll(H.Db)
		AdminUser = ORM.Usuario{
			Nombre:   "admin",
			Apellido: "admin",
			Email:    "admin@admin.cl",
		}
		AdminUser.SetPassword("test")
		H.Db.Save(&AdminUser)
		_ = H.Db.Model(&AdminUser).Association("Roles").Replace([]ORM.Rol{adminRole})
		_ = H.Db.Model(&AdminUser).Association("Procesos").Replace(Procesos)
		_ = H.Db.Model(&AdminUser).Association("Organizaciones").Replace(Orgs)
	}
	// Encriptar datos sensibles de las organizaciones
	orgs := new(ORM.Organizacion).GetAll(H.Db)
	for _, org := range orgs {
		_, err := functions.DecryptAES(H.DbKey, org.AppID)
		if err != nil {
			org.AppID, _ = functions.EncryptAES(H.DbKey, org.AppID)
		}
		_, err = functions.DecryptAES(H.DbKey, org.AppSecret)
		if err != nil {
			org.AppSecret, _ = functions.EncryptAES(H.DbKey, org.AppSecret)
		}
		if err == nil {
			continue
		}
		H.Db.Save(&org)
	}
	// Crear roles por Rutas, "/admin/organization", descripción: "El rol de organización tiene acceso a modificar, crear, eliminar y ver las organizaciones del sistema."
	organizationRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "organization").First(&organizationRole)
	if organizationRole.ID == 0 {
		organizationRole = &ORM.Rol{
			Nombre:      "organization",
			Description: "El rol de organización tiene acceso a modificar, crear, eliminar y ver las organizaciones del sistema.",
		}
		H.Db.Create(&organizationRole)
	}
	userRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "user").First(&userRole)
	if userRole.ID == 0 {
		userRole = &ORM.Rol{
			Nombre:      "user",
			Description: "El rol de usuario tiene acceso a las areas comunes de usuario, como la administración de sus incidentes asignados y/o editar su perfil.",
		}
		H.Db.Create(&userRole)
	}
	userAdministrationRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "user_administration").First(&userAdministrationRole)
	if userAdministrationRole.ID == 0 {
		userAdministrationRole = &ORM.Rol{
			Nombre:      "user_administration",
			Description: "El rol de administración de usuarios tiene acceso a modificar, crear, eliminar y ver los usuarios del sistema.",
		}
		H.Db.Create(&userAdministrationRole)
	}
	processesAdministrationRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "processes_administration").First(&processesAdministrationRole)
	if processesAdministrationRole.ID == 0 {
		processesAdministrationRole = &ORM.Rol{
			Nombre:      "processes_administration",
			Description: "El rol de administración de procesos tiene acceso a modificar, crear, eliminar y ver los procesos del sistema, asi como manejar los incidentes de los procesos, el usuario con este rol debe de estar asignado a las organizaciones donde administrará los procesos.",
		}
		H.Db.Create(&processesAdministrationRole)
	}
	monitorRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "monitor").First(&monitorRole)
	if monitorRole.ID == 0 {
		monitorRole = &ORM.Rol{
			Nombre:      "monitor",
			Description: "El rol de monitor es un rol privado del sistema, el cual permite a un servicio externo acceder a las rutas de monitorización del sistema.",
		}
		H.Db.Create(&monitorRole)
	}
	DownloaderRole := new(ORM.Rol)
	H.Db.Where("nombre = ?", "downloader").First(&DownloaderRole)
	if DownloaderRole.ID == 0 {
		DownloaderRole = &ORM.Rol{
			Nombre:      "downloader",
			Description: "El rol de downloader es el rol que permite descargar ficheros excel a modo de informe de organizaciones, usuarios y procesos.",
		}
		H.Db.Create(&DownloaderRole)
	}

	for _, route := range *routes {
		if strings.Contains(route.Route, "/admin/organization") {
			organizationRole.Rutas = append(organizationRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/user") {
			userRole.Rutas = append(userRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/admin/users") {
			userAdministrationRole.Rutas = append(userAdministrationRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/admin/processes") {
			processesAdministrationRole.Rutas = append(processesAdministrationRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/monitor") {
			monitorRole.Rutas = append(monitorRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/download") {
			DownloaderRole.Rutas = append(DownloaderRole.Rutas, route)
		}
	}
	_ = H.Db.Model(&organizationRole).Association("Rutas").Replace(organizationRole.Rutas)
	_ = H.Db.Model(&userRole).Association("Rutas").Replace(userRole.Rutas)
	_ = H.Db.Model(&userAdministrationRole).Association("Rutas").Replace(userAdministrationRole.Rutas)
	_ = H.Db.Model(&processesAdministrationRole).Association("Rutas").Replace(processesAdministrationRole.Rutas)
	_ = H.Db.Model(&monitorRole).Association("Rutas").Replace(monitorRole.Rutas)
	_ = H.Db.Model(&DownloaderRole).Association("Rutas").Replace(DownloaderRole.Rutas)
	// Crear usuario monitor, sino existe, sobre escribir contraseña
	monitorUser := new(ORM.Usuario)
	Username := os.Getenv("MONITOR_USER")
	Password := os.Getenv("MONITOR_PASS")
	H.Db.Where("email = ?", Username).First(&monitorUser)
	if monitorUser.ID == 0 {
		monitorUser = &ORM.Usuario{
			Email:    Username,
			Nombre:   "Monitor de procesos",
			Apellido: "",
		}
		monitorUser.SetPassword(Password)
		H.Db.Create(&monitorUser)
		_ = H.Db.Model(&monitorUser).Association("Roles").Replace([]ORM.Rol{*monitorRole})
	} else {
		monitorUser.SetPassword(Password)
		H.Db.Save(&monitorUser)
	}
	// "Incidente": 1,
	// "Mejora": 2,
	// "Mantenimiento": 3,
	// "Otro": 4,
	TicketsType := []string{"Incidente", "Mejora", "Mantenimiento", "Otro"}
	for _, ticketType := range TicketsType {
		ticket := new(ORM.TicketsTipo)
		H.Db.Where("nombre = ?", ticketType).First(&ticket)
		if ticket.ID == 0 {
			Diagnostico := false
			if ticketType == "Incidente" {
				Diagnostico = true
			}
			ticket = &ORM.TicketsTipo{
				Nombre:              ticketType,
				NecesitaDiagnostico: Diagnostico,
			}
			H.Db.Create(&ticket)
		}
	}
}

func (H *Handler) PingAuth(c echo.Context) error {
	// Si el usuario no está autenticado, retornar error
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	// Si el usuario no está autenticado, retornar error
	if claims["id"] == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "No autorizado",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Autorizado",
	})
}

// GetTime
func (H *Handler) GetTime(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"time": time.Now().UTC(),
	})
}
