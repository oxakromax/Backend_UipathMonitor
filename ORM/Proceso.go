package ORM

import "gorm.io/gorm"

type Proceso struct {
	gorm.Model
	Nombre            string              `gorm:"not null"`
	Alias             string              `gorm:"not null,default:''"`
	UipathProcessID   uint                `gorm:"not null,default:0"`
	Folderid          uint                `gorm:"not null"`
	Foldername        string              `gorm:"not null,default:''"`
	OrganizacionID    uint                `gorm:"not null"`
	WarningTolerance  int                 `gorm:"not null;default:10"`
	ErrorTolerance    int                 `gorm:"not null;default:0"`
	FatalTolerance    int                 `gorm:"not null;default:0"`
	ActiveMonitoring  bool                `gorm:"not null;default:false"`
	Organizacion      *Organizacion       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	IncidentesProceso []*IncidenteProceso `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Clientes          []*Cliente          `gorm:"many2many:procesos_clientes;"`
	Usuarios          []*Usuario          `gorm:"many2many:procesos_usuarios;"`
}

func (Proceso) TableName() string {
	return "procesos"
}

func (Proceso) GetAll(db *gorm.DB) []*Proceso {
	var procesos []*Proceso
	db.Preload("Organizacion").Preload("IncidentesProceso").Preload("Clientes").Preload("Usuarios").Preload("IncidentesProceso.Detalles").Find(&procesos)
	return procesos
}

func (this *Proceso) Get(db *gorm.DB, id uint) {
	db.Preload("Organizacion").Preload("IncidentesProceso").Preload("Clientes").Preload("Usuarios").Preload("IncidentesProceso.Detalles").First(&this, id)
}

func (Proceso) GetByOrganizacion(db *gorm.DB, organizacionID uint) []*Proceso {
	var procesos []*Proceso
	db.Preload("Organizacion").Preload("IncidentesProceso").Preload("Clientes").Preload("Usuarios").Preload("IncidentesProceso.Detalles").Where("organizacion_id = ?", organizacionID).Find(&procesos)
	return procesos
}

func (Proceso) GetByFolder(db *gorm.DB, folderID uint) []*Proceso {
	var procesos []*Proceso
	db.Preload("Organizacion").Preload("IncidentesProceso").Preload("Clientes").Preload("Usuarios").Preload("IncidentesProceso.Detalles").Where("folderid = ?", folderID).Find(&procesos)
	return procesos
}

func (this *Proceso) GetEmails() []string {
	var emails []string
	for _, cliente := range this.Clientes {
		emails = append(emails, cliente.Email)
	}
	for _, usuario := range this.Usuarios {
		emails = append(emails, usuario.Email)
	}
	return emails
}
