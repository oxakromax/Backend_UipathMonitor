package ORM

import (
	"gorm.io/gorm"
)

type Route struct {
	gorm.Model
	Method string `gorm:"not null"`
	Route  string `gorm:"not null"`
	Roles  []*Rol `gorm:"many2many:roles_routes;"`
}

func (Route) GetAll(db *gorm.DB) []*Route {
	var routes []*Route
	db.Preload("Roles").Find(&routes)
	return routes
}

func (this *Route) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").First(&this, id)
}
