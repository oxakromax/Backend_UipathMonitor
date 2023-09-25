package ORM

import (
	"gorm.io/gorm"
)

type TicketsTipo struct {
	gorm.Model
	Nombre              string `json:"nombre" gorm:"not null"`
	NecesitaDiagnostico bool   `gorm:"default:false"`
}

type TicketsProceso struct {
	gorm.Model
	ProcesoID        uint              `gorm:"not null"`
	Proceso          *Proceso          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Descripcion      string            `gorm:"type:text" json:"Descripcion"`
	Tipo             *TicketsTipo      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"TipoDetail"`
	TipoID           uint              `gorm:"not null" json:"Tipo"`
	Prioridad        uint              `gorm:"not null;default:5"`
	Estado           string            `gorm:"not null;type:estado_enum;default:'Iniciado'" json:"Estado"`
	UsuarioCreador   *Usuario          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"UsuarioCreadorDetail"`
	UsuarioCreadorID int               `gorm:"not null" json:"UsuarioCreadorID"`
	Detalles         []*TicketsDetalle `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (tp *TicketsProceso) Get(db *gorm.DB, id uint) {
	db.Preload("Proceso").Preload("Detalles").Preload("Tipo").First(&tp, id)
}

func (tp *TicketsProceso) GetTipo(db *gorm.DB) string {
	// "Incidente": 1,
	// "Mejora": 2,
	// "Mantenimiento": 3,
	// "Otro": 4,
	if tp.Tipo == nil && db != nil { // If the Tipo is nil, get it from the database
		Tipo := new(TicketsTipo)
		db.First(&Tipo, tp.TipoID)
		tp.Tipo = Tipo
	} else if tp.Tipo == nil { // If the Tipo is nil and the database is nil, return empty string
		return ""
	}
	return tp.Tipo.Nombre // Return the Tipo name
}

// TableName IncidenteProcesos Tablename: incidentes_procesos
func (TicketsProceso) TableName() string {
	return "tickets_procesos"
}
