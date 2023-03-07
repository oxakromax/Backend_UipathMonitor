package main

import (
	"GormTest/ORM"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/rand"
	"net/http"
	"time"
)

type Handler struct {
	db  *gorm.DB
	Key string
}

func generatePassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+{}[];':\",./<>?"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
func OpenDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=Proyecto port=5432 sslmode=disable"
	log := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log,
	})
	log.Info(nil, "Database connection successfully opened")
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&ORM.Organizacion{}, &ORM.Cliente{}, &ORM.Proceso{}, &ORM.IncidenteProceso{}, &ORM.IncidentesDetalle{}, &ORM.Usuario{}, &ORM.Rol{})
	if err != nil {
		panic("failed to migrate database")
	}
	return db
}

func (r *Handler) Login(c echo.Context) error {
	email := c.FormValue("email")       // Obtener el valor del campo "email" del formulario de inicio de sesión
	password := c.FormValue("password") // Obtener el valor del campo "password" del formulario de inicio de sesión
	if email == "" || password == "" {  // Validar si los campos son nulos o vacíos
		return c.JSON(http.StatusBadRequest, "Invalid email or password") // Devolver un error 400 de solicitud incorrecta con un mensaje de error
	}
	var user ORM.Usuario
	r.db.Where("email = ?", email).First(&user) // Buscar al usuario por su correo electrónico en la base de datos
	if user.ID == 0 {                           // Validar si el usuario no existe
		return c.JSON(http.StatusNotFound, "User not found") // Devolver un error 404 de no encontrado con un mensaje de error
	}
	if !user.CheckPassword(password) { // Validar si la contraseña es incorrecta
		return c.JSON(http.StatusUnauthorized, "Invalid email or password") // Devolver un error 401 de no autorizado con un mensaje de error
	}
	echojwt.JWT(r.Key) // Configurar la clave JWT globalmente
	// Crear token
	token := jwt.New(jwt.SigningMethodHS256)
	// Establecer los datos del token
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID                                // Establecer el ID del usuario como un campo en los datos del token
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Establecer la fecha de vencimiento del token en 72 horas a partir de la hora actual
	// Generar el token codificado y enviarlo como respuesta
	t, err := token.SignedString([]byte(r.Key))
	if err != nil {
		return err // Devolver cualquier error que ocurra al generar el token
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t, // Devolver el token codificado como un campo en la respuesta JSON
	})
}

func (r *Handler) checkRoleMiddleware(allowedRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Verificar si el usuario está autenticado y tiene un rol permitido
			id := int(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64)) // Extraer el ID del usuario del token JWT
			User := ORM.Usuario{}
			User.Get(r.db, uint(id))              // Obtener el usuario de la base de datos
			for _, UserRole := range User.Roles { // Iterar sobre los roles del usuario
				for _, role := range allowedRoles { // Iterar sobre los roles permitidos
					if UserRole.Nombre == role { // Si se encuentra un rol permitido
						return next(c) // Permitir el acceso al siguiente middleware o controlador
					}
					if UserRole.Nombre == "admin" { // Si el usuario tiene el rol de administrador
						return next(c) // Permitir el acceso al siguiente middleware o controlador
					}
				}
			}
			return echo.ErrUnauthorized // Acceso denegado si el usuario no tiene el rol permitido
		}
	}
}

func main() {
	e := echo.New()
	R := &Handler{db: OpenDB(), Key: generatePassword(32)}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(R.Key),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for signup and login requests
			if c.Path() == "/auth" || c.Path() == "/signup" {
				return true
			}
			return false
		},
	}))
	e.POST("/auth", R.Login)
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}, R.checkRoleMiddleware([]string{"admin"}))

	e.Start(":8080")
}
