package main

import (
	"GormTest/ORM"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db := OpenDB()
	user := new(ORM.Usuario).GetByProcess(db, 1)
	fmt.Println(user)
	client := new(ORM.Cliente).GetByProcess(db, 1)
	fmt.Println(client)
	process := new(ORM.Proceso)
	process.Get(db, 1)
	fmt.Println(process)

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
