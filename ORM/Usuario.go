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

func (this *Usuario) FillEmptyFields(db *gorm.DB) {
	// Check empty fields of the user and fill them with the values from the database
	var user Usuario
	db.First(&user, this.ID)
	if this.Nombre == "" {
		this.Nombre = user.Nombre
	}
	if this.Apellido == "" {
		this.Apellido = user.Apellido
	}
	if this.Email == "" {
		this.Email = user.Email
	}
	if this.Password == "" {
		this.Password = user.Password
	}
	if len(this.Roles) == 0 {
		this.Roles = user.Roles
	}
	if len(this.Procesos) == 0 {
		this.Procesos = user.Procesos
	}
	if len(this.Organizaciones) == 0 {
		this.Organizaciones = user.Organizaciones
	}
}

func (this *Usuario) GetComplete(db *gorm.DB) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Organizaciones.Procesos").First(&this, this.ID)
}
