package Server

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
)

func (H *Handler) CheckRoleMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, route := range H.UniversalRoutes {
				if strings.EqualFold(route, c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			// Verificar si el usuario está autenticado y tiene un rol permitido
			userClaim, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return echo.ErrUnauthorized
			}

			claims, ok := userClaim.Claims.(jwt.MapClaims)
			if !ok {
				return echo.ErrUnauthorized
			}

			idFloat, ok := claims["id"].(float64)
			if !ok {
				return echo.ErrUnauthorized
			}

			id := uint(idFloat)
			for _, route := range H.UserUniversalRoutes {
				if strings.EqualFold(route, c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			User := ORM.Usuario{}
			User.Get(H.Db, uint(id))              // Obtener el usuario de la base de datos
			for _, UserRole := range User.Roles { // Iterar sobre los roles del usuario
				for _, route := range UserRole.Rutas {
					if strings.EqualFold(route.Route, c.Path()) && strings.EqualFold(route.Method, c.Request().Method) {
						return next(c) // Permitir el acceso al usuario si tiene el rol permitido
					}
				}
			}
			return echo.ErrUnauthorized // Acceso denegado si el usuario no tiene el rol permitido
		}
	}
}
