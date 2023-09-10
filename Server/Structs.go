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
	DB                  *gorm.DB
	TokenKey            string
	UniversalRoutes     []string
	UserUniversalRoutes []string
	DBKey               string
}

func (h *Handler) RefreshDataBase(e *echo.Echo) {
	// Crear un mapa para las rutas de Echo
	echoRoutesMap := make(map[string]ORM.Route)
	for _, r := range e.Routes() {
		routeKey := r.Method + r.Path
		echoRoutesMap[routeKey] = ORM.Route{Route: r.Path, Method: r.Method}
	}

	// Obtener todas las rutas de la base de datos
	var dbRoutes []ORM.Route
	h.DB.Find(&dbRoutes)

	// Mapa para las rutas de la base de datos
	dbRoutesMap := make(map[string]ORM.Route)
	for _, r := range dbRoutes {
		routeKey := r.Method + r.Route
		dbRoutesMap[routeKey] = r
	}

	// Procesar y guardar rutas
	for key, route := range echoRoutesMap {
		if _, exists := dbRoutesMap[key]; !exists {
			// Si la ruta no existe en la base de datos, agregarla
			h.DB.Create(&route)
		}
		// (No es necesario agregarlo a la lista, porque ya está en la base de datos o se ha creado)
	}

	// Eliminar las rutas antiguas
	for key, route := range dbRoutesMap {
		if _, exists := echoRoutesMap[key]; !exists {
			// Si la ruta no existe en las rutas de Echo, eliminarla definitivamente
			h.DB.Exec("DELETE FROM roles_routes WHERE route_id = ?", route.ID)
			h.DB.Exec("DELETE FROM routes WHERE id = ?", route.ID)
		}
	}

	// Obtener todas las rutas de la base de datos
	routes := new(ORM.Route).GetAll(h.DB)
	// Crear o actualizar el rol de administrador con las rutas definidas en Echo
	adminRole := ORM.Rol{
		Nombre:      "admin",
		Description: "El rol de administrador tiene acceso a todas las rutas del sistema, así como permisos universales por sobre los demás roles.",
	}
	checkAdminRole := new(ORM.Rol)
	h.DB.Where("nombre = ?", "admin").First(&checkAdminRole)
	if checkAdminRole.ID == 0 {
		// Si el rol de administrador no existe, crearlo y agregarle las rutas
		h.DB.Create(&adminRole)
	} else {
		// Si el rol de administrador ya existe, actualizar sus rutas
		adminRole = *checkAdminRole
	}
	adminRole.Rutas = routes
	// Reemplazar las rutas del rol de administrador con las rutas actualizadas
	_ = h.DB.Model(&adminRole).Association("Rutas").Replace(adminRole.Rutas)
	AdminUser := ORM.Usuario{}
	// Obtener el usuario 1 (administrador) de la base de datos
	h.DB.First(&AdminUser, 1) // El usuario administrador siempre tendrá el ID 1
	if AdminUser.ID == 0 {
		// Si el usuario administrador no existe, crearlo
		Procesos := new(ORM.Proceso).GetAll(h.DB)
		Orgs := new(ORM.Organizacion).GetAll(h.DB)
		AdminUser = ORM.Usuario{
			Nombre:   "admin",
			Apellido: "admin",
			Email:    "admin@admin.cl",
		}
		AdminUser.SetPassword("test")
		h.DB.Save(&AdminUser)
		_ = h.DB.Model(&AdminUser).Association("Roles").Replace([]ORM.Rol{adminRole})
		_ = h.DB.Model(&AdminUser).Association("Procesos").Replace(Procesos)
		_ = h.DB.Model(&AdminUser).Association("Organizaciones").Replace(Orgs)
	}
	// Encriptar datos sensibles de las organizaciones
	orgs := new(ORM.Organizacion).GetAll(h.DB)
	for _, org := range orgs {
		_, err := functions.DecryptAES(h.DBKey, org.AppID)
		if err != nil {
			org.AppID, _ = functions.EncryptAES(h.DBKey, org.AppID)
		}
		_, err = functions.DecryptAES(h.DBKey, org.AppSecret)
		if err != nil {
			org.AppSecret, _ = functions.EncryptAES(h.DBKey, org.AppSecret)
		}
		if err == nil {
			continue
		}
		h.DB.Save(org)
	}
	type RoleDefinition struct {
		Name        string
		Description string
	}

	roleDefinitions := []RoleDefinition{
		{
			Name:        "organization",
			Description: "El rol de organización tiene acceso a modificar, crear, eliminar y ver las organizaciones del sistema en las cuales este sea incluido o sea propietario. Nota: Solo un administrador del sistema puede ver y gestionar todas las organizaciones sin ser parte de ellas.",
		},
		{
			Name:        "user",
			Description: "El rol de usuario tiene acceso a las areas comunes de usuario, como la administración de sus incidentes asignados y/o editar su perfil.",
		},
		{
			Name:        "user_administration",
			Description: "El rol de administración de usuarios tiene acceso a modificar, crear, eliminar y ver los usuarios del sistema así como otorgar roles por debajo del rol de administrador.",
		},
		{
			Name:        "processes_administration",
			Description: "El rol de administración de procesos tiene acceso a modificar, crear, eliminar y ver los procesos del sistema, asi como manejar los incidentes de los procesos, el usuario con este rol debe de estar asignado a las organizaciones donde administrará los procesos.",
		},
		{
			Name:        "monitor",
			Description: "El rol de monitor es un rol privado del sistema, el cual permite a un servicio externo acceder a las rutas de monitorización del sistema.",
		},
		{
			Name:        "downloader",
			Description: "El rol de downloader es el rol que permite descargar ficheros excel a modo de informe de organizaciones, usuarios y procesos.",
		},
	}

	// Mapa para almacenar roles por nombre
	rolesMap := make(map[string]*ORM.Rol)
	// Crear o buscar roles
	for _, def := range roleDefinitions {
		role := &ORM.Rol{
			Nombre:      def.Name,
			Usuarios:    make([]*ORM.Usuario, 0),
			Rutas:       make([]*ORM.Route, 0),
			Description: def.Description,
		}
		h.DB.FirstOrCreate(&role, "nombre = ?", def.Name)
		rolesMap[def.Name] = role
	}

	// Asignar rutas a roles
	for _, route := range routes {
		if strings.Contains(route.Route, "/admin/organization") {
			rolesMap["organization"].Rutas = append(rolesMap["organization"].Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/user") {
			rolesMap["user"].Rutas = append(rolesMap["user"].Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/admin/users") {
			rolesMap["user_administration"].Rutas = append(rolesMap["user_administration"].Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/admin/processes") {
			rolesMap["processes_administration"].Rutas = append(rolesMap["processes_administration"].Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/monitor") {
			rolesMap["monitor"].Rutas = append(rolesMap["monitor"].Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/download") {
			rolesMap["downloader"].Rutas = append(rolesMap["downloader"].Rutas, route)
		}
	}

	// Reemplazar asociaciones en la base de datos
	for _, role := range rolesMap {
		_ = h.DB.Model(role).Association("Rutas").Replace(role.Rutas)
	}

	// Aquí puedes agregar el código para el usuario monitor si es necesario

	monitorUser := new(ORM.Usuario)
	Username := os.Getenv("MONITOR_USER")
	Password := os.Getenv("MONITOR_PASS")
	h.DB.Where("email = ?", Username).First(&monitorUser)
	if monitorUser.ID == 0 {
		monitorUser = &ORM.Usuario{
			Email:    Username,
			Nombre:   "Monitor de procesos",
			Apellido: "",
		}
		monitorUser.SetPassword(Password)
		h.DB.Create(&monitorUser)
		_ = h.DB.Model(&monitorUser).Association("Roles").Replace([]ORM.Rol{*rolesMap["monitor"]})
	} else {
		monitorUser.SetPassword(Password)
		h.DB.Save(&monitorUser)
	}
	// Check if the tickets types exist (table isn't empty)
	TicketsTypeBD := make([]*ORM.TicketsTipo, 0)
	h.DB.Find(&TicketsTypeBD)
	if len(TicketsTypeBD) > 0 {
		return // If the table isn't empty, don't do anything
	}
	TicketsType := []string{"Incidente", "Mejora", "Mantenimiento", "Otro"}
	for _, ticketType := range TicketsType {
		ticket := new(ORM.TicketsTipo)
		h.DB.Where("nombre = ?", ticketType).First(&ticket)
		if ticket.ID == 0 {
			Diagnostico := false
			if ticketType == "Incidente" {
				Diagnostico = true
			}
			ticket = &ORM.TicketsTipo{
				Nombre:              ticketType,
				NecesitaDiagnostico: Diagnostico,
			}
			h.DB.Create(&ticket)
		}
	}
}

func (h *Handler) PingAuth(c echo.Context) error {
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
func (h *Handler) GetTime(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"time": time.Now().UTC(),
	})
}

func (h *Handler) GetTicketsType(c echo.Context) error {
	TicketsType := make([]*ORM.TicketsTipo, 0)
	h.DB.Find(&TicketsType)
	MapNameID := make(map[string]uint)
	for _, ticketType := range TicketsType {
		MapNameID[ticketType.Nombre] = ticketType.ID
	}
	return c.JSON(http.StatusOK, MapNameID)
}
