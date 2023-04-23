package Routes

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"strings"
)

func (H *Structs.Handler) RefreshDataBase(e *echo.Echo) {
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

	for _, route := range *routes {
		if strings.Contains(route.Route, "/admin/organization") {
			organizationRole.Rutas = append(organizationRole.Rutas, route)
		}
		if strings.HasPrefix(route.Route, "/user") {
			userRole.Rutas = append(userRole.Rutas, route)
		}
	}
	_ = H.Db.Model(&organizationRole).Association("Rutas").Replace(organizationRole.Rutas)
	_ = H.Db.Model(&userRole).Association("Rutas").Replace(userRole.Rutas)
}

func (H *Structs.Handler) CheckRoleMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, route := range H.UniversalRoutes {
				if strings.ToLower(route) == strings.ToLower(c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			// Verificar si el usuario está autenticado y tiene un rol permitido
			id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64)) // Extraer el ID del usuario del token JWT
			for _, route := range H.UserUniversalRoutes {
				if strings.ToLower(route) == strings.ToLower(c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			User := ORM.Usuario{}
			User.Get(H.Db, uint(id))              // Obtener el usuario de la base de datos
			for _, UserRole := range User.Roles { // Iterar sobre los roles del usuario
				for _, route := range UserRole.Rutas {
					if strings.ToLower(route.Route) == strings.ToLower(c.Path()) && strings.ToLower(route.Method) == strings.ToLower(c.Request().Method) {
						return next(c) // Permitir el acceso al usuario si tiene el rol permitido
					}
				}
			}
			return echo.ErrUnauthorized // Acceso denegado si el usuario no tiene el rol permitido
		}
	}
}
