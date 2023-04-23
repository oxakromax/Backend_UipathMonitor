package ORM

import "gorm.io/gorm"

type Cliente struct {
	gorm.Model
	Nombre         string `gorm:"not null"`
	Apellido       string `gorm:"not null"`
	Email          string `gorm:"not null"`
	OrganizacionID uint   `gorm:"not null"`
	Organizacion   *Organizacion
	Procesos       []*Proceso `gorm:"many2many:procesos_clientes;"`
}

func (Cliente) TableName() string {
	return "clientes"
}

func (Cliente) GetAll(db *gorm.DB) []*Cliente {
	var clientes []*Cliente
	db.Preload("Organizacion").Preload("Procesos").Find(&clientes)
	return clientes
}

func (this *Cliente) Get(db *gorm.DB, id uint) {
	db.Preload("Organizacion").Preload("Procesos").First(&this, id)
}
