package Server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/Mail"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
)

func (H *Handler) Login(c echo.Context) error {
	email := c.FormValue("email")       // Obtener el valor del campo "email" del formulario de inicio de sesión
	password := c.FormValue("password") // Obtener el valor del campo "password" del formulario de inicio de sesión
	if email == "" || password == "" {  // Validar si los campos son nulos o vacíos
		return c.JSON(http.StatusBadRequest, "Invalid email or password") // Devolver un error 400 de solicitud incorrecta con un mensaje de error
	}
	var user ORM.Usuario
	user.GetByEmail(H.DB, email) // Buscar al usuario por su correo electrónico en la base de datos
	if user.ID == 0 {            // Validar si el usuario no existe
		return c.JSON(http.StatusNotFound, "User not found") // Devolver un error 404 de no encontrado con un mensaje de error
	}
	if !user.CheckPassword(password) { // Validar si la contraseña es incorrecta
		return c.JSON(http.StatusUnauthorized, "Invalid email or password") // Devolver un error 401 de no autorizado con un mensaje de error
	}
	// Crear token
	token := jwt.New(jwt.SigningMethodHS512)
	// Establecer los datos del token
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID // Establecer el ID del usuario como un campo en los datos del token
	claims["user"] = user
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Establecer la fecha de vencimiento del token en 72 horas a partir de la hora actual
	// Generar el token codificado y enviarlo como respuesta
	t, err := token.SignedString([]byte(H.TokenKey))
	if err != nil {
		return err // Devolver cualquier error que ocurra al generar el token
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t, // Devolver el token codificado como un campo en la respuesta JSON
	})
}
func (H *Handler) ForgotPassword(c echo.Context) error {
	email := c.FormValue("email") // Obtener el valor del campo "email" del formulario de inicio de sesión
	if email == "" {              // Validar si los campos son nulos o vacíos
		return c.JSON(http.StatusBadRequest, "Invalid email") // Devolver un error 400 de solicitud incorrecta con un mensaje de error
	}
	var user ORM.Usuario
	H.DB.Where("email = ?", email).First(&user) // Buscar al usuario por su correo electrónico en la base de datos
	if user.ID == 0 {                           // Validar si el usuario no existe
		return c.JSON(http.StatusNotFound, "User not found") // Devolver un error 404 de no encontrado con un mensaje de error
	}
	// Generar una nueva contraseña aleatoria
	newPassword := functions.GeneratePassword(16)
	// Enviar un correo electrónico al usuario con la nueva contraseña
	Asunto := "Restablecimiento de contraseña ProcessMonitor"
	Cuerpo := Mail.GetBodyNewPassword(Mail.NewPassword{Nombre: user.Nombre + " " + user.Apellido, Password: newPassword})
	err := functions.SendMail([]string{email}, Asunto, Cuerpo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error sending email")
	}
	// Actualizar la contraseña del usuario en la base de datos
	user.SetPassword(newPassword)
	H.DB.Save(&user)
	return c.JSON(http.StatusOK, "Password reset successfully")
}
func (H *Handler) GetProfile(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Ocultar la contraseña del usuario
	User.GetComplete(H.DB)
	User.Password = ""
	for _, organization := range User.Organizaciones {
		organization.AppID = ""
		organization.AppSecret = ""
		organization.AppScope = ""
		organization.BaseURL = ""
	}
	return c.JSON(http.StatusOK, User)
}
func (H *Handler) UpdateProfile(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	User.GetComplete(H.DB) // Recargar el usuario desde la base de datos para obtener los ultimos cambios
	oldmail := User.Email
	newMail := c.FormValue("email")
	if newMail != "" {
		User.Email = newMail
	}
	newName := c.FormValue("name")
	if newName != "" {
		User.Nombre = newName
	}
	newLastName := c.FormValue("lastName")
	if newLastName != "" {
		User.Apellido = newLastName
	}
	newPassword := c.FormValue("password")
	if newPassword != "" {
		User.SetPassword(newPassword)
	}
	if oldmail != User.Email {
		// Verificar si el email ya existe en la base de datos
		checkUser := new(ORM.Usuario)
		H.DB.Where("email = ?", User.Email).First(&checkUser)
		if checkUser.ID != 0 {
			return c.JSON(http.StatusConflict, "Email already exists")
		}
	}
	// Guardar los datos actualizados del usuario en la base de datos
	H.DB.Updates(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}
func (H *Handler) GetUserOrganizations(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	User.GetComplete(H.DB) // Recargar el usuario desde la base de datos
	var Organizations []*ORM.Organizacion
	_ = H.DB.Model(&User).Association("Organizaciones").Find(&Organizations)
	for _, organization := range Organizations {
		organization.AppID = ""
		organization.AppSecret = ""
		organization.AppScope = ""
		organization.BaseURL = ""
	}
	return c.JSON(http.StatusOK, Organizations)
}

func (H *Handler) GetUserProcesses(c echo.Context) error {
	// Si el usuario posee el rol de administrador de procesos, devolver todos los procesos
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	User.GetComplete(H.DB) // Recargar el usuario desde la base de datos
	if User.HasRole("processes_administration") {
		Processes := new(ORM.Proceso).GetAll(H.DB)
		for _, process := range Processes {
			for _, user := range process.Usuarios {
				user.Password = ""
			}
			process.Organizacion.AppSecret = ""
			process.Organizacion.AppID = ""
			process.Organizacion.AppScope = ""
		}
		return c.JSON(http.StatusOK, Processes)
	}
	// Si el usuario no posee el rol de administrador de procesos, devolver los procesos a los que tiene acceso (precargando las organizaciones de los procesos)
	var Processes []*ORM.Proceso
	_ = H.DB.Model(&User).Preload("Organizacion").Preload("TicketsProcesos").Preload("Clientes").Preload("Usuarios").Preload("TicketsProcesos.Detalles").Preload("JobsHistory").Association("Procesos").Find(&Processes)
	for _, process := range Processes {
		for _, user := range process.Usuarios {
			user.Password = ""
		}
		process.Organizacion.AppSecret = ""
		process.Organizacion.AppID = ""
		process.Organizacion.AppScope = ""
	}
	return c.JSON(http.StatusOK, Processes)

}

func (H *Handler) GetUserIncidents(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	User.GetComplete(H.DB) // Recargar el usuario desde la base de datos
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	var procesosWithIncidents []*ORM.Proceso
	User.Procesos = ORM.GetListByUser(H.DB, User.ID)
	for _, proceso := range User.Procesos {
		for _, usuario := range proceso.Usuarios {
			usuario.Password = ""
		}
		proceso.Organizacion.AppSecret = ""
		proceso.Organizacion.AppID = ""
		proceso.Organizacion.AppScope = ""
		if proceso.TicketsProcesos != nil && len(proceso.TicketsProcesos) > 0 {
			procesosWithIncidents = append(procesosWithIncidents, proceso)
		}
	}
	var returnJSON = make(map[string][]*ORM.Proceso)

	for _, process := range procesosWithIncidents {
		ProcessWithOnGoingIncidents := *process
		ProcessWithOnGoingIncidents.TicketsProcesos = make([]*ORM.TicketsProceso, 0)
		ProcessWithoutIncidents := *process
		ProcessWithoutIncidents.TicketsProcesos = make([]*ORM.TicketsProceso, 0)
		for _, tickets := range process.TicketsProcesos {
			if tickets.Estado != "Finalizado" {
				ProcessWithOnGoingIncidents.TicketsProcesos = append(ProcessWithOnGoingIncidents.TicketsProcesos, tickets)
			} else {
				ProcessWithoutIncidents.TicketsProcesos = append(ProcessWithoutIncidents.TicketsProcesos, tickets)
			}
		}
		if len(ProcessWithOnGoingIncidents.TicketsProcesos) > 0 {
			returnJSON["ongoing"] = append(returnJSON["ongoing"], &ProcessWithOnGoingIncidents)
		}
		if len(ProcessWithoutIncidents.TicketsProcesos) > 0 {
			returnJSON["finished"] = append(returnJSON["finished"], &ProcessWithoutIncidents)
		}
	}
	// sort incidents inside process by process.TicketsProcesos.Detalles[0].FechaInicio
	for _, process := range returnJSON["ongoing"] {
		sort.Slice(process.TicketsProcesos, func(i, j int) bool {
			if len(process.TicketsProcesos[i].Detalles) == 0 || len(process.TicketsProcesos[j].Detalles) == 0 {
				return false
			}
			return process.TicketsProcesos[i].Detalles[0].FechaInicio.After(process.TicketsProcesos[j].Detalles[0].FechaInicio)
		})
	}
	// FirstPriority: process.TicketsProceso.Priority Highest first (No matter what, if the process had 1 ticket with priority 10, it will be first)
	// SecondPriority: process.Priority That means if there's some processes with same first priority, order them by process.Priority
	sort.Slice(returnJSON["ongoing"], func(i, j int) bool {
		MaxPriorityI := 0
		MaxPriorityJ := 0
		for _, ticket := range returnJSON["ongoing"][i].TicketsProcesos {
			if int(ticket.Prioridad) > MaxPriorityI {
				MaxPriorityI = int(ticket.Prioridad)
			}
		}
		for _, ticket := range returnJSON["ongoing"][j].TicketsProcesos {
			if int(ticket.Prioridad) > MaxPriorityJ {
				MaxPriorityJ = int(ticket.Prioridad)
			}
		}
		if MaxPriorityI == MaxPriorityJ {
			return returnJSON["ongoing"][i].Priority > returnJSON["ongoing"][j].Priority
		}
		return MaxPriorityI > MaxPriorityJ
	})
	// sort incidents inside process by process.TicketsProcesos.Prioridad Highest first
	for _, process := range returnJSON["ongoing"] {
		sort.Slice(process.TicketsProcesos, func(i, j int) bool {
			return process.TicketsProcesos[i].Prioridad > process.TicketsProcesos[j].Prioridad
		})
	}
	return c.JSON(http.StatusOK, returnJSON)
}

func (H *Handler) GetTicketSettings(c echo.Context) error {
	Settings := struct {
		NeedDiagnostic bool   `json:"needDiagnostic"`
		Type           string `json:"type"`
	}{
		NeedDiagnostic: true,
		Type:           "",
	}
	idTicket := c.QueryParam("id")
	ticketidInt, err := strconv.Atoi(idTicket)
	if err != nil {
		return c.JSON(400, "Invalid ticket ID")
	}
	ticket := new(ORM.TicketsProceso)
	ticket.Get(H.DB, uint(ticketidInt))
	if ticket.ID == 0 {
		return c.JSON(404, "ticket not found")
	}
	if ticket.Tipo.NecesitaDiagnostico {
		for _, detalle := range ticket.Detalles {
			if detalle.Diagnostico { // si algun detalle tiene diagnostico, no se necesita otro
				Settings.NeedDiagnostic = false
			}
		}
	} else {
		Settings.NeedDiagnostic = false
	}
	Settings.Type = ticket.Tipo.Nombre
	return c.JSON(200, Settings)
}

func (H *Handler) PostIncidentDetails(c echo.Context) error {
	// Form data:
	// - incidentID: ID del incidente
	// - details: Detalles del incidente
	// - fechaInicio: Fecha de inicio del detalle
	// - fechaFin: Fecha de fin del detalle
	// - estado: Nuevo estado del incidente
	// - IsDiagnostic: Indica si el detalle es un diagnostico
	// Obtener el ID del usuario del token JWT
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener el incidente de la base de datos
	Incident := new(ORM.TicketsProceso)
	IncidentID, _ := strconv.Atoi(c.FormValue("incidentID"))
	Incident.Get(H.DB, uint(IncidentID))
	OldState := Incident.Estado
	if Incident.ID == 0 {
		return c.JSON(http.StatusNotFound, "Incident not found")
	}
	// Verifica que el incidente no esté cerrado
	if Incident.Estado == "Finalizado" {
		return c.JSON(http.StatusForbidden, "Incident is already closed")
	}
	// Obtener el proceso del incidente
	Process := new(ORM.Proceso)
	Process.Get(H.DB, Incident.ProcesoID)
	if Process.ID == 0 {
		return c.JSON(http.StatusNotFound, "Process not found")
	}
	// Revisar que el usuario tenga acceso al proceso
	var UserHasAccess bool
	for _, user := range Process.Usuarios {
		if user.ID == User.ID {
			UserHasAccess = true
		}
	}
	if !UserHasAccess && !User.HasRole("processes_administration") {
		return c.JSON(http.StatusForbidden, "User does not have access to process")
	}
	// Obtener el estado del incidente
	IncidentState := c.FormValue("estado")
	if IncidentState != "Finalizado" && IncidentState != "En Progreso" {
		return c.JSON(http.StatusBadRequest, "Invalid incident state")
	}
	// Obtener la fecha de inicio del detalle (DateTime from Dart)
	IncidentDetailStartDate, err := time.Parse("2006-01-02 15:04:05", c.FormValue("fechaInicio"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid incident detail start date")
	}
	// Obtener la fecha de fin del detalle (DateTime from Dart)
	IncidentDetailEndDate, err := time.Parse("2006-01-02 15:04:05", c.FormValue("fechaFin"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid incident detail end date")
	}
	// Verificar que la fecha de fin sea mayor a la fecha de inicio
	if IncidentDetailEndDate.Before(IncidentDetailStartDate) {
		return c.JSON(http.StatusBadRequest, "End date must be greater than start date")
	}
	// Se omite la validación horaria entre la fecha de ultimo detalle, debido a que existe la posibilidad de que el usuario este en otra zona horaria
	// y esto es disruptivo al trabajar con dart y tener un servidor en una ubicación distinta

	// Crear el detalle del incidente

	IncidentDetail := &ORM.TicketsDetalle{
		TicketID:    int(Incident.ID),
		Detalle:     c.FormValue("details"),
		FechaInicio: IncidentDetailStartDate,
		FechaFin:    IncidentDetailEndDate,
		UsuarioID:   int(User.ID),
		Diagnostico: strings.ToLower(c.FormValue("IsDiagnostic")) == "true",
	}
	Incident.Estado = IncidentState
	Incident.Detalles = append(Incident.Detalles, IncidentDetail)
	H.DB.Save(Incident)
	if Incident.Estado != OldState {
		// Enviar notificación a los usuarios del proceso
		go func() {
			// Se debe de redactar un correo con el ID del ticket, el nombre del proceso, el nuevo estado
			body := Mail.GetBodyTicketChange(Mail.IncidentChange{
				ID:            int(Incident.ID),
				NombreProceso: Process.Nombre,
				NuevoEstado:   Incident.Estado,
			})
			subject := fmt.Sprintf("Cambió de estado en incidente %d de proceso %s", Incident.ID, Process.Nombre)
			_ = functions.SendMail(Process.GetEmails(), subject, body)
		}()
	}
	return c.JSON(http.StatusOK, Incident)
}

// NewIncident
func (H *Handler) NewIncident(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	// check if the user had "processes_administration" role Or if the user is the owner of the process
	User, err := H.GetUserJWT(c)
	if err != nil {
		return err
	}
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole && !User.HasRole("monitor") {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.DB, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the incident data from the request
	Incident := new(ORM.TicketsProceso)
	err = c.Bind(Incident)
	if err != nil {
		return c.JSON(400, "Invalid incident data")
	}
	// Check if the process has incidents of the same type ongoing (not Estado 3)
	for _, incident := range Process.TicketsProcesos {
		if incident.TipoID == Incident.TipoID && incident.Estado != "Finalizado" {
			return c.JSON(400, "Ya existe un incidente de este tipo en el proceso")
		}
	}

	Incident.ProcesoID = Process.ID
	Incident.Proceso = &Process
	Incident.UsuarioCreadorID = int(User.ID)
	if len(Incident.Detalles) != 0 {
		for _, detail := range Incident.Detalles {
			detail.UsuarioID = int(User.ID)
			detail.FechaInicio = time.Now().UTC()
			detail.FechaFin = time.Now().UTC()
		}
	}

	if Incident.Prioridad == 0 { // if the priority is not set, set it to the process priority
		Incident.Prioridad = uint(Process.Priority)
	}

	// Create the incident to retrieve the ID
	H.DB.Create(Incident)

	if Incident.TipoID == 1 {
		Process.ActiveMonitoring = false
		H.DB.Save(&Process)
	}
	// Check if there's a Detail in the incident
	if len(Incident.Detalles) == 0 {
		DefaultDetail := new(ORM.TicketsDetalle)
		DefaultDetail.FechaInicio = time.Now().UTC()
		DefaultDetail.FechaFin = time.Now().UTC()
		DefaultDetail.Detalle = "Evento creado por el usuario " + User.Nombre + " " + User.Apellido
		DefaultDetail.TicketID = int(Incident.ID)
		DefaultDetail.UsuarioID = int(User.ID)
		H.DB.Create(DefaultDetail)
	}

	// Send the notification to the users
	Emails := Process.GetEmails()
	// Send the email to the users
	body := Mail.GetBodyNewTicket(Mail.NewIncident{
		ID:            int(Incident.ID),
		NombreProceso: Process.Nombre,
		Tipo:          Incident.GetTipo(H.DB),
		Descripcion:   Incident.Descripcion,
	})
	subject := "Nuevo Ticket en el proceso " + Process.Nombre
	err = functions.SendMail(Emails, subject, body)
	if err != nil {
		return c.JSON(500, "Error sending email")
	}

	return c.JSON(200, Incident)
}
