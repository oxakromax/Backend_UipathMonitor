package Server

import (
	"archive/zip"
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/xuri/excelize/v2"
)

func (H *Handler) GetOrgData(c echo.Context) error {

	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)

	if !User.HasRole("organization") && !User.HasRole("downloader") {
		return c.JSON(403, "You don't have permission to do this")
	}

	type OrgSummary struct {
		Organizacion      string
		UsuariosAsignados int
		ClientesAsignados int
		CantidadProcesos  int
		CantidadTickets   map[string]int
	}

	OrgID := c.QueryParam("id")
	OrgData := new(ORM.Organizacion)
	if OrgID != "" {
		OrgIDInt, err := strconv.Atoi(OrgID)
		if err != nil {
			return c.JSON(400, "Invalid organization ID")
		}
		OrgData.Get(H.Db, uint(OrgIDInt))
		if OrgData.ID == 0 {
			return c.JSON(404, "Organization not found")
		}
	}
	orgSummary := OrgSummary{
		Organizacion:      OrgData.Nombre,
		UsuariosAsignados: len(OrgData.Usuarios),
		ClientesAsignados: len(OrgData.Clientes),
		CantidadProcesos:  len(OrgData.Procesos),
		CantidadTickets:   make(map[string]int),
	}
	for _, proceso := range OrgData.Procesos {
		proceso.Get(H.Db, proceso.ID) // obtener toda la data del proceso
		for _, ticket := range proceso.TicketsProcesos {
			ticket.Get(H.Db, ticket.ID) // obtener toda la data del ticket
			orgSummary.CantidadTickets[ticket.Estado]++
		}
	}
	// Create a new Excel file
	file := excelize.NewFile()
	// Create sheet Summary
	sheetName := "Organization Data"

	file.SetSheetName("Sheet1", sheetName)
	// Set sheet title
	file.SetCellValue(sheetName, "A1", "Organization Data")
	// Set Org Summary Data headers in column A and Data in column B
	file.SetCellValue(sheetName, "A2", "Organization")
	file.SetCellValue(sheetName, "B2", orgSummary.Organizacion)
	file.SetCellValue(sheetName, "A3", "Assigned Users")
	file.SetCellValue(sheetName, "B3", orgSummary.UsuariosAsignados)
	file.SetCellValue(sheetName, "A4", "Assigned Clients")
	file.SetCellValue(sheetName, "B4", orgSummary.ClientesAsignados)
	file.SetCellValue(sheetName, "A5", "Processes")
	file.SetCellValue(sheetName, "B5", orgSummary.CantidadProcesos)
	file.SetCellValue(sheetName, "A6", "Tickets")
	ActualRow := 7
	for Key, Value := range orgSummary.CantidadTickets {
		file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), Key)
		file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), Value)
		ActualRow++
	}

	// Add Clients Sheet to Excel file
	sheetName = "Clients"
	file.NewSheet(sheetName)
	// Set sheet title
	file.SetCellValue(sheetName, "A1", "Clients")
	// Set Clients Data headers ID, Name, LastName, Email, Cantity of Processes in charge
	file.SetCellValue(sheetName, "A2", "ID")
	file.SetCellValue(sheetName, "B2", "Name")
	file.SetCellValue(sheetName, "C2", "Last Name")
	file.SetCellValue(sheetName, "D2", "Email")
	file.SetCellValue(sheetName, "E2", "Processes in charge")
	ActualRow = 3
	for _, cliente := range OrgData.Clientes {
		cliente.Get(H.Db, cliente.ID) // obtener toda la data del cliente
		file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), cliente.ID)
		file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), cliente.Nombre)
		file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), cliente.Apellido)
		file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), cliente.Email)
		file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), len(cliente.Procesos))
		ActualRow++
	}
	// Add Users Sheet to Excel file
	sheetName = "Users"
	file.NewSheet(sheetName)
	// Set sheet title
	file.SetCellValue(sheetName, "A1", "Users")
	// Set Users Data headers ID, Name, LastName, Email, Cantity of Processes in charge, Tickets resolved, Tickets Pending, Average Duration per Ticket
	file.SetCellValue(sheetName, "A2", "ID")
	file.SetCellValue(sheetName, "B2", "Name")
	file.SetCellValue(sheetName, "C2", "Last Name")
	file.SetCellValue(sheetName, "D2", "Email")
	file.SetCellValue(sheetName, "E2", "Processes in charge")
	file.SetCellValue(sheetName, "F2", "Tickets resolved")
	file.SetCellValue(sheetName, "G2", "Tickets pending")
	file.SetCellValue(sheetName, "H2", "Average Time per Ticket")
	ActualRow = 3
	for _, usuario := range OrgData.Usuarios {
		usuario.Get(H.Db, usuario.ID) // obtener toda la data del usuario
		file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), usuario.ID)
		file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), usuario.Nombre)
		file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), usuario.Apellido)
		file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), usuario.Email)
		file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), len(usuario.Procesos))

		TicketsResolved := 0
		TicketsPending := 0
		AverageDuration := new(time.Duration)
		Participation := 0
		for _, proceso := range OrgData.Procesos {
			proceso.Get(H.Db, proceso.ID) // obtener toda la data del proceso
			for _, ticketsProceso := range proceso.TicketsProcesos {
				ticketsProceso.Get(H.Db, ticketsProceso.ID) // obtener toda la data del ticket
				ParticipationBool := false
				if ticketsProceso.Estado == "Finalizado" {
					// MUST BE THE LAST USER IN THE TICKET
					if uint(ticketsProceso.Detalles[len(ticketsProceso.Detalles)-1].UsuarioID) == usuario.ID {
						TicketsResolved++
					}
				} else {
					TicketsPending++
				}
				for _, ticketsDetalle := range ticketsProceso.Detalles {
					if uint(ticketsDetalle.UsuarioID) == usuario.ID {
						// Must be the user that created the detail to count as participation
						ParticipationBool = true
						duration := ticketsDetalle.FechaFin.Sub(ticketsDetalle.FechaInicio)
						*AverageDuration += duration
					}
				}
				if ParticipationBool { // If the user participated in the ticket
					Participation++ // Add 1 to the participation counter
				}
			}
		}
		if Participation > 0 {
			*AverageDuration /= time.Duration(Participation)
		}
		file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), TicketsResolved)
		file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), TicketsPending)
		file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), AverageDuration)
		ActualRow++
	}
	// Add Processes Sheet to Excel file
	sheetName = "Processes"
	file.NewSheet(sheetName)
	// Set sheet title
	file.SetCellValue(sheetName, "A1", "Processes")
	// Set Processes Data headers ID, Name, Alias, Folder ID, Folder Name, Warning Tolerance, Error Tolerance, Fatal Tolerance, Active Monitoring, Priority, Max Queue Time, Tickets, UsersEmails, ClientsEmails
	file.SetCellValue(sheetName, "A2", "ID")
	file.SetCellValue(sheetName, "B2", "Name")
	file.SetCellValue(sheetName, "C2", "Alias")
	file.SetCellValue(sheetName, "D2", "Folder ID")
	file.SetCellValue(sheetName, "E2", "Folder Name")
	file.SetCellValue(sheetName, "F2", "Warning Tolerance")
	file.SetCellValue(sheetName, "G2", "Error Tolerance")
	file.SetCellValue(sheetName, "H2", "Fatal Tolerance")
	file.SetCellValue(sheetName, "I2", "Active Monitoring")
	file.SetCellValue(sheetName, "J2", "Priority")
	file.SetCellValue(sheetName, "K2", "Max Queue Time")
	file.SetCellValue(sheetName, "L2", "Tickets")
	file.SetCellValue(sheetName, "M2", "Users Emails") // separate emails with semicolon
	file.SetCellValue(sheetName, "N2", "Clients Emails")
	ActualRow = 3
	for _, proceso := range OrgData.Procesos {
		proceso.Get(H.Db, proceso.ID) // obtener toda la data del proceso
		file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), proceso.ID)
		file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), proceso.Nombre)
		file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), proceso.Alias)
		file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), proceso.Folderid)
		file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), proceso.Foldername)
		file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), proceso.WarningTolerance)
		file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), proceso.ErrorTolerance)
		file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), proceso.FatalTolerance)
		file.SetCellValue(sheetName, "I"+strconv.Itoa(ActualRow), proceso.ActiveMonitoring)
		file.SetCellValue(sheetName, "J"+strconv.Itoa(ActualRow), proceso.Priority)
		file.SetCellValue(sheetName, "K"+strconv.Itoa(ActualRow), proceso.MaxQueueTime)
		file.SetCellValue(sheetName, "L"+strconv.Itoa(ActualRow), len(proceso.TicketsProcesos))
		UsersEmails := ""
		for _, usuario := range proceso.Usuarios {
			UsersEmails += usuario.Email + ";"
		}
		file.SetCellValue(sheetName, "M"+strconv.Itoa(ActualRow), UsersEmails)
		ClientsEmails := ""
		for _, cliente := range proceso.Clientes {
			ClientsEmails += cliente.Email + ";"
		}
		file.SetCellValue(sheetName, "N"+strconv.Itoa(ActualRow), ClientsEmails)
		ActualRow++
	}

	// Add Tickets Sheet to Excel file
	sheetName = "Tickets"
	file.NewSheet(sheetName)
	// Set sheet title
	file.SetCellValue(sheetName, "A1", "Tickets")
	// add Headers: ID, ProcessID, Description, State, Priority, CreatorUserID,CreatedAt, StartDate, EndDate, Duration, Details_Count, type
	file.SetCellValue(sheetName, "A2", "ID")
	file.SetCellValue(sheetName, "B2", "Process ID")
	file.SetCellValue(sheetName, "C2", "Description")
	file.SetCellValue(sheetName, "D2", "State")
	file.SetCellValue(sheetName, "E2", "Priority")
	file.SetCellValue(sheetName, "F2", "Creator User ID")
	file.SetCellValue(sheetName, "G2", "Created At")
	file.SetCellValue(sheetName, "H2", "Start Date")
	file.SetCellValue(sheetName, "I2", "End Date")
	file.SetCellValue(sheetName, "J2", "Duration")
	file.SetCellValue(sheetName, "K2", "Details Count")
	file.SetCellValue(sheetName, "L2", "Type")
	ActualRow = 3
	for _, proceso := range OrgData.Procesos {
		proceso.Get(H.Db, proceso.ID) // obtener toda la data del proceso
		for _, ticket := range proceso.TicketsProcesos {
			ticket.Get(H.Db, ticket.ID) // obtener toda la data del ticket
			file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), ticket.ID)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), ticket.ProcesoID)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), ticket.Descripcion)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), ticket.Estado)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), ticket.Prioridad)
			file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), ticket.UsuarioCreadorID)
			file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), ticket.CreatedAt)
			var StartDate time.Time
			var EndDate time.Time
			var Duration time.Duration
			if len(ticket.Detalles) > 0 {
				StartDate = ticket.Detalles[0].FechaInicio
				if ticket.Estado == "Finalizado" {
					EndDate = ticket.Detalles[len(ticket.Detalles)-1].FechaFin
				} else {
					// Empty end date
					EndDate = time.Time{}
				}
				for _, detalle := range ticket.Detalles {
					Duration += detalle.FechaFin.Sub(detalle.FechaInicio)
				}
			}
			file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), StartDate)
			file.SetCellValue(sheetName, "I"+strconv.Itoa(ActualRow), EndDate)
			file.SetCellValue(sheetName, "J"+strconv.Itoa(ActualRow), Duration)
			file.SetCellValue(sheetName, "K"+strconv.Itoa(ActualRow), len(ticket.Detalles))
			file.SetCellValue(sheetName, "L"+strconv.Itoa(ActualRow), ticket.GetTipo(H.Db))
			ActualRow++
		}
	}
	// Write Excel file to memory buffer
	buf := new(bytes.Buffer)
	if err := file.Write(buf); err != nil {
		return c.JSON(500, "Failed to write organization data to memory")
	}

	// Set response headers
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=organization_data.xlsx")

	// Send Excel file buffer as response
	return c.Blob(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())

}

// GetUserData returns an excel file with the user data
// if there's more than 1 id in id query param separated by coma, it will return a zip with excel files of users requested
func (H *Handler) GetUserData(c echo.Context) error {
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if !User.HasRole("user_administration") && !User.HasRole("downloader") {
		return c.JSON(403, "You don't have permission to do this")
	}
	// Get the users to retrieve from query params
	UserIDStr := c.QueryParam("id") // comma separated list of users id, like: 1,2,3,4
	filesMap := make(map[string]*bytes.Buffer)
	for _, ID := range strings.Split(UserIDStr, ",") {
		UserIDInt, err := strconv.Atoi(ID)
		if err != nil {
			return c.JSON(400, "Invalid user ID:"+ID)
		}
		UserData := new(ORM.Usuario)
		UserData.Get(H.Db, uint(UserIDInt))
		if UserData.ID == 0 {
			// User not found, continue with the next one
			continue
		}
		// Create a new Excel file
		file := excelize.NewFile()
		// Create sheet Summary
		sheetName := "User Data"

		file.SetSheetName("Sheet1", sheetName)
		// Set sheet title
		file.SetCellValue(sheetName, "A1", "User Data")
		// Set User Summary Data headers in column A and Data in column B
		// Name, Last Name, Email, Cantity of Processes in charge, Tickets resolved, Tickets Pending, Average Duration per Ticket, Average Time spent per day, Cantity of Organizations
		file.SetCellValue(sheetName, "A2", "Name")
		file.SetCellValue(sheetName, "B2", UserData.Nombre)
		file.SetCellValue(sheetName, "A3", "Last Name")
		file.SetCellValue(sheetName, "B3", UserData.Apellido)
		file.SetCellValue(sheetName, "A4", "Email")
		file.SetCellValue(sheetName, "B4", UserData.Email)
		file.SetCellValue(sheetName, "A5", "Processes in charge")
		file.SetCellValue(sheetName, "B5", len(UserData.Procesos))
		file.SetCellValue(sheetName, "A6", "Tickets Closed")
		TicketsResolved := UserData.GetCantityTicketsClosed(H.Db)
		file.SetCellValue(sheetName, "B6", TicketsResolved)
		file.SetCellValue(sheetName, "A7", "Tickets Pending")
		file.SetCellValue(sheetName, "B7", UserData.GetCantityTicketsPending(H.Db))
		file.SetCellValue(sheetName, "A8", "Average Time per Ticket")
		file.SetCellValue(sheetName, "B8", UserData.AverageDurationPerTicket(H.Db))
		file.SetCellValue(sheetName, "A9", "Average Time spent per day")
		file.SetCellValue(sheetName, "B9", UserData.AverageTimeSpentPerDay(H.Db))
		file.SetCellValue(sheetName, "A10", "Organizations")
		file.SetCellValue(sheetName, "B10", len(UserData.Organizaciones))

		// Add Processes Sheet to Excel file
		sheetName = "Processes"
		file.NewSheet(sheetName)
		// Set sheet title
		file.SetCellValue(sheetName, "A1", "Processes")
		// Set Processes Data headers ID, Name, Alias, Folder ID, Folder Name, Warning Tolerance, Error Tolerance, Fatal Tolerance, Active Monitoring, Priority, Max Queue Time, Tickets, User Participation (Cantity of tickets where there's almost 1 detail contributed), Time Spent, ORG ID
		file.SetCellValue(sheetName, "A2", "ID")
		file.SetCellValue(sheetName, "B2", "Name")
		file.SetCellValue(sheetName, "C2", "Alias")
		file.SetCellValue(sheetName, "D2", "Folder ID")
		file.SetCellValue(sheetName, "E2", "Folder Name")
		file.SetCellValue(sheetName, "F2", "Warning Tolerance")
		file.SetCellValue(sheetName, "G2", "Error Tolerance")
		file.SetCellValue(sheetName, "H2", "Fatal Tolerance")
		file.SetCellValue(sheetName, "I2", "Active Monitoring")
		file.SetCellValue(sheetName, "J2", "Priority")
		file.SetCellValue(sheetName, "K2", "Max Queue Time")
		file.SetCellValue(sheetName, "L2", "Tickets")
		file.SetCellValue(sheetName, "M2", "User Participation (Per ticket)") // separate emails with semicolon
		file.SetCellValue(sheetName, "N2", "Time Spent")
		file.SetCellValue(sheetName, "O2", "ID Organization")
		ActualRow := 3
		for _, proceso := range UserData.Procesos {
			proceso.Get(H.Db, proceso.ID) // obtener toda la data del proceso
			file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), proceso.ID)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), proceso.Nombre)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), proceso.Alias)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), proceso.Folderid)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), proceso.Foldername)
			file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), proceso.WarningTolerance)
			file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), proceso.ErrorTolerance)
			file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), proceso.FatalTolerance)
			file.SetCellValue(sheetName, "I"+strconv.Itoa(ActualRow), proceso.ActiveMonitoring)
			file.SetCellValue(sheetName, "J"+strconv.Itoa(ActualRow), proceso.Priority)
			file.SetCellValue(sheetName, "K"+strconv.Itoa(ActualRow), proceso.MaxQueueTime)
			file.SetCellValue(sheetName, "L"+strconv.Itoa(ActualRow), len(proceso.TicketsProcesos))
			Participation := 0
			TimeSpent := new(time.Duration)
			for _, ticketsProceso := range proceso.TicketsProcesos {
				ticketsProceso.Get(H.Db, ticketsProceso.ID) // obtener toda la data del ticket
				ParticipationBool := false
				for _, ticketsDetalle := range ticketsProceso.Detalles {
					if uint(ticketsDetalle.UsuarioID) == UserData.ID {
						// Must be the user that created the detail to count as participation
						ParticipationBool = true
						duration := ticketsDetalle.FechaFin.Sub(ticketsDetalle.FechaInicio)
						*TimeSpent += duration
					}
				}
				if ParticipationBool { // If the user participated in the ticket
					Participation++ // Add 1 to the participation counter
				}
			}
			file.SetCellValue(sheetName, "M"+strconv.Itoa(ActualRow), Participation)
			file.SetCellValue(sheetName, "N"+strconv.Itoa(ActualRow), *TimeSpent)
			file.SetCellValue(sheetName, "O"+strconv.Itoa(ActualRow), proceso.OrganizacionID)
			ActualRow++
		}

		// Add Tickets Sheet to Excel file
		sheetName = "Tickets" // Tickets where the user are involved
		file.NewSheet(sheetName)
		// Set sheet title
		file.SetCellValue(sheetName, "A1", "Tickets")
		// add Headers: ID, ProcessID, Description, State, Priority, CreatorUserID,CreatedAt, StartDate, EndDate, Duration, Details_Count, type
		file.SetCellValue(sheetName, "A2", "ID")
		file.SetCellValue(sheetName, "B2", "Process ID")
		file.SetCellValue(sheetName, "C2", "Description")
		file.SetCellValue(sheetName, "D2", "State")
		file.SetCellValue(sheetName, "E2", "Priority")
		file.SetCellValue(sheetName, "F2", "Creator User ID")
		file.SetCellValue(sheetName, "G2", "Created At")
		file.SetCellValue(sheetName, "H2", "Start Date")
		file.SetCellValue(sheetName, "I2", "End Date")
		file.SetCellValue(sheetName, "J2", "Duration")
		file.SetCellValue(sheetName, "K2", "Details Count")
		file.SetCellValue(sheetName, "L2", "Type")
		ActualRow = 3
		var tickets []*ORM.TicketsProceso
		H.Db.Preload("Detalles", "usuario_id = ?", UserData.ID).Find(&tickets)
		for _, ticket := range tickets {
			if len(ticket.Detalles) == 0 {
				continue
			}
			ticket.Get(H.Db, ticket.ID) // obtener toda la data del ticket
			file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), ticket.ID)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), ticket.ProcesoID)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), ticket.Descripcion)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), ticket.Estado)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), ticket.Prioridad)
			file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), ticket.UsuarioCreadorID)
			file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), ticket.CreatedAt)
			var StartDate time.Time
			var EndDate time.Time
			var Duration time.Duration
			if len(ticket.Detalles) > 0 {
				StartDate = ticket.Detalles[0].FechaInicio
				if ticket.Estado == "Finalizado" {
					EndDate = ticket.Detalles[len(ticket.Detalles)-1].FechaFin
				} else {
					// Empty end date
					EndDate = time.Time{}
				}
				for _, detalle := range ticket.Detalles {
					Duration += detalle.FechaFin.Sub(detalle.FechaInicio)
				}
			}
			file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), StartDate)
			file.SetCellValue(sheetName, "I"+strconv.Itoa(ActualRow), EndDate)
			file.SetCellValue(sheetName, "J"+strconv.Itoa(ActualRow), Duration)
			file.SetCellValue(sheetName, "K"+strconv.Itoa(ActualRow), len(ticket.Detalles))
			file.SetCellValue(sheetName, "L"+strconv.Itoa(ActualRow), ticket.GetTipo(H.Db))
			ActualRow++
		}

		// Add Tickets_Details Sheet to Excel file
		sheetName = "Tickets_Details" // Tickets where the user are involved
		file.NewSheet(sheetName)
		// Set sheet title
		file.SetCellValue(sheetName, "A1", "Tickets Details")
		// Headers:
		// ID
		//TicketID
		//Detalle
		//FechaInicio
		//FechaFin
		//UsuarioID
		//Diagnostico
		//Duration
		file.SetCellValue(sheetName, "A2", "ID")
		file.SetCellValue(sheetName, "B2", "Ticket ID")
		file.SetCellValue(sheetName, "C2", "Detalle")
		file.SetCellValue(sheetName, "D2", "Fecha Inicio")
		file.SetCellValue(sheetName, "E2", "Fecha Fin")
		file.SetCellValue(sheetName, "F2", "Usuario ID")
		file.SetCellValue(sheetName, "G2", "Diagnostico")
		file.SetCellValue(sheetName, "H2", "Duration")
		ActualRow = 3
		var detalles []*ORM.TicketsDetalle
		H.Db.Where("usuario_id = ?", UserData.ID).Find(&detalles)
		for _, detalle := range detalles {
			file.SetCellValue(sheetName, "A"+strconv.Itoa(ActualRow), detalle.ID)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(ActualRow), detalle.TicketID)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(ActualRow), detalle.Detalle)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(ActualRow), detalle.FechaInicio)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(ActualRow), detalle.FechaFin)
			file.SetCellValue(sheetName, "F"+strconv.Itoa(ActualRow), detalle.UsuarioID)
			file.SetCellValue(sheetName, "G"+strconv.Itoa(ActualRow), detalle.Diagnostico)
			file.SetCellValue(sheetName, "H"+strconv.Itoa(ActualRow), detalle.FechaFin.Sub(detalle.FechaInicio))
			ActualRow++
		}
		// Write Excel file to memory buffer
		buf := new(bytes.Buffer)
		if err := file.Write(buf); err != nil {
			return c.JSON(500, "Failed to write user data to memory")
		}
		// ID_Name_LastName.xlsx
		fileName := strconv.Itoa(int(UserData.ID)) + "_" + UserData.Nombre + "_" + UserData.Apellido + ".xlsx"
		filesMap[fileName] = buf
	}
	if len(filesMap) == 0 {
		return c.JSON(404, "No users found")
	} else if len(filesMap) == 1 {
		// Just download a excel file
		// Set response headers

		fileName := ""
		buff := new(bytes.Buffer)
		for s, buffer := range filesMap {
			fileName = s
			buff = buffer
		}
		c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Response().Header().Set("Content-Disposition", "attachment; filename="+fileName)

		return c.Blob(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())
	} else {
		// Download a zip file with all the excel files
		// Set response headers
		c.Response().Header().Set("Content-Type", "application/zip")
		c.Response().Header().Set("Content-Disposition", "attachment; filename=users_data.zip")
		return c.Blob(200, "application/zip", ZipFiles(filesMap))
	}
}

func ZipFiles(filesMap map[string]*bytes.Buffer) []byte {
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)
	for fileName, fileBuffer := range filesMap {
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			return nil
		}
		_, err = fileWriter.Write(fileBuffer.Bytes())
		if err != nil {
			return nil
		}
	}
	zipWriter.Close()
	return zipBuffer.Bytes()

}
