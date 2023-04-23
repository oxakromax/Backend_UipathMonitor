package ORM

import (
	"gorm.io/gorm"
	"time"
)

type IncidentesDetalle struct {
	gorm.Model
	IncidenteID int       `gorm:"not null"`
	Detalle     string    `gorm:"type:text"`
	FechaInicio time.Time `gorm:"precision:6"`
	FechaFin    time.Time `gorm:"precision:6"`
}

// TableName TableName IncidentesDetalles Tablename: incidentes_detalle
func (IncidentesDetalle) TableName() string {
	return "incidentes_detalle"
}
