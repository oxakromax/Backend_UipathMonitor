package Server

import (
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
)

// GetOrgs
func (H *Handler) GetOrgs(c echo.Context) error {
	var organizaciones []*ORM.Organizacion
	H.Db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").Preload("Procesos.TicketsProcesos").Find(&organizaciones)
	return c.JSON(200, organizaciones)
}
