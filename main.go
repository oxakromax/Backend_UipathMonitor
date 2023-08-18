package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/Server"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenDB() *gorm.DB {
	Host := os.Getenv("DB_HOST")
	User := os.Getenv("DB_USER")
	Password := os.Getenv("DB_PASSWORD")
	Database := os.Getenv("DB_NAME")
	Port := os.Getenv("DB_PORT")
	SSLMode := os.Getenv("DB_SSLMODE")
	if Host == "" {
		Host = "localhost"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", Host, User, Password, Database, Port, SSLMode)
	log := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	db.Logger.Info(nil, "Database connection successfully opened")
	if err != nil {
		panic("failed to connect database")
	}
	createEnumTypeIfNotExists(db)
	err = db.AutoMigrate(&ORM.Organizacion{}, &ORM.Cliente{}, &ORM.Proceso{}, &ORM.TicketsTipo{}, &ORM.TicketsProceso{}, &ORM.TicketsDetalle{}, &ORM.Usuario{}, &ORM.Rol{},
		&ORM.Route{}, &ORM.JobHistory{})
	if err != nil {
		fmt.Println(err)
		panic("failed to migrate database")
	}
	return db
}

func createEnumTypeIfNotExists(db *gorm.DB) {
	var exists int
	db.Raw("SELECT 1 FROM pg_type WHERE typname = 'estado_enum'").Scan(&exists)
	if exists == 0 {
		db.Exec("CREATE TYPE estado_enum AS ENUM ('Iniciado', 'En Progreso', 'Finalizado')")
	}
}

func main() {
	err := godotenv.Load()
	e := echo.New()
	H := &Server.Handler{
		Db:                  OpenDB(),
		TokenKey:            os.Getenv("TOKEN_KEY"),
		UniversalRoutes:     []string{"/auth", "/forgot", "/client/tickets"},
		UserUniversalRoutes: []string{"/user/profile", "/pingAuth"},
		DbKey:               os.Getenv("DB_KEY"),
	}
	if H.DbKey == "" {
		fmt.Println("DB_KEY environment variable not set")
		NewKey, _ := functions.GenerateAESKey()
		fmt.Println("Generated key: " + NewKey)
		return
	}
	e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept},
	}))
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
	e.Use(H.CheckRoleMiddleware())
	e.POST("/auth", H.Login)
	e.POST("/forgot", H.ForgotPassword)
	e.GET("/pingAuth", H.PingAuth)
	e.GET("/Time", H.GetTime)
	e.GET("/user/profile", H.GetProfile)
	e.PUT("/user/profile", H.UpdateProfile)
	e.GET("/user/organizations", H.GetUserOrganizations)
	e.GET("/user/processes", H.GetUserProcesses)
	e.GET("/user/incidents", H.GetUserIncidents)
	e.POST("/user/incidents/details", H.PostIncidentDetails)
	e.DELETE("/admin/users", H.DeleteUser)
	e.GET("/admin/users", H.GetUsers)
	e.POST("/admin/users", H.CreateUser)
	e.PUT("/admin/users", H.UpdateUser)
	e.GET("/admin/users/roles", H.GetAllRoles)
	e.POST("/admin/organization", H.CreateOrganization)
	e.PUT("/admin/organization", H.UpdateOrganization)
	e.DELETE("/admin/organization", H.DeleteOrganization)
	e.GET("/admin/organization", H.GetOrganizations)
	e.POST("/admin/organization/client", H.CreateUpdateOrganizationClient)
	e.DELETE("/admin/organization/client", H.DeleteOrganizationClient)
	e.PUT("/admin/organization/process", H.UpdateProcessAlias)
	e.PUT("/admin/organization/user", H.UpdateOrganizationUser)
	e.PATCH("/admin/UpdateUipathProcess", H.UpdateUipathProcess)
	e.GET("/user/processes", H.GetProcesses)
	e.GET("/user/processes/:id", H.GetProcess)
	e.DELETE("/user/processes/:id/clients", H.DeleteClientsFromProcess)
	e.POST("/user/processes/:id/clients", H.AddClientsToProcess)
	e.POST("/user/processes/:id/users", H.AddUsersToProcess)
	e.DELETE("/user/processes/:id/users", H.DeleteUsersFromProcess)
	e.PUT("/user/processes/:id", H.UpdateProcess)
	e.GET("/user/processes/:id/possibleUsers", H.GetPossibleUsers)
	e.GET("/user/processes/:id/possibleClients", H.GetPossibleClients)
	e.POST("/user/processes/:id/newIncident", H.NewIncident)
	e.POST("/monitor/:id/newIncident", H.NewIncident)
	e.PATCH("/monitor/RefreshOrgs", H.UpdateUipathProcess)
	e.GET("/monitor/Orgs", H.GetOrgs)
	e.PUT("/monitor/UpdateExceptionJob", H.UpdateExceptionJob)
	e.PATCH("/monitor/PatchJobHistory", H.PatchJobHistory)
	e.GET("/client/tickets", H.GetClientTicket)
	e.GET("/admin/downloads/orgs", H.GetOrgData)
	e.GET("/admin/downloads/users", H.GetUserData)
	H.RefreshDataBase(e)
	// listener, err := localtunnel.Listen(localtunnel.Options{
	// 	Subdomain: "golanguipathmonitortunnel",
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Tunnel URL: " + listener.URL())
	// e.Listener = listener
	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	err = e.Start(":" + port)
	if err != nil {
		fmt.Println(err)
	}
}
