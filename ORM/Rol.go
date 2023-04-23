package ORM

import "gorm.io/gorm"

type Rol struct {
	gorm.Model
	Nombre      string     `gorm:"not null"`
	Usuarios    []*Usuario `gorm:"many2many:usuarios_roles;"`
	Rutas       []*Route   `gorm:"many2many:roles_routes;"`
	Description string     `gorm:"not null default:''"`
}

func (Rol) GetAll(db *gorm.DB) []*Rol {
	var roles []*Rol
	db.Preload("Rutas").Find(&roles)
	return roles
}

func (this *Rol) Get(db *gorm.DB, id uint) {
	db.Preload("Rutas").First(&this, id)
}

func (this *Rol) GetUsuarios(db *gorm.DB) {
	err := db.Model(&this).Association("Usuarios").Find(&this.Usuarios)
	if err != nil {
		return
	}
}

func (Rol) TableName() string {
	return "roles"
}
