package Server

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/Mail"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/oxakromax/Backend_UipathMonitor/functions"

	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GetUsers(c echo.Context) error {
	// if query id is not empty, return the user with that id
	id := c.QueryParam("id")
	if id != "" {
		// Convertir el ID de la consulta en un número entero
		ID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid ID")
		}
		// Obtener el usuario de la base de datos
		User := new(ORM.Usuario)
		User.Get(h.DB, uint(ID))
		if User.ID == 0 {
			return c.JSON(http.StatusNotFound, "User not found")
		}
		// Ocultar la contraseña del usuario
		User.Password = ""
		return c.JSON(http.StatusOK, []*ORM.Usuario{User})
	}
	ParamsJson := new(struct {
		Query               string `json:"query" form:"query" query:"query"`
		RelationalCondition string `json:"relational_condition" form:"relational_condition" query:"relational_condition"`
		RelationalQuery     string `json:"relational_query" form:"relational_query" query:"relational_query"`
	})
	// Obtener la consulta de la solicitud
	err := c.Bind(ParamsJson)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid query")
	}
	query := ParamsJson.Query
	Users := make([]*ORM.Usuario, 0)
	if query != "" {
		// Obtener todos los usuarios de la base de datos que coincidan con la consulta
		h.DB.Where(query).Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Find(&Users)
	} else {
		// Obtener todos los usuarios de la base de datos
		Users = new(ORM.Usuario).GetAll(h.DB)
	}
	// Ocultar la contraseña de los usuarios
	for i := range Users {
		Users[i].Password = ""
		for i2 := range Users[i].Organizaciones {
			Users[i].Organizaciones[i2].AppID = ""
			Users[i].Organizaciones[i2].AppSecret = ""
		}
	}
	RelationalCondition := ParamsJson.RelationalCondition
	if RelationalCondition != "" {
		switch RelationalCondition {
		case "NotInOrg":
			// if Not In Org is used, Query must be like: "id NOT IN (?)"
			orgsID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			orgsIDInt := make([]uint, 0)
			for _, orgID := range orgsID {
				ID, err := strconv.Atoi(orgID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				orgsIDInt = append(orgsIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserOrg := range user.Organizaciones {
					for _, u := range orgsIDInt {
						if UserOrg.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if !Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		case "InOrg":
			// if In Org is used, Query must be like: "id IN (?)"
			orgsID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			orgsIDInt := make([]uint, 0)
			for _, orgID := range orgsID {
				ID, err := strconv.Atoi(orgID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				orgsIDInt = append(orgsIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserOrg := range user.Organizaciones {
					for _, u := range orgsIDInt {
						if UserOrg.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		case "NotInRole":
			// if Not In Role is used, Query must be like: "id NOT IN (?)"
			rolesID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			rolesIDInt := make([]uint, 0)
			for _, roleID := range rolesID {
				ID, err := strconv.Atoi(roleID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				rolesIDInt = append(rolesIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserRole := range user.Roles {
					for _, u := range rolesIDInt {
						if UserRole.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if !Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		case "InRole":
			// if In Role is used, Query must be like: "id IN (?)"
			rolesID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			rolesIDInt := make([]uint, 0)
			for _, roleID := range rolesID {
				ID, err := strconv.Atoi(roleID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				rolesIDInt = append(rolesIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserRole := range user.Roles {
					for _, u := range rolesIDInt {
						if UserRole.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		case "NotInProcess":
			// if Not In Process is used, Query must be like: "id NOT IN (?)"
			processesID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			processesIDInt := make([]uint, 0)
			for _, processID := range processesID {
				ID, err := strconv.Atoi(processID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				processesIDInt = append(processesIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserProcess := range user.Procesos {
					for _, u := range processesIDInt {
						if UserProcess.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if !Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		case "InProcess":
			// if In Process is used, Query must be like: "id IN (?)"
			processesID := strings.Split(ParamsJson.RelationalQuery, ",")
			// Convertir los ID de la consulta en números enteros
			processesIDInt := make([]uint, 0)
			for _, processID := range processesID {
				ID, err := strconv.Atoi(processID)
				if err != nil {
					return c.JSON(http.StatusBadRequest, "Invalid ID")
				}
				processesIDInt = append(processesIDInt, uint(ID))
			}
			FinalUsers := make([]*ORM.Usuario, 0)
			for _, user := range Users {
				Found := false
				for _, UserProcess := range user.Procesos {
					for _, u := range processesIDInt {
						if UserProcess.ID == u {
							Found = true
							break
						}
					}
					if Found {
						break
					}
				}
				if Found {
					FinalUsers = append(FinalUsers, user)
				}
			}
			Users = FinalUsers
		}
	}

	// Remueve al usuario monitor de la lista
	for i, user := range Users {
		if user.Email == os.Getenv("MONITOR_USER") {
			Users = append(Users[:i], Users[i+1:]...)
			break
		}
	}

	return c.JSON(http.StatusOK, Users)
}
func (h *Handler) DeleteUser(c echo.Context) error {
	id := c.QueryParam("id")
	// Convertir el ID de la consulta en un número entero
	ID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(h.DB, uint(ID))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Eliminar el usuario de la base de datos
	h.DB.Delete(&User)
	return c.JSON(http.StatusOK, "User deleted")
}
func (h *Handler) CreateUser(c echo.Context) error {
	// Obtener el usuario de la solicitud
	User := new(ORM.Usuario)
	if err := c.Bind(User); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Verificar si el usuario ya existe en la base de datos
	checkUser := new(ORM.Usuario)
	h.DB.Where("email = ?", User.Email).First(&checkUser)
	if checkUser.ID != 0 {
		return c.JSON(http.StatusConflict, "User already exists")
	}
	// Encriptar la contraseña del usuario
	User.Password = functions.GeneratePassword(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(User.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while encrypting password")
	}
	// Asignar el rol de usuario al usuario
	rol := new(ORM.Rol)
	h.DB.Where("nombre = ?", "user").First(&rol)
	User.Roles = append(User.Roles, rol)
	// Enviar la contraseña al correo del usuario
	err = functions.SendMail([]string{User.Email}, "Bienvenido al Monitor de procesos RPA", Mail.GetBodyNewUser(Mail.NewUser{Nombre: User.Nombre, Email: User.Email, Password: User.Password}))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while sending email")
	}
	// Guardar el usuario en la base de datos
	User.Password = string(hash)
	h.DB.Create(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}
func (h *Handler) UpdateUser(c echo.Context) error {
	// Obtener ID desde query
	id := c.QueryParam("id")
	// Convertir el ID de la consulta en un número entero
	ID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}
	// Obtener el usuario de la base de datos
	User := new(ORM.Usuario)
	User.Get(h.DB, uint(ID))
	if User.ID == 0 {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	// Obtener el usuario de la solicitud
	if err := c.Bind(User); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	// Sobre escribir roles del usuario y verificar que aun mantenga el rol 3 "user"
	// Si el usuario no tiene el rol 3 "user" agregarlo
	rol := new(ORM.Rol)
	h.DB.Where("nombre = ?", "user").First(&rol)
	Found := false
	for _, userRol := range User.Roles {
		if userRol.ID == rol.ID {
			Found = true
			break
		}
	}
	if !Found {
		User.Roles = append(User.Roles, rol)
	}
	err = h.DB.Model(&User).Association("Roles").Replace(User.Roles)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error while updating user")
	}
	// Rellenar campos faltantes
	User.FillEmptyFields(h.DB)
	// Guardar el usuario en la base de datos
	h.DB.Save(&User)
	// Ocultar la contraseña del usuario
	User.Password = ""
	return c.JSON(http.StatusOK, User)
}

func (h *Handler) GetAllRoles(c echo.Context) error {
	// Get user from token
	User, err := h.GetUserJWT(c)
	if err != nil {
		return err
	}
	// Obtener los roles de la base de datos
	Roles := new([]*ORM.Rol)
	h.DB.Order("Nombre").Find(&Roles)
	// Si el usuario no es admin eliminar los roles de admin
	// además siempre eliminar el rol de monitor
	// para verificar usa User.HasRole("admin")

	FinalRoles := make([]*ORM.Rol, 0)
	for _, role := range *Roles {
		if role.Nombre != "admin" && role.Nombre != "monitor" && role.Nombre != "user" {
			FinalRoles = append(FinalRoles, role)
		}
		if role.Nombre == "admin" && User.HasRole("admin") {
			FinalRoles = append(FinalRoles, role)
		}
	}
	Roles = &FinalRoles

	return c.JSON(http.StatusOK, Roles)
}
