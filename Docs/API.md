Documentación de la función `GetUsers`
======================================

Descripción
-----------

La función `GetUsers` es un controlador para la API que permite obtener usuarios de la base de datos. Puede realizar
consultas específicas y filtrar usuarios según diferentes criterios, como su pertenencia a una organización, rol o
proceso.

Parámetros
----------

La función acepta un JSON con los siguientes campos opcionales:

* `query`: Una consulta de filtrado que se aplica a la base de datos de usuarios. Ejemplo: `"name LIKE 'John%'"`.
* `relational_condition`: Una cadena que especifica la condición relacional que se aplicará al filtrado. Valores
  posibles: `"NotInOrg"`, `"InOrg"`, `"NotInRole"`, `"InRole"`, `"NotInProcess"`, `"InProcess"`.
* `relational_query`: Una cadena que contiene una lista de ID de elementos (organizaciones, roles o procesos) separados
  por comas, que se utilizarán en la condición relacional. Ejemplo: `"1,2,3"`.

Ejemplo de uso
--------------

Realizar una solicitud HTTP GET a la API con el siguiente formato:

```json
{
  "query": "name LIKE 'John%'",
  "relational_condition": "NotInOrg",
  "relational_query": "1,2,3"
}
```

Esto devuelve todos los usuarios cuyo nombre comienza con "John" y que no están en las organizaciones con ID 1, 2 y 3.

Código
------

A continuación se presenta el código completo de la función `GetUsers`:

```go
func (H *Handler) GetUsers(c echo.Context) error {
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
		User.Get(H.Db, uint(ID))
		if User.ID == 0 {
			return c.JSON(http.StatusNotFound, "User not found")
		}
		// Ocultar la contraseña del usuario
		User.Password = ""
		return c.JSON(http.StatusOK, User)
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
		H.Db.Where(query).Preload("Roles").Preload("Procesos").Preload("Roles.Rutas").Preload("Organizaciones").Find(&Users)
	} else {
		// Obtener todos los usuarios de la base de datos
		Users = new(ORM.Usuario).GetAll(H.Db)
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

	return c.JSON(http.StatusOK, Users)
}
```

Detalles de implementación
--------------------------

La función realiza las siguientes acciones:

1. Verifica si se proporciona el parámetro "id" en la URL. Si es así, busca al usuario con ese ID y lo devuelve como
   resultado.
2. Analiza los campos opcionales del JSON proporcionado (`query`, `relational_condition`, `relational_query`) y los
   utiliza para realizar consultas y filtrar usuarios en la base de datos.
3. Si se proporciona un valor para `query`, realiza una consulta en la base de datos de usuarios utilizando ese valor.
4. Aplica la condición relacional (`relational_condition`) y la consulta relacional (`relational_query`) si se
   proporcionan.
5. Devuelve el resultado en formato JSON.