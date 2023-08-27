package ORM

import (
	"gorm.io/gorm"
	"time"
)

type JobHistory struct {
	gorm.Model
	Proceso          *Proceso  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ProcesoID        uint      `gorm:"not null"`
	CreationTime     time.Time `gorm:"type:timestamptz(3);precision:3"`
	StartTime        time.Time `gorm:"type:timestamptz(3);precision:3"`
	EndTime          time.Time `gorm:"type:timestamptz(3);precision:3"`
	HostMachineName  string
	Source           string           `gorm:"not null"`
	State            string           `gorm:"not null"`
	JobKey           string           `gorm:"not null;unique"`
	JobID            int              `gorm:"not null;unique"`
	Duration         time.Duration    `gorm:"not null"`
	MonitorException bool             `gorm:"not null;default:false"`
	Logs             []*LogJobHistory `gorm:"foreignKey:JobID"`
}

func (this *JobHistory) Get(db *gorm.DB, id uint) {
	// preload logs and order logs by newest first (timestamp)
	db.Preload("Logs", func(db *gorm.DB) *gorm.DB {
		return db.Order("time_stamp DESC")
	}).First(&this, id)
}

type LogJobHistory struct {
	gorm.Model
	Level           string      `gorm:"not null"`
	WindowsIdentity string      `gorm:"not null"`
	ProcessName     string      `gorm:"not null"`
	TimeStamp       time.Time   `gorm:"type:timestamptz(3);precision:3"`
	Message         string      `gorm:"not null"`
	JobKey          string      `gorm:"not null"`
	RawMessage      string      `gorm:"not null"`
	RobotName       string      `gorm:"not null"`
	HostMachineName string      `gorm:"not null"`
	MachineId       int         `gorm:"not null"`
	MachineKey      string      `gorm:"not null"`
	JobID           int         `gorm:"not null"`
	Job             *JobHistory `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

type Proceso struct {
	gorm.Model
	Nombre           string            `gorm:"not null"`
	Alias            string            `gorm:"not null,default:''"`
	UipathProcessID  uint              `gorm:"not null,default:0"`
	Folderid         uint              `gorm:"not null"`
	Foldername       string            `gorm:"not null,default:''"`
	OrganizacionID   uint              `gorm:"not null"`
	WarningTolerance int               `gorm:"not null;default:10"`
	ErrorTolerance   int               `gorm:"not null;default:0"`
	FatalTolerance   int               `gorm:"not null;default:0"`
	ActiveMonitoring bool              `gorm:"not null;default:false"`
	Priority         int               `gorm:"not null;default:1"`
	MaxQueueTime     int               `gorm:"not null;default:30"`
	Organizacion     *Organizacion     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	TicketsProcesos  []*TicketsProceso `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"IncidentesProceso"`
	Clientes         []*Cliente        `gorm:"many2many:procesos_clientes;"`
	Usuarios         []*Usuario        `gorm:"many2many:procesos_usuarios;"`
	JobsHistory      []*JobHistory     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (Proceso) TableName() string {
	return "procesos"
}

func (Proceso) GetAll(db *gorm.DB) []*Proceso {
	var procesos []*Proceso
	db.Preload("Organizacion").Preload("Organizacion.Clientes").Preload("Organizacion.Usuarios").Preload("TicketsProcesos").Preload("Clientes").Preload("Usuarios").Preload("TicketsProcesos.Detalles").Find(&procesos)
	return procesos
}

func (this *Proceso) Get(db *gorm.DB, id uint) {
	db.Preload("Organizacion").Preload("TicketsProcesos").Preload("Clientes").Preload("Usuarios").Preload("TicketsProcesos.Detalles").Preload("JobsHistory").First(&this, id)
}

func (Proceso) GetByFolder(db *gorm.DB, folderID uint) []*Proceso {
	var procesos []*Proceso
	db.Preload("Organizacion").Preload("TicketsProcesos").Preload("Clientes").Preload("Usuarios").Preload("TicketsProcesos.Detalles").Preload("TicketsProcesos.Tipo").Where("folderid = ?", folderID).Find(&procesos)
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

func GetListByUser(db *gorm.DB, userID uint) []*Proceso {
	var procesos []*Proceso
	db.Preload("TicketsProcesos").Preload("Organizacion").Preload("Clientes").Preload("Usuarios").Preload("TicketsProcesos.Detalles").Preload("TicketsProcesos.Tipo").Joins("JOIN procesos_usuarios ON procesos_usuarios.proceso_id = procesos.id").Where("procesos_usuarios.usuario_id = ?", userID).Find(&procesos)
	return procesos
}

func (this *Proceso) RemoveClients(db *gorm.DB, list []int) error {
	var clientes []*Cliente
	for _, id := range list {
		cliente := &Cliente{}
		cliente.Get(db, uint(id))
		clientes = append(clientes, cliente)
	}
	return db.Model(&this).Association("Clientes").Delete(&clientes)
}

func (this *Proceso) AddClients(db *gorm.DB, list []int) error {
	var clientes []*Cliente
	for _, id := range list {
		cliente := &Cliente{}
		cliente.Get(db, uint(id))
		clientes = append(clientes, cliente)
	}
	return db.Model(&this).Association("Clientes").Append(&clientes)
}

func (this Proceso) RemoveUsers(db *gorm.DB, list []int) error {
	var usuarios []*Usuario
	for _, id := range list {
		usuario := &Usuario{}
		usuario.Get(db, uint(id))
		usuarios = append(usuarios, usuario)
	}
	return db.Model(&this).Association("Usuarios").Delete(&usuarios)

}

func (this Proceso) AddUsers(db *gorm.DB, list []int) error {
	var usuarios []*Usuario
	for _, id := range list {
		usuario := &Usuario{}
		usuario.Get(db, uint(id))
		usuarios = append(usuarios, usuario)
	}
	return db.Model(&this).Association("Usuarios").Append(&usuarios)

}
