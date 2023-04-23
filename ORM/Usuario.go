package ORM

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Usuario struct {
	gorm.Model
	Nombre         string          `gorm:"not null"`
	Apellido       string          `gorm:"not null"`
	Email          string          `gorm:"not null"`
	Password       string          `gorm:"not null"`
	Roles          []*Rol          `gorm:"many2many:usuarios_roles;"`
	Procesos       []*Proceso      `gorm:"many2many:procesos_usuarios;"`
	Organizaciones []*Organizacion `gorm:"many2many:usuarios_organizaciones;"`
}

func (this *Usuario) GetAdmins(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("roles.Nombre = ?", "admin").Find(&usuarios)
	return usuarios
}

func (this *Usuario) SetPassword(password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	this.Password = string(hash)
}

func (this *Usuario) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(this.Password), []byte(password))
	return err == nil
}

func (this *Usuario) GetByEmail(db *gorm.DB, email string) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("email = ?", email).First(&this)
}

func (Usuario) TableName() string {
	return "usuarios"
}

func (Usuario) GetAll(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Find(&usuarios)
	return usuarios
}

func (this *Usuario) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").First(&this, id)
}

func (this *Usuario) GetComplete(db *gorm.DB) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Organizaciones.Procesos").First(&this, this.ID)
}
