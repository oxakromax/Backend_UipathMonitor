package Server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
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
	user.GetByEmail(H.Db, email) // Buscar al usuario por su correo electrónico en la base de datos
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
	claims["id"] = user.ID                                // Establecer el ID del usuario como un campo en los datos del token
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
	H.Db.Where("email = ?", email).First(&user) // Buscar al usuario por su correo electrónico en la base de datos
	if user.ID == 0 {                           // Validar si el usuario no existe
		return c.JSON(http.StatusNotFound, "User not found") // Devolver un error 404 de no encontrado con un mensaje de error
	}
	// Generar una nueva contraseña aleatoria
	newPassword := functions.GeneratePassword(16)
	// Enviar un correo electrónico al usuario con la nueva contraseña
	Asunto := "Restablecimiento de contraseña ProcessMonitor"
	Cuerpo := "Su nueva contraseña es: " + newPassword
	err := functions.SendMail([]string{email}, Asunto, Cuerpo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error sending email")
	}
	// Actualizar la contraseña del usuario en la base de datos
	user.SetPassword(newPassword)
	H.Db.Save(&user)
	return c.JSON(http.StatusOK, "Password reset successfully")
}
func (H *Handler) GetProfile(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(id))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Ocultar la contraseña del usuario
	User.GetComplete(H.Db)
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
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	H.Db.First(&User, id)
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
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
		H.Db.Where("email = ?", User.Email).First(&checkUser)
		if checkUser.ID != 0 {
			return c.JSON(http.StatusConflict, "Email already exists")
		}
	}
	// Guardar los datos actualizados del usuario en la base de datos
	H.Db.Updates(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}
func (H *Handler) GetUserOrganizations(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(id))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener las organizaciones del usuario
	var Organizations []*ORM.Organizacion
	_ = H.Db.Model(&User).Association("Organizaciones").Find(&Organizations)
	if Organizations == nil || len(Organizations) == 0 {
		return c.JSON(http.StatusNotFound, "Organizations not found")
	}
	for _, organization := range Organizations {
		organization.AppID = ""
		organization.AppSecret = ""
		organization.AppScope = ""
		organization.BaseURL = ""
	}
	return c.JSON(http.StatusOK, Organizations)
}

func (H *Handler) GetUserProcesses(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(id))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener las organizaciones del usuario
	var Organizations []*ORM.Organizacion
	_ = H.Db.Model(&User).Association("Organizaciones").Find(&Organizations)
	if Organizations == nil || len(Organizations) == 0 {
		return c.JSON(http.StatusNotFound, "Organizations not found")
	}
	// Obtener los procesos de cada organización
	var Processes []*ORM.Proceso
	for _, organization := range Organizations {
		_ = H.Db.Model(&organization).Association("Procesos").Find(&Processes)
	}
	if Processes == nil || len(Processes) == 0 {
		return c.JSON(http.StatusNotFound, "Processes not found")
	}
	return c.JSON(http.StatusOK, Processes)
}

func (H *Handler) GetUserIncidents(c echo.Context) error {
	// Obtener el ID del usuario del token JWT
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(id))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	var procesosWithIncidents []*ORM.Proceso
	for _, proceso := range User.Procesos {
		proceso.Get(H.Db, proceso.ID)
		for _, usuario := range proceso.Usuarios {
			usuario.Password = ""
		}
		proceso.Organizacion.AppSecret = ""
		proceso.Organizacion.AppID = ""
		proceso.Organizacion.AppScope = ""
		if proceso.IncidentesProceso != nil && len(proceso.IncidentesProceso) > 0 {
			procesosWithIncidents = append(procesosWithIncidents, proceso)
		}
	}
	var returnJson = make(map[string][]*ORM.Proceso)

	for _, process := range procesosWithIncidents {
		ProcessWithOnGoingIncidents := *process
		ProcessWithOnGoingIncidents.IncidentesProceso = make([]*ORM.IncidenteProceso, 0)
		ProcessWithoutIncidents := *process
		ProcessWithoutIncidents.IncidentesProceso = make([]*ORM.IncidenteProceso, 0)
		for _, incidentes := range process.IncidentesProceso {
			if incidentes.Estado != 3 {
				ProcessWithOnGoingIncidents.IncidentesProceso = append(ProcessWithOnGoingIncidents.IncidentesProceso, incidentes)
			} else {
				ProcessWithoutIncidents.IncidentesProceso = append(ProcessWithoutIncidents.IncidentesProceso, incidentes)
			}
		}
		returnJson["ongoing"] = append(returnJson["ongoing"], &ProcessWithOnGoingIncidents)
		returnJson["finished"] = append(returnJson["finished"], &ProcessWithoutIncidents)
	}
	// sort incidents inside process by incidentes.Detalles[0].FechaInicio
	for _, process := range returnJson["ongoing"] {
		sort.Slice(process.IncidentesProceso, func(i, j int) bool {
			if len(process.IncidentesProceso[i].Detalles) == 0 || len(process.IncidentesProceso[j].Detalles) == 0 {
				return false
			}
			return process.IncidentesProceso[i].Detalles[0].FechaInicio.After(process.IncidentesProceso[j].Detalles[0].FechaInicio)
		})
	}
	return c.JSON(http.StatusOK, returnJson)
}

func (H *Handler) PostIncidentDetails(c echo.Context) error {
	// Form data:
	// - incidentID: ID del incidente
	// - details: Detalles del incidente
	// - fechaInicio: Fecha de inicio del detalle
	// - fechaFin: Fecha de fin del detalle
	// - estado: Nuevo estado del incidente
	// Obtener el ID del usuario del token JWT
	id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(id))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener el incidente de la base de datos
	Incident := new(ORM.IncidenteProceso)
	IncidentID, _ := strconv.Atoi(c.FormValue("incidentID"))
	Incident.Get(H.Db, uint(IncidentID))
	OldState := Incident.Estado
	if Incident.ID == 0 {
		return c.JSON(http.StatusNotFound, "Incident not found")
	}
	// Verifica que el incidente no esté cerrado
	if Incident.Estado == 3 {
		return c.JSON(http.StatusForbidden, "Incident is already closed")
	}
	// Obtener el proceso del incidente
	Process := new(ORM.Proceso)
	Process.Get(H.Db, Incident.ProcesoID)
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
	// Obtener el estado del incidente, debe ser 2 o 3
	IncidentState, err := strconv.Atoi(c.FormValue("estado"))
	if err != nil || IncidentState != 3 && IncidentState != 2 {
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
	// Verificar que la fecha de inicio sea mayor a la fecha del último detalle
	if len(Incident.Detalles) > 0 {
		if IncidentDetailStartDate.Before(Incident.Detalles[len(Incident.Detalles)-1].FechaInicio) {
			return c.JSON(http.StatusBadRequest, "Start date must be greater than last detail start date")
		}
	}
	// Crear el detalle del incidente
	IncidentDetail := &ORM.IncidentesDetalle{
		IncidenteID: int(Incident.ID),
		Detalle:     c.FormValue("details"),
		FechaInicio: IncidentDetailStartDate,
		FechaFin:    IncidentDetailEndDate,
		UsuarioID:   int(User.ID),
	}
	Incident.Estado = IncidentState
	Incident.Detalles = append(Incident.Detalles, IncidentDetail)
	H.Db.Save(Incident)
	if Incident.Estado != OldState {
		// Enviar notificación a los usuarios del proceso
		go func() {
			// Se debe de redactar un correo con el ID del ticket, el nombre del proceso, el nuevo estado
			body := fmt.Sprintf("El incidente %d del proceso %s ha cambiado de estado a %s.\n\n Esta es una notificación automática. Por favor, no responda.", Incident.ID, Process.Nombre, Incident.GetTipo())
			subject := fmt.Sprintf("Cambió el estado del incidente %d del proceso %s", Incident.ID, Process.Nombre)
			functions.SendMail(Process.GetEmails(), subject, body)
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
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole && !User.HasRole("monitor") {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the incident data from the request
	Incident := new(ORM.IncidenteProceso)
	err = c.Bind(Incident)
	if err != nil {
		return c.JSON(400, "Invalid incident data")
	}
	// Check if the process has incidents of the same type ongoing (not Estado 3)
	for _, incident := range Process.IncidentesProceso {
		if incident.Tipo == Incident.Tipo && incident.Estado != 3 {
			return c.JSON(400, "Ya existe un incidente de este tipo en el proceso")
		}
	}

	Incident.ProcesoID = Process.ID
	Incident.Proceso = &Process
	// Create the incident to retrieve the ID
	H.Db.Create(Incident)
	// Check if there's a Detail in the incident
	if len(Incident.Detalles) == 0 {
		DefaultDetail := new(ORM.IncidentesDetalle)
		DefaultDetail.FechaInicio = time.Now()
		DefaultDetail.FechaFin = time.Now()
		DefaultDetail.Detalle = "Evento creado por el usuario " + User.Nombre + " " + User.Apellido
		DefaultDetail.IncidenteID = int(Incident.ID)
		DefaultDetail.UsuarioID = int(User.ID)
		H.Db.Create(DefaultDetail)
	} else {
		for i := 0; i < len(Incident.Detalles); i++ {
			Incident.Detalles[i].IncidenteID = int(Incident.ID)
			Incident.Detalles[i].UsuarioID = int(User.ID)
			// Fechas
			Incident.Detalles[i].FechaInicio = time.Now()
			Incident.Detalles[i].FechaFin = time.Now()
			H.Db.Create(Incident.Detalles[i])
		}
	}

	if Incident.Tipo == 1 {
		Process.ActiveMonitoring = false
		H.Db.Save(Process)
	}
	go func() {
		// Send the notification to the users
		Emails := Process.GetEmails()
		// Send the email to the users
		body := "Se ha creado un nuevo Ticket de tipo " + Incident.GetTipo() + " en el proceso " + Process.Nombre + " con ID " + strconv.Itoa(int(Incident.ID)) + ".\n\n" +
			"Para ver el ticket, ingrese a la plataforma de monitoreo de procesos." +
			"\n\n" +
			"Este es un mensaje automático, por favor no responda a este correo."
		subject := "Nuevo Ticket en el proceso " + Process.Nombre
		err := functions.SendMail(Emails, subject, body)
		if err != nil {
			fmt.Println("Error sending email: ", err)
		}

	}()
	return c.JSON(200, Incident)
}
