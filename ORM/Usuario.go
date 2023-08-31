package ORM

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Usuario struct {
	gorm.Model
	Nombre           string            `gorm:"not null"`
	Apellido         string            `gorm:"not null"`
	Email            string            `gorm:"not null"`
	Password         string            `gorm:"not null"`
	Roles            []*Rol            `gorm:"many2many:usuarios_roles;"`
	Procesos         []*Proceso        `gorm:"many2many:procesos_usuarios;"`
	Organizaciones   []*Organizacion   `gorm:"many2many:usuarios_organizaciones;"`
	Tickets_Procesos []*TicketsProceso `gorm:"foreignKey:UsuarioCreadorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Tickets_Detalle  []*TicketsDetalle `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (u *Usuario) GetAdmins(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("roles.Nombre = ?", "admin").Find(&usuarios)
	return usuarios
}

func (u *Usuario) SetPassword(password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.Password = string(hash)
}

func (u *Usuario) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *Usuario) GetByEmail(db *gorm.DB, email string) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("email = ?", email).First(&u)
}

func (Usuario) TableName() string {
	return "usuarios"
}

func (Usuario) GetAll(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Tickets_Procesos").Preload("Tickets_Detalle").Find(&usuarios)
	return usuarios
}

func (u *Usuario) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Tickets_Procesos").Preload("Tickets_Detalle").First(&u, id)
}

func (u *Usuario) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r.Nombre == role || r.Nombre == "admin" {
			return true
		}
	}
	return false
}

func (u *Usuario) HasProcess(process int) bool {
	for _, p := range u.Procesos {
		if int(p.ID) == process {
			return true
		}
	}
	return false
}

func (u *Usuario) FillEmptyFields(db *gorm.DB) {
	// Check empty fields of the user and fill them with the values from the database
	var user Usuario
	db.First(&user, u.ID)
	if u.Nombre == "" {
		u.Nombre = user.Nombre
	}
	if u.Apellido == "" {
		u.Apellido = user.Apellido
	}
	if u.Email == "" {
		u.Email = user.Email
	}
	if u.Password == "" {
		u.Password = user.Password
	}
	if len(u.Roles) == 0 {
		u.Roles = user.Roles
	}
	if len(u.Procesos) == 0 {
		u.Procesos = user.Procesos
	}
	if len(u.Organizaciones) == 0 {
		u.Organizaciones = user.Organizaciones
	}
}

func (u *Usuario) GetComplete(db *gorm.DB) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Organizaciones.Procesos").Preload("Tickets_Procesos").Preload("Tickets_Detalle").First(&u, u.ID)
}

func (u *Usuario) GetCantityTicketsClosed(db *gorm.DB) int64 {
	var Tickets []*TicketsProceso
	db.Preload("Detalles", "usuario_id = ?", u.ID).Where("estado = ?", "Finalizado").Find(&Tickets)
	// El ultimo detalle es el que da la finalización, si lo hizo el usuario, se cuenta
	var count int64 = 0
	for _, ticket := range Tickets {
		if len(ticket.Detalles) == 0 {
			continue
		}
		if ticket.Detalles[len(ticket.Detalles)-1].UsuarioID == int(u.ID) {
			count++
		}
	}
	return count
}

func (u *Usuario) GetCantityTicketsCreated(db *gorm.DB) int64 {
	var Tickets []*TicketsProceso
	db.Where("usuario_creador_id = ?", u.ID).Find(&Tickets)
	return int64(len(Tickets))
}

func (u *Usuario) GetCantityTicketsPending(db *gorm.DB) int64 {
	if u.Procesos == nil {
		u.GetComplete(db)
	}
	count := 0
	for _, proceso := range u.Procesos {
		proceso.TicketsProcesos = nil
		db.Where("estado != ?", "Finalizado").Find(&proceso.TicketsProcesos)
		count += len(proceso.TicketsProcesos)
	}
	return int64(count)
}

func (u *Usuario) AverageDurationPerTicket(db *gorm.DB) time.Duration {
	var Tickets []*TicketsProceso
	db.Preload("Detalles", "usuario_id = ?", u.ID).Find(&Tickets)
	var total time.Duration = 0
	for _, ticket := range Tickets {
		if len(ticket.Detalles) == 0 {
			continue
		}
		// sumamos la diferencia entre la fecha de inicio y la fecha de fin de cada detalle realizado por el usuario
		for _, detalle := range ticket.Detalles {
			if detalle.UsuarioID == int(u.ID) {
				total += detalle.FechaFin.Sub(detalle.FechaInicio)
			}
		}
	}
	if len(Tickets) == 0 {
		return 0
	}
	return total / time.Duration(len(Tickets))
}

// Average time spent in tickets a day
func (u *Usuario) AverageTimeSpentPerDay(db *gorm.DB) time.Duration {
	var Tickets []*TicketsDetalle
	db.Where("usuario_id = ?", u.ID).Order("fecha_inicio").Find(&Tickets)
	// Discriminar por dia
	var total time.Duration = 0
	var dias = 0
	var dia time.Time
	for _, ticket := range Tickets {
		total += ticket.FechaFin.Sub(ticket.FechaInicio)
		if dia.Day() != ticket.FechaInicio.Day() {
			dia = ticket.FechaInicio
			dias++
		}
	}
	if dias == 0 {
		return 0
	}
	return total / time.Duration(dias)
}
