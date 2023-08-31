package Server

import (
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"strconv"
)

// GetClientTicket
func (h *Handler) GetClientTicket(c echo.Context) error {
	// Query Params:
	// Email - Email del cliente
	// ID - Id Ticket
	//Obtener el cliente
	cliente := new(ORM.Cliente)
	email := c.QueryParam("email")
	h.DB.Where("Email = ?", email).First(&cliente)
	if cliente.ID == 0 {
		return c.JSON(404, "Cliente no encontrado")
	}
	cliente.Get(h.DB, cliente.ID)
	//Obtener el ticket
	ticket := new(ORM.TicketsProceso)
	NumericID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(500, err)
	}
	ticket.Get(h.DB, uint(NumericID))
	if ticket.ID == 0 {
		return c.JSON(404, "Ticket no encontrado")
	}
	// Verificar que el ticket pertenezca al cliente
	owned := false
	returnProcess := new(ORM.Proceso)
	for _, proceso := range cliente.Procesos {
		if proceso.ID == ticket.ProcesoID {
			owned = true
			returnProcess = proceso
			break
		}
	}
	if !owned {
		return c.JSON(403, "El ticket no pertenece al cliente")
	}
	returnProcess.Get(h.DB, returnProcess.ID)
	returnProcess.Organizacion = nil
	returnProcess.TicketsProcesos = nil
	returnProcess.TicketsProcesos = append(returnProcess.TicketsProcesos, ticket)
	returnProcess.JobsHistory = nil
	for _, usuario := range returnProcess.Usuarios {
		usuario.Password = ""
	}
	return c.JSON(200, returnProcess)

}
