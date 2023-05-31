package ORM

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/google/go-querystring/query"
	"github.com/oxakromax/Backend_UipathMonitor/UipathAPI"
	"github.com/oxakromax/Backend_UipathMonitor/functions"
	"gorm.io/gorm"
)

var (
	UipathBearerTokenMap sync.Map // sync map because there's a lot of goroutines reading and writing to it
)

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

func (Organizacion) TableName() string {
	return "organizaciones"
}

func (o *Organizacion) RefreshUiPathToken() error {
	var url = o.BaseURL + "identity_/connect/token"
	const method = "POST"
	ClientID, _ := functions.DecryptAES(os.Getenv("DB_KEY"), o.AppID)
	ClientSecret, _ := functions.DecryptAES(os.Getenv("DB_KEY"), o.AppSecret)
	var QueryAuth = struct {
		GrantType    string `json:"grant_type" url:"grant_type"`
		ClientId     string `json:"client_id" url:"client_id"`
		ClientSecret string `json:"client_secret" url:"client_secret"`
		Scope        string `json:"scope" url:"scope"`
	}{
		GrantType:    "client_credentials",
		ClientId:     ClientID,
		ClientSecret: ClientSecret,
		Scope:        o.AppScope,
	}
	vals, err := query.Values(QueryAuth)
	if err != nil {
		return err
	}
	payload := bytes.NewBufferString(vals.Encode())
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
	UipathBearerTokenMap.Store(o.ID, UiPathToken.AccessToken)
	return res.Body.Close()
}

func (o *Organizacion) GetUrl() string {
	return o.BaseURL + o.Uipathname + "/" + o.Tenantname + "/orchestrator_/odata/"
}

func (o *Organizacion) RequestAPI(method, path string, body io.Reader, conds ...interface{}) (*http.Response, error) {
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
		req, err := http.NewRequest(method, o.GetUrl()+path, body)
		if err != nil {
			return nil, err
		}
		var UipathBearerToken string
		if val, ok := UipathBearerTokenMap.Load(o.ID); ok {
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
			err = o.RefreshUiPathToken()
			if err != nil {
				return nil, err
			}
			continue
		}
		return res, nil
	}
	return nil, errors.New("error al obtener el token de UiPath")
}

func (o *Organizacion) GetFromApi(structType interface{}, folderid ...int) error {
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
		resp, err = o.RequestAPI("GET", "Folders", nil)
	case *UipathAPI.LogResponse:
		resp, err = o.RequestAPI("GET", "RobotLogs", nil, folderID)
	case *UipathAPI.ReleasesResponse:
		resp, err = o.RequestAPI("GET", "Releases", nil, folderID)
	case *UipathAPI.JobsResponse:
		resp, err = o.RequestAPI("GET", "Jobs", nil, folderID)
	case *UipathAPI.ProcessSchedulesResponse:
		resp, err = o.RequestAPI("GET", "ProcessSchedules", nil, folderID)
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

func (o *Organizacion) CheckAccessAPI() error {
	err := o.RefreshUiPathToken()
	if err != nil {
		return err
	}
	_, err = o.RequestAPI("GET", "Folders", nil)
	return err
}

func (o *Organizacion) GetAll(db *gorm.DB) []*Organizacion {
	var organizaciones []*Organizacion
	db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").Find(&organizaciones)
	return organizaciones
}

func (o *Organizacion) Get(db *gorm.DB, id uint) {
	db.Preload("Procesos").Preload("Clientes").Preload("Usuarios").First(&o, id)
}
