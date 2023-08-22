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
	OdataContext string     `json:"@odata.context"`
	OdataCount   int        `json:"@odata.count"`
	Value        []LogValue `json:"value"`
}
type LogValue struct {
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
	MachineKey      string      `json:"MachineKey"`
	RuntimeType     interface{} `json:"RuntimeType"`
	ID              int         `json:"Id"`
}

// ReleasesResponse es una estructura que representa la respuesta de una petición a la API de UiPath para obtener información de los procesos.
// La respuesta contiene el contexto OData, el número de elementos incluidos en la respuesta, y un arreglo con la información de cada proceso.

type ReleasesResponse struct {
	OdataContext string          `json:"@odata.context"`
	OdataCount   int             `json:"@odata.count"`
	Value        []ReleasesValue `json:"value"`
}
type ReleasesValue struct {
	Key                                string      `json:"Key"`
	ProcessKey                         string      `json:"ProcessKey"`
	ProcessVersion                     string      `json:"ProcessVersion"`
	IsLatestVersion                    bool        `json:"IsLatestVersion"`
	IsProcessDeleted                   bool        `json:"IsProcessDeleted"`
	Description                        string      `json:"Description"`
	Name                               string      `json:"Name"`
	EnvironmentId                      interface{} `json:"EnvironmentId"`
	EnvironmentName                    string      `json:"EnvironmentName"`
	EntryPointId                       int         `json:"EntryPointId"`
	InputArguments                     interface{} `json:"InputArguments"`
	ProcessType                        string      `json:"ProcessType"`
	SupportsMultipleEntryPoints        bool        `json:"SupportsMultipleEntryPoints"`
	RequiresUserInteraction            bool        `json:"RequiresUserInteraction"`
	IsAttended                         bool        `json:"IsAttended"`
	IsCompiled                         bool        `json:"IsCompiled"`
	AutomationHubIdeaUrl               interface{} `json:"AutomationHubIdeaUrl"`
	AutoUpdate                         bool        `json:"AutoUpdate"`
	FeedId                             string      `json:"FeedId"`
	JobPriority                        string      `json:"JobPriority"`
	SpecificPriorityValue              int         `json:"SpecificPriorityValue"`
	OrganizationUnitId                 int         `json:"OrganizationUnitId"`
	OrganizationUnitFullyQualifiedName string      `json:"OrganizationUnitFullyQualifiedName"`
	TargetFramework                    string      `json:"TargetFramework"`
	RobotSize                          interface{} `json:"RobotSize"`
	AutoCreateConnectedTriggers        bool        `json:"AutoCreateConnectedTriggers"`
	RemoteControlAccess                string      `json:"RemoteControlAccess"`
	LastModificationTime               time.Time   `json:"LastModificationTime"`
	LastModifierUserId                 int         `json:"LastModifierUserId"`
	CreationTime                       time.Time   `json:"CreationTime"`
	CreatorUserId                      int         `json:"CreatorUserId"`
	ID                                 int         `json:"Id"`
	Arguments                          struct {
		Input  string `json:"Input"`
		Output string `json:"Output"`
	} `json:"Arguments"`
	ProcessSettings        interface{}   `json:"ProcessSettings"`
	VideoRecordingSettings interface{}   `json:"VideoRecordingSettings"`
	Tags                   []interface{} `json:"Tags"`
	ResourceOverwrites     []interface{} `json:"ResourceOverwrites"`
}

type FoldersResponse struct {
	OdataContext string         `json:"@odata.context"`
	OdataCount   int            `json:"@odata.count"`
	Value        []FoldersValue `json:"value"`
}
type FoldersValue struct {
	Key                         string      `json:"Key"`
	DisplayName                 string      `json:"DisplayName"`
	FullyQualifiedName          string      `json:"FullyQualifiedName"`
	FullyQualifiedNameOrderable string      `json:"FullyQualifiedNameOrderable"`
	Description                 interface{} `json:"Description"`
	FolderType                  string      `json:"FolderType"`
	ProvisionType               string      `json:"ProvisionType"`
	PermissionModel             string      `json:"PermissionModel"`
	ParentID                    interface{} `json:"ParentId"`
	ParentKey                   interface{} `json:"ParentKey"`
	IsActive                    bool        `json:"IsActive"`
	FeedType                    string      `json:"FeedType"`
	ID                          int         `json:"Id"`
}

type JobsResponse struct {
	OdataContext string      `json:"@odata.context"`
	OdataCount   int         `json:"@odata.count"`
	Value        []JobsValue `json:"value"`
}
type JobsValue struct {
	Key                                string      `json:"Key"`
	StartTime                          time.Time   `json:"StartTime"`
	EndTime                            time.Time   `json:"EndTime"`
	State                              string      `json:"State"`
	JobPriority                        string      `json:"JobPriority"`
	SpecificPriorityValue              *int        `json:"SpecificPriorityValue"`
	ResourceOverwrites                 interface{} `json:"ResourceOverwrites"`
	Source                             string      `json:"Source"`
	SourceType                         string      `json:"SourceType"`
	BatchExecutionKey                  string      `json:"BatchExecutionKey"`
	Info                               *string     `json:"Info"`
	CreationTime                       time.Time   `json:"CreationTime"`
	StartingScheduleId                 *int        `json:"StartingScheduleId"`
	ReleaseName                        string      `json:"ReleaseName"`
	Type                               string      `json:"Type"`
	InputArguments                     *string     `json:"InputArguments"`
	OutputArguments                    *string     `json:"OutputArguments"`
	HostMachineName                    *string     `json:"HostMachineName"`
	HasMediaRecorded                   bool        `json:"HasMediaRecorded"`
	HasVideoRecorded                   bool        `json:"HasVideoRecorded"`
	PersistenceId                      interface{} `json:"PersistenceId"`
	ResumeVersion                      interface{} `json:"ResumeVersion"`
	StopStrategy                       *string     `json:"StopStrategy"`
	RuntimeType                        string      `json:"RuntimeType"`
	RequiresUserInteraction            bool        `json:"RequiresUserInteraction"`
	ReleaseVersionId                   *int        `json:"ReleaseVersionId"`
	EntryPointPath                     string      `json:"EntryPointPath"`
	OrganizationUnitId                 int         `json:"OrganizationUnitId"`
	OrganizationUnitFullyQualifiedName string      `json:"OrganizationUnitFullyQualifiedName"`
	Reference                          string      `json:"Reference"`
	ProcessType                        string      `json:"ProcessType"`
	ProfilingOptions                   interface{} `json:"ProfilingOptions"`
	ResumeOnSameContext                bool        `json:"ResumeOnSameContext"`
	LocalSystemAccount                 string      `json:"LocalSystemAccount"`
	OrchestratorUserIdentity           interface{} `json:"OrchestratorUserIdentity"`
	RemoteControlAccess                string      `json:"RemoteControlAccess"`
	MaxExpectedRunningTimeSeconds      *int        `json:"MaxExpectedRunningTimeSeconds"`
	ServerlessJobType                  interface{} `json:"ServerlessJobType"`
	ID                                 int         `json:"Id"`
}

type ProcessSchedulesValue struct {
	Key                             string      `json:"Key"`
	Enabled                         bool        `json:"Enabled"`
	Name                            string      `json:"Name"`
	ReleaseId                       int         `json:"ReleaseId"`
	ReleaseKey                      string      `json:"ReleaseKey"`
	ReleaseName                     string      `json:"ReleaseName"`
	PackageName                     string      `json:"PackageName"`
	EnvironmentName                 interface{} `json:"EnvironmentName"`
	EnvironmentId                   string      `json:"EnvironmentId"`
	JobPriority                     interface{} `json:"JobPriority"`
	SpecificPriorityValue           interface{} `json:"SpecificPriorityValue"`
	RuntimeType                     string      `json:"RuntimeType"`
	StartProcessCron                string      `json:"StartProcessCron"`
	StartProcessCronDetails         string      `json:"StartProcessCronDetails"`
	StartProcessCronSummary         string      `json:"StartProcessCronSummary"`
	StartProcessNextOccurrence      time.Time   `json:"StartProcessNextOccurrence"`
	StartStrategy                   int         `json:"StartStrategy"`
	StopProcessExpression           string      `json:"StopProcessExpression"`
	StopStrategy                    interface{} `json:"StopStrategy"`
	KillProcessExpression           interface{} `json:"KillProcessExpression"`
	ExternalJobKey                  interface{} `json:"ExternalJobKey"`
	ExternalJobKeyScheduler         string      `json:"ExternalJobKeyScheduler"`
	TimeZoneId                      string      `json:"TimeZoneId"`
	TimeZoneIana                    string      `json:"TimeZoneIana"`
	UseCalendar                     bool        `json:"UseCalendar"`
	CalendarId                      interface{} `json:"CalendarId"`
	CalendarName                    interface{} `json:"CalendarName"`
	StopProcessDate                 interface{} `json:"StopProcessDate"`
	InputArguments                  interface{} `json:"InputArguments"`
	QueueDefinitionId               interface{} `json:"QueueDefinitionId"`
	QueueDefinitionName             interface{} `json:"QueueDefinitionName"`
	ItemsActivationThreshold        int         `json:"ItemsActivationThreshold"`
	ItemsPerJobActivationTarget     int         `json:"ItemsPerJobActivationTarget"`
	MaxJobsForActivation            int         `json:"MaxJobsForActivation"`
	ResumeOnSameContext             bool        `json:"ResumeOnSameContext"`
	Description                     interface{} `json:"Description"`
	AlertPendingExpression          interface{} `json:"AlertPendingExpression"`
	AlertRunningExpression          interface{} `json:"AlertRunningExpression"`
	RunAsMe                         bool        `json:"RunAsMe"`
	ConsecutiveJobFailuresThreshold int         `json:"ConsecutiveJobFailuresThreshold"`
	JobFailuresGracePeriodInHours   int         `json:"JobFailuresGracePeriodInHours"`
	IsConnected                     bool        `json:"IsConnected"`
	Id                              int         `json:"Id"`
	MachineRobots                   []struct {
		MachineId     int         `json:"MachineId"`
		MachineName   string      `json:"MachineName"`
		RobotId       int         `json:"RobotId"`
		RobotUserName interface{} `json:"RobotUserName"`
		SessionId     int         `json:"SessionId"`
		SessionName   interface{} `json:"SessionName"`
	} `json:"MachineRobots"`
	Tags []interface{} `json:"Tags"`
}

type ProcessSchedulesResponse struct {
	OdataContext string                  `json:"@odata.context"`
	OdataCount   int                     `json:"@odata.count"`
	Value        []ProcessSchedulesValue `json:"value"`
}
