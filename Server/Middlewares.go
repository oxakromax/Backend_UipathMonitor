package Server

import (
	"encoding/json"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
)

func (h *Handler) CheckDBState() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if the database is connected
			// if not then send a 503 Service Unavailable
			if h.DB.Error == nil {
				return next(c)
			}
			return echo.ErrServiceUnavailable
		}
	}
}

func (h *Handler) CheckRoleMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, route := range h.UniversalRoutes {
				if strings.EqualFold(route, c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			// Verificar si el usuario está autenticado y tiene un rol permitido
			User, err := h.GetUserJWT(c)
			if err != nil {
				return err
			}
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

func (h *Handler) GetUserJWT(c echo.Context) (*ORM.Usuario, error) {
	userClaim, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, echo.ErrUnauthorized
	}

	claims, ok := userClaim.Claims.(jwt.MapClaims)
	if !ok {
		return nil, echo.ErrUnauthorized
	}

	userMap, ok := claims["user"].(map[string]interface{})
	if !ok {
		return nil, echo.ErrUnauthorized
	}

	// Convertir el mapa a la estructura Usuario
	userBytes, err := json.Marshal(userMap)
	if err != nil {
		return nil, echo.ErrInternalServerError
	}

	var User *ORM.Usuario
	err = json.Unmarshal(userBytes, &User)
	if err != nil {
		return nil, echo.ErrInternalServerError
	}
	return User, nil
}
