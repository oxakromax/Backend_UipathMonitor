package ORM

import (
	"time"

	"gorm.io/gorm"
)

type IncidentesDetalle struct {
	gorm.Model
	IncidenteID int       `gorm:"not null"`
	Detalle     string    `gorm:"type:text"`
	FechaInicio time.Time `gorm:"precision:6"`
	FechaFin    time.Time `gorm:"precision:6"`
	UsuarioID   int       `gorm:"not null"`
}

// TableName TableName IncidentesDetalles Tablename: incidentes_detalle
func (IncidentesDetalle) TableName() string {
	return "incidentes_detalle"
}
