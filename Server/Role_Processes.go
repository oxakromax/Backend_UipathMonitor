package Server

import (
	"sort"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
)

func (H *Handler) GetProcesses(c echo.Context) error {
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if User.HasRole("processes_administration") {
		Process := new(ORM.Proceso).GetAll(H.Db)
		return c.JSON(200, Process)
	}
	return c.JSON(200, User.Procesos)
}

func (H *Handler) GetProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	Process := new(ORM.Proceso)
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if User.HasRole("processes_administration") || User.HasProcess(ProcessID) {
		Process.Get(H.Db, uint(ProcessID))
		sort.Slice(Process.IncidentesProceso, func(i, j int) bool {
			return Process.IncidentesProceso[i].ID > Process.IncidentesProceso[j].ID
		})
		return c.JSON(200, Process)
	}

	return c.JSON(404, "Process not found")

}

// GetPossibleUsers returns all users that can be added to a process, excluding the users that are already in the process
func (H *Handler) GetPossibleUsers(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	Process := new(ORM.Proceso)
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if User.HasRole("processes_administration") || User.HasProcess(ProcessID) {
		Process.Get(H.Db, uint(ProcessID))
		Org := new(ORM.Organizacion)
		Org.Get(H.Db, Process.OrganizacionID)
		UsersReturn := []*ORM.Usuario{}
		for _, OrgUser := range Org.Usuarios {
			// check if the user is already in the process
			IsInProcess := false
			for _, ProcessUser := range Process.Usuarios {
				if OrgUser.ID == ProcessUser.ID {
					IsInProcess = true
				}
			}
			if !IsInProcess {
				UsersReturn = append(UsersReturn, OrgUser)
			}
		}
		return c.JSON(200, UsersReturn)
	}
	return c.JSON(404, "Process not found")
}

// GetPossibleClients returns all clients that can be added to a process, excluding the clients that are already in the process
func (H *Handler) GetPossibleClients(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	Process := new(ORM.Proceso)
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if User.HasRole("processes_administration") || User.HasProcess(ProcessID) {
		Process.Get(H.Db, uint(ProcessID))
		Org := new(ORM.Organizacion)
		Org.Get(H.Db, Process.OrganizacionID)
		ClientsReturn := []*ORM.Cliente{}
		for _, OrgClient := range Org.Clientes {
			// check if the client is already in the process
			IsInProcess := false
			for _, ProcessClient := range Process.Clientes {
				if OrgClient.ID == ProcessClient.ID {
					IsInProcess = true
				}
			}
			if !IsInProcess {
				ClientsReturn = append(ClientsReturn, OrgClient)
			}
		}
		return c.JSON(200, ClientsReturn)
	}
	return c.JSON(404, "Process not found")
}

// Update process (and check if the user has the role "processes_administration" or if the user is the owner of the process)
func (H *Handler) UpdateProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	Process := new(ORM.Proceso)
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	if User.HasRole("processes_administration") || User.HasProcess(ProcessID) {
		Process.Get(H.Db, uint(ProcessID))
		if Process.ID == 0 {
			return c.JSON(404, "Process not found")
		}
		if err := c.Bind(Process); err != nil {
			return c.JSON(400, "Invalid JSON")
		}
		H.Db.Save(Process)
		return c.JSON(200, Process)
	}
	return c.JSON(403, "Forbidden")
}

// Remove clients from process
func (H *Handler) DeleteClientsFromProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	// check if the user had "processes_administration" role Or if the user is the owner of the process
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the clients to remove from query params
	ClientIDStr := c.QueryParam("clients_id") // comma separated list of clients id, like: 1,2,3,4
	ClientList := []int{}
	for _, ClientID := range strings.Split(ClientIDStr, ",") {
		ClientIDInt, err := strconv.Atoi(ClientID)
		if err != nil {
			return c.JSON(400, "Invalid client ID:"+ClientID)
		}
		ClientList = append(ClientList, ClientIDInt)
	}
	// Remove the clients from the process
	err = Process.RemoveClients(H.Db, ClientList)
	if err != nil {
		return c.JSON(500, "Error removing clients from process, error:"+err.Error())
	}
	return c.JSON(200, "Clients removed from process")
}

// Add clients to process
func (H *Handler) AddClientsToProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	// check if the user had "processes_administration" role Or if the user is the owner of the process
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the clients to add from query params
	ClientIDStr := c.QueryParam("clients_id") // comma separated list of clients id, like: 1,2,3,4
	ClientList := []int{}
	for _, ClientID := range strings.Split(ClientIDStr, ",") {
		ClientIDInt, err := strconv.Atoi(ClientID)
		if err != nil {
			return c.JSON(400, "Invalid client ID:"+ClientID)
		}
		ClientList = append(ClientList, ClientIDInt)
	}
	// Add the clients to the process
	err = Process.AddClients(H.Db, ClientList)
	if err != nil {
		return c.JSON(500, "Error adding clients to process, error:"+err.Error())
	}
	return c.JSON(200, "Clients added to process")
}

// Remove users from process
func (H *Handler) DeleteUsersFromProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	// check if the user had "processes_administration" role Or if the user is the owner of the process
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the users to remove from query params
	UserIDStr := c.QueryParam("users_id") // comma separated list of users id, like: 1,2,3,4
	UserList := []int{}
	for _, UserID := range strings.Split(UserIDStr, ",") {
		UserIDInt, err := strconv.Atoi(UserID)
		if err != nil {
			return c.JSON(400, "Invalid user ID:"+UserID)
		}
		UserList = append(UserList, UserIDInt)
	}
	// Remove the users from the process
	err = Process.RemoveUsers(H.Db, UserList)
	if err != nil {
		return c.JSON(500, "Error removing users from process, error:"+err.Error())
	}
	return c.JSON(200, "Users removed from process")
}

// Add users to process, and if the user doesn't is in the organization, add it
func (H *Handler) AddUsersToProcess(c echo.Context) error {
	ProcessIDStr := c.Param("id")
	ProcessID, err := strconv.Atoi(ProcessIDStr)
	if err != nil {
		return c.JSON(400, "Invalid process ID")
	}
	// check if the user had "processes_administration" role Or if the user is the owner of the process
	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)
	HasProcess := User.HasProcess(ProcessID)
	HasRole := User.HasRole("processes_administration")
	if !HasProcess && !HasRole {
		return c.JSON(403, "Forbidden")
	}
	var Process ORM.Proceso

	Process.Get(H.Db, uint(ProcessID))
	if Process.ID == 0 {
		return c.JSON(404, "Process not found")
	}
	// Get the users to add from query params
	UserIDStr := c.QueryParam("users_id") // comma separated list of users id, like: 1,2,3,4
	UserList := []int{}
	for _, UserID := range strings.Split(UserIDStr, ",") {
		UserIDInt, err := strconv.Atoi(UserID)
		if err != nil {
			return c.JSON(400, "Invalid user ID:"+UserID)
		}
		UserList = append(UserList, UserIDInt)
	}
	// Add the users to the process
	err = Process.AddUsers(H.Db, UserList)
	if err != nil {
		return c.JSON(500, "Error adding users to process, error:"+err.Error())
	}
	return c.JSON(200, "Users added to process")
}
