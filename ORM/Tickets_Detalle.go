package ORM

import (
	"time"

	"gorm.io/gorm"
)

type TicketsDetalle struct {
	gorm.Model
	IncidenteID int       `gorm:"not null"`
	Detalle     string    `gorm:"type:text" json:"Detalle"`
	FechaInicio time.Time `gorm:"precision:6"`
	FechaFin    time.Time `gorm:"precision:6"`
	UsuarioID   int       `gorm:"not null"`
}

// TableName TableName IncidentesDetalles Tablename: incidentes_detalle
func (TicketsDetalle) TableName() string {
	return "tickets_detalle"
}
