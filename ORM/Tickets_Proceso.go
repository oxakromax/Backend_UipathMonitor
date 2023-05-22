package ORM

import "gorm.io/gorm"

type TicketsProceso struct {
	gorm.Model
	ProcesoID   uint              `gorm:"not null"`
	Proceso     *Proceso          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Descripcion string            `gorm:"type:text" json:"Incidente"`
	Tipo        int               `gorm:"not null;default:1"`
	Estado      int               `gorm:"not null;default:1"`
	Detalles    []*TicketsDetalle `gorm:"foreignKey:IncidenteID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (this *TicketsProceso) Get(db *gorm.DB, id uint) {
	db.Preload("Proceso").Preload("Detalles").First(&this, id)
}

func (this *TicketsProceso) GetTipo() string {
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
		return "Desconocido"
	}
}

func (this *TicketsProceso) GetEstado() string {
	// "Iniciado": 1,
	// "En Progreso": 2,
	// "Finalizado": 3,
	switch this.Estado {
	case 1:
		return "Iniciado"
	case 2:
		return "En Progreso"
	case 3:
		return "Finalizado"
	default:
		return "Desconocido"
	}
}

// TableName IncidenteProcesos Tablename: incidentes_procesos
func (TicketsProceso) TableName() string {
	return "tickets_procesos"
}