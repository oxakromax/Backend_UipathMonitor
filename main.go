package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/localtunnel/go-localtunnel"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func generatePassword(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+{}[];:,./<>?"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letters[r.Intn(len(letters))]
	}
	return string(result)
}
func OpenDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=Proyecto port=5432 sslmode=disable"
	log := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log,
	})
	db.Logger.Info(nil, "Database connection successfully opened")
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&ORM.Organizacion{}, &ORM.Cliente{}, &ORM.Proceso{}, &ORM.IncidenteProceso{}, &ORM.IncidentesDetalle{}, &ORM.Usuario{}, &ORM.Rol{},
		&ORM.Route{})
	if err != nil {
		panic("failed to migrate database")
	}
	return db
}

type Handler struct {
	Db              *gorm.DB
	TokenKey        string
	UniversalRoutes []string
	DbKey           string
}

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
	newPassword := generatePassword(16)
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
func (H *Handler) checkRoleMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, route := range H.UniversalRoutes {
				if strings.ToLower(route) == strings.ToLower(c.Path()) {
					return next(c) // Permitir el acceso a la ruta
				}
			}
			// Verificar si el usuario está autenticado y tiene un rol permitido
			id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64)) // Extraer el ID del usuario del token JWT
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
func (H *Handler) RefreshDataBase(e *echo.Echo) {
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
		Nombre: "admin",
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
	AdminUser.Get(H.Db, 1)
	// Reemplazar los roles del usuario administrador con el rol de administrador
	AdminUser.SetPassword("test")
	H.Db.Save(&AdminUser)
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
}
func (H *Handler) GetUsers(c echo.Context) error {
	// if query id is not empty, return the user with that id
	id := c.QueryParam("id")
	if id != "" {
		// Convertir el ID de la consulta en un número entero
		ID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid ID")
		}
		// Obtener el usuario de la base de datos
		User := new(ORM.Usuario)
		User.Get(H.Db, uint(ID))
		if User.ID == 0 {
			return c.JSON(http.StatusNotFound, "User not found")
		}
		// Ocultar la contraseña del usuario
		User.Password = ""
		return c.JSON(http.StatusOK, User)
	}

	// Obtener todos los usuarios de la base de datos
	Users := new(ORM.Usuario).GetAll(H.Db)
	// Ocultar la contraseña de los usuarios
	for i := range Users {
		Users[i].Password = ""
	}
	return c.JSON(http.StatusOK, Users)
}
func (H *Handler) DeleteUser(c echo.Context) error {
	id := c.QueryParam("id")
	// Convertir el ID de la consulta en un número entero
	ID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(ID))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Eliminar el usuario de la base de datos
	H.Db.Delete(&User)
	return c.JSON(http.StatusOK, "User deleted")
}
func (H *Handler) CreateUser(c echo.Context) error {
	// Obtener el usuario de la solicitud
	User := new(ORM.Usuario)
	if err := c.Bind(User); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Verificar si el usuario ya existe en la base de datos
	checkUser := new(ORM.Usuario)
	H.Db.Where("email = ?", User.Email).First(&checkUser)
	if checkUser.ID != 0 {
		return c.JSON(http.StatusConflict, "User already exists")
	}
	// Encriptar la contraseña del usuario
	User.Password = generatePassword(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(User.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while encrypting password")
	}
	err = functions.SendMail([]string{User.Email}, "Welcome to Uipath Monitor", "Your password is: "+User.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while sending email")
	}
	// Guardar el usuario en la base de datos
	User.Password = string(hash)
	H.Db.Create(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}
func (H *Handler) UpdateUser(c echo.Context) error {
	// Obtener ID desde query
	id := c.QueryParam("id")
	// Convertir el ID de la consulta en un número entero
	ID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(H.Db, uint(ID))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener el usuario de la solicitud
	if err := c.Bind(User); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Sobre escribir roles del usuario
	err = H.Db.Model(&User).Association("Roles").Replace(User.Roles)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while updating user")
	}
	// Guardar el usuario en la base de datos
	H.Db.Save(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
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
	User.Password = ""
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
	// Actualizar los datos del usuario
	if err := c.Bind(&User); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid data")
	}
	if oldmail != User.Email {
		// Verificar si el email ya existe en la base de datos
		checkUser := new(ORM.Usuario)
		H.Db.Where("email = ?", User.Email).First(&checkUser)
		if checkUser.ID != 0 {
			return c.JSON(http.StatusConflict, "Email already exists")
		}
	}
	// verifica si envio una nueva contraseña
	if User.Password != "" {
		// Encriptar la contraseña del usuario
		hash, err := bcrypt.GenerateFromPassword([]byte(User.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Error while encrypting password")
		}
		User.Password = string(hash)
	}
	// Guardar los datos actualizados del usuario en la base de datos
	H.Db.Updates(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}
func (H *Handler) CreateOrganization(c echo.Context) error {
	// Obtener la organización de la solicitud
	Organization := new(ORM.Organizacion)
	if err := c.Bind(Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Verificar si la organización ya existe en la base de datos
	checkOrganization := new(ORM.Organizacion)
	H.Db.Where("uipathname = ? and tenantname = ?", Organization.Uipathname, Organization.Tenantname).First(&checkOrganization)
	if checkOrganization.ID != 0 {
		return c.JSON(http.StatusConflict, "Organization already exists")
	}

	// Cifrar datos sensibles app_id y app_secret
	Organization.AppID, _ = functions.EncryptAES(H.DbKey, Organization.AppID)
	Organization.AppSecret, _ = functions.EncryptAES(H.DbKey, Organization.AppSecret)
	// Verificar que los datos son correctos
	err := Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath credentials")
	}
	// Guardar la organización en la base de datos
	H.Db.Create(&Organization)
	// Agregar a cada Administrador de la organización
	Admins := new(ORM.Usuario).GetAdmins(H.Db)
	for _, admin := range Admins {
		_ = H.Db.Model(&Organization).Association("Usuarios").Append(admin)
	}
	JsonFolders := new(UipathAPI.FoldersResponse)
	err = Organization.GetFromApi(JsonFolders)
	if err != nil {
		for _, Folder := range JsonFolders.Value {
			IDFolder := Folder.Id
			JsonProcesses := new(UipathAPI.ReleasesResponse)
			err = Organization.GetFromApi(JsonProcesses, IDFolder)
			if err != nil {
				for _, Process := range JsonProcesses.Value {
					// Obtener el proceso de la base de datos
					ProcessDB := ORM.Proceso{
						Nombre:           Process.Name,
						Alias:            "",
						Folderid:         uint(IDFolder),
						Foldername:       Folder.DisplayName,
						OrganizacionID:   Organization.ID,
						WarningTolerance: 999, // 999 = no limit
						ErrorTolerance:   999, // 999 = no limit
						FatalTolerance:   999, // 999 = no limit
					}
					// Guardar el proceso en la base de datos
					H.Db.Create(&ProcessDB)
				}
			}

		}
	}
	return c.JSON(http.StatusOK, Organization)
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
	return c.JSON(http.StatusOK, Organizations)
}
func (H *Handler) GetOrganizations(c echo.Context) error {
	Organization := new(ORM.Organizacion)
	if c.QueryParam("id") != "" {
		// Obtener ID de la organización de la solicitud
		organizationID, err := strconv.Atoi(c.QueryParam("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid organization ID")
		}
		// Obtener la organización de la base de datos
		Organization.Get(H.Db, uint(organizationID))
		if Organization.ID == 0 {
			return c.JSON(http.StatusNotFound, "Organization not found")
		}
		for _, usuario := range Organization.Usuarios {
			usuario.Password = ""
		}
		return c.JSON(http.StatusOK, Organization)
	}
	// Obtener las organizaciones de la base de datos
	AllOrgs := Organization.GetAll(H.Db)
	if AllOrgs == nil || len(AllOrgs) == 0 {
		return c.JSON(http.StatusNotFound, "Organizations not found")
	}
	for _, org := range AllOrgs {
		for _, usuario := range org.Usuarios {
			usuario.Password = ""
		}
	}
	return c.JSON(http.StatusOK, AllOrgs)

}
func (H *Handler) UpdateOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Actualizar los datos de la organización
	if err := c.Bind(&Organization); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid data")
	}
	_, errDecryption1 := functions.DecryptAES(H.DbKey, Organization.AppID) // Verificar si los datos ya están cifrados
	_, errDecryption2 := functions.DecryptAES(H.DbKey, Organization.AppSecret)
	if errDecryption1 != nil || errDecryption2 != nil { // Si no estaban cifrados, significa que se actualizaron
		Organization.AppID, _ = functions.EncryptAES(H.DbKey, Organization.AppID) // Se encriptan primero
		Organization.AppSecret, _ = functions.EncryptAES(H.DbKey, Organization.AppSecret)
	}
	err = Organization.CheckAccessAPI()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Please check UiPath Data")
	}

	// Guardar los datos actualizados de la organización en la base de datos
	H.Db.Updates(&Organization)
	return c.JSON(http.StatusOK, Organization)
}
func (H *Handler) DeleteOrganization(c echo.Context) error {
	// Obtener ID de la organización de la solicitud
	organizationID, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid organization ID")
	}
	// Obtener la organización de la base de datos
	Organization := new(ORM.Organizacion)
	Organization.Get(H.Db, uint(organizationID))
	if Organization.ID == 0 {
		return c.JSON(http.StatusNotFound, "Organization not found")
	}
	// Eliminar la organización de la base de datos
	H.Db.Delete(&Organization)
	// Eliminar Procesos de la organización
	for _, proceso := range Organization.Procesos {
		H.Db.Delete(&proceso)
	}
	// Eliminar Clientes de la organización
	for _, cliente := range Organization.Clientes {
		H.Db.Delete(&cliente)
	}
	return c.JSON(http.StatusOK, "Organization deleted successfully")
}

func main() {
	e := echo.New()
	H := &Handler{
		Db:              OpenDB(),
		TokenKey:        generatePassword(32),
		UniversalRoutes: []string{"/auth", "/forgot"},
		DbKey:           os.Getenv("DB_KEY"),
	}
	if H.DbKey == "" {
		fmt.Println("DB_KEY environment variable not set")
		NewKey, _ := functions.GenerateAESKey()
		fmt.Println("Generated key: " + NewKey)
		return
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(H.TokenKey),
		SigningMethod: "HS512",
		Skipper: func(c echo.Context) bool {
			// Skip authentication for signup and login requests
			for _, route := range H.UniversalRoutes {
				if strings.ToLower(route) == strings.ToLower(c.Path()) {
					return true
				}
			}
			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, "Invalid or expired JWT")
		},
	}))
	e.Use(H.checkRoleMiddleware())
	e.POST("/auth", H.Login)
	e.POST("/forgot", H.ForgotPassword)
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/users", H.GetUsers)
	e.DELETE("/users", H.DeleteUser)
	e.POST("/users", H.CreateUser)
	e.PUT("/users", H.UpdateUser)
	e.GET("/profile", H.GetProfile)
	e.PUT("/profile", H.UpdateProfile)
	e.POST("/organization", H.CreateOrganization)
	e.PUT("/organization", H.UpdateOrganization)
	e.DELETE("/organization", H.DeleteOrganization)
	e.GET("/organization", H.GetOrganizations)
	e.GET("/user-orgs", H.GetUserOrganizations)
	H.RefreshDataBase(e)
	listener, err := localtunnel.Listen(localtunnel.Options{
		Subdomain: "golanguipathmonitortunnel",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Tunnel URL: " + listener.URL())
	e.Listener = listener
	err = e.Start(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
