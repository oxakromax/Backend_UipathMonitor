package ORM

import (
	"time"

	"gorm.io/gorm"
)

type TicketsTipo struct {
	gorm.Model
	Nombre string `json:"nombre" gorm:"not null"`
}

type IncidenteDiagnosticos struct {
	gorm.Model
	TicketID    uint      `gorm:"not null"`
	Diagnostico string    `gorm:"type:text" json:"Diagnostico"`
	UsuarioID   int       `gorm:"not null"`
	FechaInicio time.Time `gorm:"precision:6"`
	FechaFin    time.Time `gorm:"precision:6"`
}

type TicketsProceso struct {
	gorm.Model
	ProcesoID        uint                     `gorm:"not null"`
	Proceso          *Proceso                 `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Descripcion      string                   `gorm:"type:text" json:"Incidente"`
	Tipo             *TicketsTipo             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"TipoDetail"`
	TipoID           uint                     `gorm:"not null" json:"Tipo"`
	Estado           int                      `gorm:"not null;default:1"`
	UsuarioCreador   *Usuario                 `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"UsuarioCreadorDetail"`
	UsuarioCreadorID int                      `gorm:"not null" json:"UsuarioCreadorID"`
	Detalles         []*TicketsDetalle        `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Diagnosticos     []*IncidenteDiagnosticos `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (this *TicketsProceso) Get(db *gorm.DB, id uint) {
	db.Preload("Proceso").Preload("Detalles").Preload("Tipo").First(&this, id)
}

func (this *TicketsProceso) GetTipo(db *gorm.DB) string {
	// "Incidente": 1,
	// "Mejora": 2,
	// "Mantenimiento": 3,
	// "Otro": 4,
	if this.Tipo == nil {
		Tipo := &TicketsTipo{}
		db.First(&Tipo, this.TipoID)
		this.Tipo = Tipo
	}
	return this.Tipo.Nombre
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
