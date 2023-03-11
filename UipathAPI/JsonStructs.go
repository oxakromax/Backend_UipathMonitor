package UipathAPI

import "time"

// LogResponse es una estructura que contiene una lista de registros de log.
// Los campos OdataContext y OdataCount contienen información adicional sobre la respuesta.
// El campo Value contiene la lista de registros de log.
// Log es una estructura que representa un registro de log en el sistema.
// Los campos Level, WindowsIdentity, ProcessName, TimeStamp, Message, JobKey, RawMessage,
// RobotName, HostMachineName, MachineId y RuntimeType contienen información detallada sobre el log.
// El campo Id contiene un identificador único para el log.
type LogResponse struct {
	OdataContext string `json:"@odata.context"`
	OdataCount   int    `json:"@odata.count"`
	Value        []struct {
		Level           string      `json:"Level"`
		WindowsIdentity string      `json:"WindowsIdentity"`
		ProcessName     string      `json:"ProcessName"`
		TimeStamp       time.Time   `json:"TimeStamp"`
		Message         string      `json:"Message"`
		JobKey          string      `json:"JobKey"`
		RawMessage      string      `json:"RawMessage"`
		RobotName       string      `json:"RobotName"`
		HostMachineName string      `json:"HostMachineName"`
		MachineId       int         `json:"MachineId"`
		RuntimeType     interface{} `json:"RuntimeType"`
		Id              int         `json:"Id"`
	} `json:"value"`
}

// ReleasesResponse es una estructura que representa la respuesta de una petición a la API de UiPath para obtener información de los procesos.
// La respuesta contiene el contexto OData, el número de elementos incluidos en la respuesta, y un arreglo con la información de cada proceso.
type ReleasesResponse struct {
	OdataContext string `json:"@odata.context"`
	OdataCount   int    `json:"@odata.count"`
	Value        []struct {
		Key              string      `json:"TokenKey"`
		ProcessKey       string      `json:"ProcessKey"`
		ProcessVersion   string      `json:"ProcessVersion"`
		IsLatestVersion  bool        `json:"IsLatestVersion"`
		IsProcessDeleted bool        `json:"IsProcessDeleted"`
		Description      interface{} `json:"Description"`
		Name             string      `json:"Name"`
		EnvironmentId    int         `json:"EnvironmentId"`
		EnvironmentName  string      `json:"EnvironmentName"`
		InputArguments   interface{} `json:"InputArguments"`
		Id               int         `json:"Id"`
		Arguments        interface{} `json:"Arguments"`
	} `json:"value"`
}

type FoldersResponse struct {
	OdataContext string `json:"@odata.context"`
	OdataCount   int    `json:"@odata.count"`
	Value        []struct {
		Key                         string      `json:"TokenKey"`
		DisplayName                 string      `json:"DisplayName"`
		FullyQualifiedName          string      `json:"FullyQualifiedName"`
		FullyQualifiedNameOrderable string      `json:"FullyQualifiedNameOrderable"`
		Description                 interface{} `json:"Description"`
		ProvisionType               string      `json:"ProvisionType"`
		PermissionModel             string      `json:"PermissionModel"`
		ParentId                    interface{} `json:"ParentId"`
		ParentKey                   interface{} `json:"ParentKey"`
		IsActive                    bool        `json:"IsActive"`
		FeedType                    string      `json:"FeedType"`
		Id                          int         `json:"Id"`
	} `json:"value"`
}

type JobsResponse struct {
	OdataContext string `json:"@odata.context"`
	OdataCount   int    `json:"@odata.count"`
	Value        []struct {
		Key                string      `json:"TokenKey"`
		StartTime          time.Time   `json:"StartTime"`
		EndTime            time.Time   `json:"EndTime"`
		State              string      `json:"State"`
		Source             string      `json:"Source"`
		BatchExecutionKey  string      `json:"BatchExecutionKey"`
		Info               string      `json:"Info"`
		CreationTime       time.Time   `json:"CreationTime"`
		StartingScheduleId interface{} `json:"StartingScheduleId"`
		Id                 int         `json:"Id"`
	} `json:"value"`
}
