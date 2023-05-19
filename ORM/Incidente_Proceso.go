package ORM

import "gorm.io/gorm"

type IncidenteProceso struct {
	gorm.Model
	ProcesoID uint                 `gorm:"not null"`
	Proceso   *Proceso             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Incidente string               `gorm:"type:text"`
	Tipo      int                  `gorm:"not null;default:1"`
	Estado    int                  `gorm:"not null;default:1"`
	Detalles  []*IncidentesDetalle `gorm:"foreignKey:IncidenteID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (this *IncidenteProceso) Get(db *gorm.DB, id uint) {
	db.Preload("Proceso").Preload("Detalles").First(&this, id)
}

func (this *IncidenteProceso) GetTipo() string {
	// "Incidente": 1,
	// "Mejora": 2,
	// "Mantenimiento": 3,
	// "Otro": 4,
	switch this.Tipo {
	case 1:
		return "Incidente"
	case 2:
		return "Mejora"
	case 3:
		return "Mantenimiento"
	case 4:
		return "Otro"
	default:
		return "Incidente"
	}
}

// TableName IncidenteProcesos Tablename: incidentes_procesos
func (IncidenteProceso) TableName() string {
	return "incidentes_procesos"
}
