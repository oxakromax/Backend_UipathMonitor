package main

import (
	"fmt"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/Server"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"strings"
)

func OpenDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=Nh52895390 dbname=Proyecto port=5432 sslmode=disable"
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

func main() {
	e := echo.New()
	H := &Server.Handler{
		Db:                  OpenDB(),
		TokenKey:            functions.GeneratePassword(32),
		UniversalRoutes:     []string{"/auth", "/forgot"},
		UserUniversalRoutes: []string{"/user/profile"},
		DbKey:               os.Getenv("DB_KEY"),
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
	e.Use(H.CheckRoleMiddleware())
	e.POST("/auth", H.Login)
	e.POST("/forgot", H.ForgotPassword)
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
	e.POST("/admin/organization", H.CreateOrganization)
	e.PUT("/admin/organization", H.UpdateOrganization)
	e.DELETE("/admin/organization", H.DeleteOrganization)
	e.GET("/admin/organization", H.GetOrganizations)
	e.POST("/admin/organization/client", H.CreateUpdateOrganizationClient)
	e.DELETE("/admin/organization/client", H.DeleteOrganizationClient)
	e.PUT("/admin/organization/process", H.UpdateProcessAlias)
	e.PUT("/admin/organization/user", H.UpdateOrganizationUser)
	e.PATCH("/admin/UpdateUipathProcess", H.UpdateUipathProcess)
	H.RefreshDataBase(e)
	var err error
	//listener, err := localtunnel.Listen(localtunnel.Options{
	//	Subdomain: "golanguipathmonitortunnel",
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println("Tunnel URL: " + listener.URL())
	//e.Listener = listener
	err = e.Start(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
