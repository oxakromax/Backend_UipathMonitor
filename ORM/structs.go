package ORM

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Organizacion struct {
	gorm.Model
	Nombre     string `gorm:"not null"`
	Uipathname string `gorm:"not null"`
	Tenantname string `gorm:"not null"`
	AppID      string `gorm:"not null;default:''"`
	AppSecret  string `gorm:"not null;default:''"`
	AppScope   string `gorm:"not null;default:''"`
	Clientes   []*Cliente
	Procesos   []*Proceso
}

func (Organizacion) GetAll(db *gorm.DB) []*Organizacion {
	var organizaciones []*Organizacion
	db.Preload("Procesos").Preload("Clientes").Find(&organizaciones)
	return organizaciones
}

func (this *Organizacion) Get(db *gorm.DB, id uint) {
	db.Preload("Procesos").Preload("Clientes").First(&this, id)
}

func (Organizacion) TableName() string {
	return "organizaciones"
}

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

func (Cliente) GetByProcess(db *gorm.DB, id uint) []*Cliente {
	var clientes []*Cliente
	db.Preload("Organizacion").Preload("Procesos").Joins("JOIN procesos_clientes ON procesos_clientes.cliente_id = clientes.id").Where("procesos_clientes.proceso_id = ?", id).Find(&clientes)
	return clientes
}

type Proceso struct {
	gorm.Model
	Nombre            string              `gorm:"not null"`
	Folderid          uint                `gorm:"not null"`
	OrganizacionID    uint                `gorm:"not null"`
	WarningTolerance  int                 `gorm:"not null;default:10"`
	ErrorTolerance    int                 `gorm:"not null;default:0"`
	FatalTolerance    int                 `gorm:"not null;default:0"`
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

type IncidenteProceso struct {
	gorm.Model
	ProcesoID uint                 `gorm:"not null"`
	Proceso   *Proceso             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Incidente string               `gorm:"type:text"`
	Tipo      int                  `gorm:"not null;default:1"`
	Estado    int                  `gorm:"not null;default:1"`
	Detalles  []*IncidentesDetalle `gorm:"foreignKey:IncidenteID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (IncidenteProceso) GetAll(db *gorm.DB) []*IncidenteProceso {
	var incidentes []*IncidenteProceso
	db.Preload("Proceso").Preload("Detalles").Find(&incidentes)
	return incidentes
}

func (this *IncidenteProceso) Get(db *gorm.DB, id uint) {
	db.Preload("Proceso").Preload("Detalles").First(&this, id)
}

func (this *IncidenteProceso) GetByProceso(db *gorm.DB, procesoID uint) []*IncidenteProceso {
	var incidentes []*IncidenteProceso
	db.Preload("Proceso").Preload("Detalles").Where("proceso_id = ?", procesoID).Find(&incidentes)
	return incidentes
}

// IncidenteProcesos Tablename: incidentes_procesos
func (IncidenteProceso) TableName() string {
	return "incidentes_procesos"
}

type IncidentesDetalle struct {
	gorm.Model
	IncidenteID int       `gorm:"not null"`
	Detalle     string    `gorm:"type:text"`
	FechaInicio time.Time `gorm:"precision:6"`
	FechaFin    time.Time `gorm:"precision:6"`
}

func (IncidentesDetalle) GetAll(db *gorm.DB) []*IncidentesDetalle {
	var detalles []*IncidentesDetalle
	db.Find(&detalles)
	return detalles
}

func (this *IncidentesDetalle) Get(db *gorm.DB, id uint) {
	db.First(&this, id)
}

func (IncidentesDetalle) GetByIncidente(db *gorm.DB, incidenteID int) []*IncidentesDetalle {
	var detalles []*IncidentesDetalle
	db.Where("incidente_id = ?", incidenteID).Find(&detalles)
	return detalles
}

// IncidentesDetalles Tablename: incidentes_detalle
func (IncidentesDetalle) TableName() string {
	return "incidentes_detalle"
}

type Usuario struct {
	gorm.Model
	Nombre   string     `gorm:"not null"`
	Apellido string     `gorm:"not null"`
	Email    string     `gorm:"not null"`
	Password string     `gorm:"not null"`
	Roles    []*Rol     `gorm:"many2many:usuarios_roles;"`
	Procesos []*Proceso `gorm:"many2many:procesos_usuarios;"`
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
	db.Preload("Roles").Preload("Procesos").Where("email = ?", email).First(&this)
}

func (Usuario) TableName() string {
	return "usuarios"
}

func (Usuario) GetAll(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Find(&usuarios)
	return usuarios
}

func (this *Usuario) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").Preload("Procesos").First(&this, id)
}

func (Usuario) GetByProcess(db *gorm.DB, procesoID uint) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Joins("JOIN procesos_usuarios ON procesos_usuarios.usuario_id = usuarios.id").Where("procesos_usuarios.proceso_id = ?", procesoID).Find(&usuarios)
	return usuarios
}

type Rol struct {
	gorm.Model
	Nombre   string     `gorm:"not null"`
	Usuarios []*Usuario `gorm:"many2many:usuarios_roles;"`
}

func (Rol) GetAll(db *gorm.DB) []*Rol {
	var roles []*Rol
	db.Preload("Usuarios").Find(&roles)
	return roles
}

func (this *Rol) Get(db *gorm.DB, id uint) {
	db.Preload("Usuarios").First(&this, id)
}

func (Rol) TableName() string {
	return "roles"
}
