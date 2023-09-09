package main

import (
	"fmt"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/Server"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
)

func OpenDB() *gorm.DB {
	Host := os.Getenv("PGHOST")
	User := os.Getenv("PGUSER")
	Password := os.Getenv("PGPASSWORD")
	Database := os.Getenv("PGDATABASE")
	Port := os.Getenv("PGPORT")
	SSLMode := os.Getenv("PGSSLMODE")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", Host, User, Password, Database, Port, SSLMode)
	db, err := gorm.Open(postgres.Open(dsn),
		&gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		},
	)
	db.Logger = db.Logger.LogMode(4) // 4 = Info
	db.Logger.Info(nil, "Database connection successfully opened")
	if err != nil {
		panic("failed to connect database")
	}
	createEnumTypeIfNotExists(db)
	err = db.AutoMigrate(
		&ORM.Organizacion{}, &ORM.Cliente{}, &ORM.Proceso{},
		&ORM.TicketsTipo{}, &ORM.TicketsProceso{}, &ORM.TicketsDetalle{}, &ORM.Usuario{},
		&ORM.Rol{}, &ORM.Route{}, &ORM.JobHistory{}, &ORM.LogJobHistory{},
	)
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
	var err = godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, please be sure ENV variables are set")
	}
	e := echo.New()
	H := &Server.Handler{
		TokenKey:            os.Getenv("TOKEN_KEY"),
		UniversalRoutes:     []string{"/auth", "/forgot", "/client/tickets", "/Time"},
		UserUniversalRoutes: []string{"/user/profile", "/pingAuth"},
		DBKey:               os.Getenv("DB_KEY"),
	}
	if H.DBKey == "" {
		fmt.Println("DB_KEY environment variable not set")
		NewKey, _ := functions.GenerateAESKey()
		fmt.Println("Generated key: " + NewKey)
		return // Exit, don't start the server
	}
	middlewares(e, H)
	routes(e, H)

	go func() {
		// Fl0 needs to open the port in less than 60 seconds, so we do it in a goroutine
		H.DB = OpenDB()
		H.RefreshDataBase(e)
	}()
	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	if SSlCert, SSlKey := os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"); SSlCert != "" && SSlKey != "" {
		err = e.StartTLS(":"+port, SSlCert, SSlKey)
		if err != nil {
			panic(err)
		}
	} else {
		err = e.Start(":" + port)
		if err != nil {
			panic(err)
		}
	}
}

func middlewares(e *echo.Echo, h *Server.Handler) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(h.TokenKey),
		SigningMethod: "HS512",
		Skipper: func(c echo.Context) bool {
			// Skip authentication for signup and login requests
			for _, route := range h.UniversalRoutes {
				if strings.EqualFold(route, c.Path()) {
					return true
				}
			}
			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, "Invalid or expired JWT")
		},
	}))
	e.Use(h.CheckDBState()) // Check if the database is connected
	e.Use(h.CheckRoleMiddleware())
}

func routes(e *echo.Echo, H *Server.Handler) {
	e.GET("/user/profile", H.GetProfile)
	e.GET("/user/tickets", H.GetUserTickets)
	e.GET("/user/tickets/details", H.GetTicketSettings)
	e.GET("/user/organizations", H.GetUserOrganizations)
	e.GET("/user/processes", H.GetUserProcesses)
	e.GET("/user/processes/:id", H.GetProcess)
	e.GET("/user/processes/:id/possibleClients", H.GetPossibleClients)
	e.GET("/user/processes/:id/possibleUsers", H.GetPossibleUsers)
	e.GET("/user/processes/TicketsType", H.GetTicketsType)
	e.POST("/user/tickets/details", H.PostIncidentDetails)
	e.POST("/user/processes/:id/clients", H.AddClientsToProcess)
	e.POST("/user/processes/:id/newTicket", H.NewTicket)
	e.POST("/user/processes/:id/users", H.AddUsersToProcess)
	e.PUT("/user/processes/:id", H.UpdateProcess)
	e.PUT("/user/profile", H.UpdateProfile)
	e.DELETE("/user/processes/:id/clients", H.DeleteClientsFromProcess)
	e.GET("/admin/downloads/orgs", H.GetOrgData)
	e.GET("/admin/downloads/processes", H.GetProcessData)
	e.GET("/admin/downloads/users", H.GetUserData)
	e.GET("/admin/organization", H.GetOrganizations)
	e.PUT("/admin/organization", H.UpdateOrganization)
	e.PUT("/admin/organization/process", H.UpdateProcessAlias)
	e.PUT("/admin/organization/user", H.UpdateOrganizationUser)
	e.POST("/admin/organization", H.CreateOrganization)
	e.POST("/admin/organization/client", H.CreateUpdateOrganizationClient)
	e.DELETE("/admin/organization", H.DeleteOrganization)
	e.DELETE("/admin/organization/client", H.DeleteOrganizationClient)
	e.GET("/admin/users", H.GetUsers)
	e.GET("/admin/users/roles", H.GetAllRoles)
	e.PUT("/admin/users", H.UpdateUser)
	e.POST("/admin/users", H.CreateUser)
	e.DELETE("/admin/users", H.DeleteUser)
	e.PATCH("/admin/UpdateUipathProcess", H.UpdateUipathProcess)
	e.GET("/monitor/Orgs", H.GetOrgs)
	e.POST("/monitor/:id/newTicket", H.NewTicket)
	e.PUT("/monitor/UpdateExceptionJob", H.UpdateExceptionJob)
	e.PATCH("/monitor/RefreshOrgs", H.UpdateUipathProcess)
	e.PATCH("/monitor/PatchJobHistory", H.PatchJobHistory)
	e.GET("/pingAuth", H.PingAuth)
	e.GET("/Time", H.GetTime)
	e.GET("/client/tickets", H.GetClientTicket)
	e.POST("/auth", H.Login)
	e.POST("/forgot", H.ForgotPassword)
}
