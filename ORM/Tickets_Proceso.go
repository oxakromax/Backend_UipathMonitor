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
	Estado           string            `gorm:"not null;type:estado_enum;default:'Iniciado'" json:"Estado"`
	UsuarioCreador   *Usuario          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"UsuarioCreadorDetail"`
	UsuarioCreadorID int               `gorm:"not null" json:"UsuarioCreadorID"`
	Detalles         []*TicketsDetalle `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
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

// TableName IncidenteProcesos Tablename: incidentes_procesos
func (TicketsProceso) TableName() string {
	return "tickets_procesos"
}
