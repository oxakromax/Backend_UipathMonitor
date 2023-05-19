package Server

import (
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
)

// GetOrgs
func (H *Handler) GetOrgs(c echo.Context) error {
	Org := new(ORM.Organizacion)
	return c.JSON(200, Org.GetAll(H.Db))
}
