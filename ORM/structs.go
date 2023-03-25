package ORM

import (
	"encoding/json"
	"errors"
	"github.com/google/go-querystring/query"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	UipathBearerTokenMap sync.Map // sync map because there's a lot of goroutines reading and writing to it
)

func (this *Organizacion) RefreshUiPathToken() error {
	const url = "https://cloud.uipath.com/identity_/connect/token"
	const method = "POST"
	ClientID, _ := functions.DecryptAES(os.Getenv("DB_KEY"), this.AppID)
	ClientSecret, _ := functions.DecryptAES(os.Getenv("DB_KEY"), this.AppSecret)
	var QueryAuth = struct {
		GrantType    string `json:"grant_type" url:"grant_type"`
		ClientId     string `json:"client_id" url:"client_id"`
		ClientSecret string `json:"client_secret" url:"client_secret"`
		Scope        string `json:"scope" url:"scope"`
	}{
		GrantType:    "client_credentials",
		ClientId:     ClientID,
		ClientSecret: ClientSecret,
		Scope:        this.AppScope,
	}
	vals, err := query.Values(QueryAuth)
	if err != nil {
		return err
	}
	payload := strings.NewReader(vals.Encode())
	client := new(http.Client)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New("Error refreshing UiPath token: " + res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var UiPathToken struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	err = json.Unmarshal(body, &UiPathToken) // Refresh the token
	if err != nil {
		return err
	}
	UipathBearerTokenMap.Store(this.ID, UiPathToken.AccessToken)
	return res.Body.Close()
}

type Organizacion struct {
	gorm.Model
	Nombre     string `gorm:"not null"`
	Uipathname string `gorm:"not null"`
	Tenantname string `gorm:"not null"`
	AppID      string `gorm:"not null;default:''"`
	AppSecret  string `gorm:"not null;default:''"`
	AppScope   string `gorm:"not null;default:''"`
	BaseURL    string `gorm:"not null;default:'https://cloud.uipath.com/'"`
	Clientes   []*Cliente
	Procesos   []*Proceso
	Usuarios   []*Usuario `gorm:"many2many:usuarios_organizaciones;"`
}

func (this *Organizacion) GetUrl() string {
	return this.BaseURL + this.Uipathname + "/" + this.Tenantname + "/orchestrator_/odata/"
}

func (this *Organizacion) RequestAPI(method, path string, body io.Reader, conds ...interface{}) (*http.Response, error) {
	// Examples:
	//res, err := RequestAPI("GET", "/releases", nil, "application/xml") // contentType será "application/xml" y folderID será 0
	//res, err := RequestAPI("GET", "/releases", nil, 1234) // contentType será "application/json" y folderID será 1234
	//res, err := RequestAPI("GET", "/releases", nil, 1234, "application/xml") // contentType será "application/xml" y folderID será 1234

	// if there's no folder, then folderID = 0
	// Possible Paths:
	// Releases
	// Folders
	// Jobs
	// RobotLogs
	for i := 0; i < 2; i++ { // Try twice, in case the token is expired
		client := new(http.Client)
		req, err := http.NewRequest(method, this.GetUrl()+path, body)
		if err != nil {
			return nil, err
		}
		var UipathBearerToken string
		if val, ok := UipathBearerTokenMap.Load(this.ID); ok {
			UipathBearerToken = val.(string)
		} else {
			UipathBearerToken = ""
		}
		req.Header.Add("Authorization", "Bearer "+UipathBearerToken)
		contentType := "application/json"
		for _, cond := range conds {
			if cont, ok := cond.(string); ok {
				contentType = cont
			}
		}
		req.Header.Add("Content-Type", contentType)
		for _, cond := range conds {
			if folderID, ok := cond.(int); ok {
				req.Header.Add("X-UIPATH-OrganizationUnitId", strconv.Itoa(folderID))
			}
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 401 { // Unauthorized
			err = this.RefreshUiPathToken()
			if err != nil {
				return nil, err
			}
			continue
		}
		return res, nil
	}
	return nil, errors.New("error al obtener el token de UiPath")
}

func (this *Organizacion) GetFromApi(structType interface{}, folderid ...int) error {
	// Possible Paths:
	// Releases
	// Folders
	// Jobs
	// RobotLogs

	var resp *http.Response
	var err error
	var folderID int
	if len(folderid) > 0 {
		folderID = folderid[0]
	} else {
		folderID = 0
	}
	switch structType.(type) {
	case *UipathAPI.FoldersResponse:
		resp, err = this.RequestAPI("GET", "Folders", nil)
	case *UipathAPI.LogResponse:
		resp, err = this.RequestAPI("GET", "RobotLogs", nil, folderID)
	case *UipathAPI.ReleasesResponse:
		resp, err = this.RequestAPI("GET", "Releases", nil, folderID)
	case *UipathAPI.JobsResponse:
		resp, err = this.RequestAPI("GET", "Jobs", nil, folderID)
	default:
		return errors.New("tipo de estructura no soportada, debe ser un puntero a una de las siguientes estructuras: FoldersResponse, LogResponse, ReleasesResponse, JobsResponse")
	}
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, structType)
	if err != nil {
		return err
	}
	return resp.Body.Close() // last possible error
}

func (this *Organizacion) CheckAccessAPI() error {
	_, err := this.RequestAPI("GET", "Folders", nil)
	return err
}

func (this *Organizacion) GetAll(db *gorm.DB) []*Organizacion {
	var organizaciones []*Organizacion
	db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").Find(&organizaciones)
	return organizaciones
}

func (this *Organizacion) Get(db *gorm.DB, id uint) {
	db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").First(&this, id)
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
	Alias             string              `gorm:"not null,default:''"`
	Folderid          uint                `gorm:"not null"`
	Foldername        string              `gorm:"not null,default:''"`
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

// TableName TableName IncidenteProcesos Tablename: incidentes_procesos
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

// TableName TableName IncidentesDetalles Tablename: incidentes_detalle
func (IncidentesDetalle) TableName() string {
	return "incidentes_detalle"
}

type Usuario struct {
	gorm.Model
	Nombre         string          `gorm:"not null"`
	Apellido       string          `gorm:"not null"`
	Email          string          `gorm:"not null"`
	Password       string          `gorm:"not null"`
	Roles          []*Rol          `gorm:"many2many:usuarios_roles;"`
	Procesos       []*Proceso      `gorm:"many2many:procesos_usuarios;"`
	Organizaciones []*Organizacion `gorm:"many2many:usuarios_organizaciones;"`
}

func (this *Usuario) GetAdmins(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("roles.Nombre = ?", "admin").Find(&usuarios)
	return usuarios
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
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Where("email = ?", email).First(&this)
}

func (Usuario) TableName() string {
	return "usuarios"
}

func (Usuario) GetAll(db *gorm.DB) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Find(&usuarios)
	return usuarios
}

func (this *Usuario) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").First(&this, id)
}

func (this *Usuario) GetComplete(db *gorm.DB) {
	db.Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Preload("Organizaciones.Procesos").First(&this, this.ID)
}

func (Usuario) GetByProcess(db *gorm.DB, procesoID uint) []*Usuario {
	var usuarios []*Usuario
	db.Preload("Roles").Preload("Procesos").Joins("JOIN procesos_usuarios ON procesos_usuarios.usuario_id = usuarios.id").Where("procesos_usuarios.proceso_id = ?", procesoID).Find(&usuarios)
	return usuarios
}

func (this *Usuario) GetProcesos(db *gorm.DB) []*Proceso {
	var procesos []*Proceso
	_ = db.Model(&this).Association("Procesos").Find(&procesos)
	return procesos
}

type Rol struct {
	gorm.Model
	Nombre   string     `gorm:"not null"`
	Usuarios []*Usuario `gorm:"many2many:usuarios_roles;"`
	Rutas    []*Route   `gorm:"many2many:roles_routes;"`
}

func (Rol) GetAll(db *gorm.DB) []*Rol {
	var roles []*Rol
	db.Preload("Rutas").Find(&roles)
	return roles
}

func (this *Rol) Get(db *gorm.DB, id uint) {
	db.Preload("Rutas").First(&this, id)
}

func (this *Rol) GetUsuarios(db *gorm.DB) {
	err := db.Model(&this).Association("Usuarios").Find(&this.Usuarios)
	if err != nil {
		return
	}
}

func (Rol) TableName() string {
	return "roles"
}

type Route struct {
	gorm.Model
	Method string `gorm:"not null"`
	Route  string `gorm:"not null"`
	Roles  []*Rol `gorm:"many2many:roles_routes;"`
}

func (Route) GetAll(db *gorm.DB) []*Route {
	var routes []*Route
	db.Preload("Roles").Find(&routes)
	return routes
}

func (this *Route) Get(db *gorm.DB, id uint) {
	db.Preload("Roles").First(&this, id)
}
