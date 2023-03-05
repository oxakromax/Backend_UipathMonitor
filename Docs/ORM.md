# Inicialización

```go
func OpenDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=Proyecto port=5432 sslmode=disable"
	log := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log,
	})
	log.Info(nil, "Database connection successfully opened")
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&ORM.Organizacion{}, &ORM.Cliente{}, &ORM.Proceso{}, &ORM.IncidenteProceso{}, &ORM.IncidentesDetalle{}, &ORM.Usuario{}, &ORM.Rol{})
	if err != nil {
		panic("failed to migrate database")
	}
	return db
}
```

La función `OpenDB` se utiliza para conectarse a la base de datos del proyecto y devolver un objeto de base de datos de tipo *gorm.DB. Esta función utiliza los siguientes parámetros:

`dsn`: Una cadena de conexión de base de datos PostgreSQL que incluye información de autenticación y otros detalles de conexión.

`log`: Un objeto logger utilizado para registrar información sobre la conexión a la base de datos.

La función devuelve un objeto *gorm.DB que se puede utilizar para realizar operaciones de base de datos en el proyecto.

## Estructura: Organización

La estructura `Organización` tiene los siguientes campos:

- `ID`: campo predeterminado de GORM para la clave primaria del registro
- `CreatedAt`: campo predeterminado de GORM que almacena la fecha de creación del registro
- `UpdatedAt`: campo predeterminado de GORM que almacena la fecha de actualización del registro
- `DeletedAt`: campo predeterminado de GORM que se utiliza para suavizar la eliminación de registros
- `Nombre`: campo que almacena el nombre de la organización
- `Uipathname`: campo que almacena la dirección URL de la organización en la plataforma UiPath
- `Tenantname`: campo que almacena el nombre del inquilino de la organización en la plataforma UiPath
- `AppID`: campo que almacena el ID de la aplicación de la organización en la plataforma UiPath
- `AppSecret`: campo que almacena el secreto de la aplicación de la organización en la plataforma UiPath
- `Clientes`: relación uno a muchos con la tabla `clientes`
- `Procesos`: relación uno a muchos con la tabla `procesos`

```json
{
  "ID": 1,
  "CreatedAt": "2021-10-05T18:00:00Z",
  "UpdatedAt": "2021-10-05T18:00:00Z",
  "DeletedAt": null,
  "Nombre": "Organizacion de Ejemplo",
  "Uipathname": "https://cloud.uipath.com/org-example",
  "Tenantname": "org-example",
  "AppID": "********",
  "AppSecret": "********",
  "Clientes": [...],
  "Procesos": [...]
}
```

### GetAll

La función `GetAll` se utiliza para recuperar todas las organizaciones de la base de datos. Toma un objeto `db` de GORM como argumento y devuelve un slice de punteros a la estructura `Organización`.

```go
func (Organización) GetAll(db *gorm.DB) []*Organización 
```
La función utiliza la función `Preload` de GORM para precargar las relaciones de `Procesos` y `Clientes` de cada registro de
Organización. Luego, utiliza la función `Find` de GORM para recuperar todas las organizaciones y almacenarlas en un slice
de punteros a la estructura `Organización`, que se devuelve como resultado.

### Get

La función Get recibe un objeto `db` de tipo `*gorm.DB` y un identificador `id` de tipo `uint`. La función actualiza la
estructura `Organización` que se llama en `this` con los datos correspondientes de la organización que tiene el
identificador especificado.
    
```go
func (this *Organización) Get(db *gorm.DB, id uint) 
``` 

### TableName

La función `TableName` se utiliza para especificar el nombre de la tabla de la base de datos a la que se hace referencia en la estructura `Organización`.

```go
func (Organización) TableName() string {
    return "organizaciones"
}
```

En este caso, devuelve la cadena `"organizaciones"`, que es el nombre de la tabla correspondiente en la base de datos.

## Estructura: Cliente

La estructura `Cliente` representa un cliente de una organización y tiene los siguientes campos:

- `ID`: identificador único del cliente.
- `Nombre`: nombre del cliente (no puede ser nulo).
- `Apellido`: apellido del cliente (no puede ser nulo).
- `Email`: dirección de correo electrónico del cliente (no puede ser nulo).
- `OrganizacionID`: identificador único de la organización a la que pertenece el cliente (no puede ser nulo).
- `Organizacion`: organización a la que pertenece el cliente.
- `Procesos`: lista de procesos en los que participa el cliente.

```json
{
  "ID": 1,
  "Nombre": "John",
  "Apellido": "Doe",
  "Email": "johndoe@example.com",
  "OrganizacionID": 1,
  "Organizacion": {
    "ID": 1,
    "Nombre": "Mi Organizacion",
    "Uipathname": "mi_organizacion",
    "Tenantname": "mi_tenant",
    "AppID": "app_id",
    "AppSecret": "app_secret",
    "Clientes": [...],
    "Procesos": [...]
  },
  "Procesos": [...]
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (Cliente) GetAll(db *gorm.DB) []*Cliente
```

Este método devuelve todos los clientes registrados en la base de datos, incluyendo la organización y los procesos a los que pertenecen.

### Get

```go
func (this *Cliente) Get(db *gorm.DB, id uint)
```

Este método devuelve el cliente con el identificador `id`, incluyendo la organización y los procesos a los que pertenece.

### GetByProcess

```go
func (Cliente) GetByProcess(db *gorm.DB, id uint) []*Cliente
```

Este método devuelve todos los clientes que participan en el proceso con el identificador `id`, incluyendo la organización a la que pertenecen y los demás procesos en los que participan.

### Función TableName

```go
func (Cliente) TableName() string
```

Esta función devuelve el nombre de la tabla de la base de datos correspondiente a la estructura `Cliente`.

## Estructura Proceso

La estructura `Proceso` representa un proceso de Uipath y tiene los siguientes campos:

- `ID`: identificador único del proceso.
- `Nombre`: nombre del proceso (no puede ser nulo).
- `Folderid`: identificador único de la carpeta en la que se encuentra el proceso en Uipath (no puede ser nulo).
- `OrganizacionID`: identificador único de la organización a la que pertenece el proceso (no puede ser nulo).
- `WarningTolerance`: tolerancia a las advertencias en el proceso (no puede ser nulo, valor por defecto: 10).
- `ErrorTolerance`: tolerancia a los errores en el proceso (no puede ser nulo, valor por defecto: 0).
- `FatalTolerance`: tolerancia a los errores fatales en el proceso (no puede ser nulo, valor por defecto: 0).
- `Organizacion`: organización a la que pertenece el proceso.
- `IncidentesProceso`: lista de incidentes relacionados con el proceso.
- `Clientes`: lista de clientes que participan en el proceso.
- `Usuarios`: lista de usuarios asignados al proceso.

```json
{
  "id": 1,
  "nombre": "Proceso de ejemplo",
  "folderid": 1234,
  "warning_tolerance": 10,
  "error_tolerance": 0,
  "fatal_tolerance": 0,
  "organizacion_id": 1,
  "organizacion": {
    "id": 1,
    "nombre": "Organización de ejemplo",
    "uipathname": "https://cloud.uipath.com/",
    "tenantname": "tenant_name",
    "app_id": "app_id_example",
    "app_secret": "app_secret_example",
    "clientes": [...],
    "procesos": [...]
  },
  "incidentes_proceso": [...],
  "clientes": [...],
  "usuarios": [...]
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (Proceso) GetAll(db *gorm.DB) []*Proceso
```

Este método devuelve todos los procesos registrados en la base de datos, incluyendo la organización, los incidentes relacionados, los clientes y usuarios asignados.

### Get

```go
func (this *Proceso) Get(db *gorm.DB, id uint)
```

Este método devuelve el proceso con el identificador `id`, incluyendo la organización, los incidentes relacionados, los clientes y usuarios asignados.

### GetByOrganizacion

```go
func (Proceso) GetByOrganizacion(db *gorm.DB, organizacionID uint) []*Proceso
```

Este método devuelve todos los procesos pertenecientes a la organización con el identificador `organizacionID`, incluyendo los incidentes relacionados, los clientes y usuarios asignados.

### GetByFolder

```go
func (Proceso) GetByFolder(db *gorm.DB, folderID uint) []*Proceso
```

Este método devuelve todos los procesos que se encuentran en la carpeta con el identificador `folderID`, incluyendo la organización, los incidentes relacionados, los clientes y usuarios asignados.

### GetEmails

```go
func (this *Proceso) GetEmails() []string
```

Este método devuelve una lista con las direcciones de correo electrónico de todos los clientes y usuarios asignados al proceso.

## Estructura IncidenteProceso

La estructura `IncidenteProceso` representa un incidente relacionado con un proceso y tiene los siguientes campos:

- `ID`: identificador único del incidente.
- `ProcesoID`: identificador único del proceso al que pertenece el incidente (no puede ser nulo).
- `Proceso`: proceso al que pertenece el incidente.
- `Incidente`: descripción del incidente (puede ser nulo).
- `Tipo`: tipo de incidente (no puede ser nulo, valor por defecto: 1).
- `Estado`: estado del incidente (no puede ser nulo, valor por defecto: 1).
- `Detalles`: lista de detalles relacionados con el incidente.

```json
{
  "ID": 1,
  "ProcesoID": 1,
  "Proceso": {...},
  "Incidente": "Error en la lectura de datos",
  "Tipo": 1,
  "Estado": 1,
  "Detalles": [
    {
      "ID": 1,
      "IncidenteID": 1,
      "Detalle": "Se produjo un error al intentar leer el archivo de datos",
      "FechaInicio": "2021-10-01T10:00:00Z",
      "FechaFin": "2021-10-01T10:05:00Z"
    },
    {
      "ID": 2,
      "IncidenteID": 1,
      "Detalle": "Se intentó nuevamente leer el archivo de datos y se logró con éxito",
      "FechaInicio": "2021-10-01T10:05:00Z",
      "FechaFin": "2021-10-01T10:10:00Z"
    }
  ]
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (IncidenteProceso) GetAll(db *gorm.DB) []*IncidenteProceso
```

Este método devuelve todos los incidentes relacionados con procesos registrados en la base de datos, incluyendo el proceso al que pertenecen y los detalles relacionados.

### Get

```go
func (this *IncidenteProceso) Get(db *gorm.DB, id uint)
```

Este método devuelve el incidente relacionado con proceso con el identificador `id`, incluyendo el proceso al que pertenece y los detalles relacionados.

### GetByProceso

```go
func (this *IncidenteProceso) GetByProceso(db *gorm.DB, procesoID uint) []*IncidenteProceso
```

Este método devuelve todos los incidentes relacionados con el proceso con el identificador `procesoID`, incluyendo el proceso al que pertenecen y los detalles relacionados.

## Estructura IncidentesDetalle

La estructura `IncidentesDetalle` representa un detalle relacionado con un incidente y tiene los siguientes campos:

- `ID`: identificador único del detalle.
- `IncidenteID`: identificador único del incidente al que pertenece el detalle (no puede ser nulo).
- `Detalle`: descripción del detalle (puede ser nulo).
- `FechaInicio`: fecha de inicio del detalle (puede ser nulo).
- `FechaFin`: fecha de finalización del detalle (puede ser nulo).

```json
{
    "id": 1,
    "incidente_id": 1,
    "detalle": "Error al procesar la factura",
    "fecha_inicio": "2021-05-01T10:30:00Z",
    "fecha_fin": "2021-05-01T11:00:00Z",
    "created_at": "2021-05-01T12:00:00Z",
    "updated_at": "2021-05-01T12:00:00Z",
    "deleted_at": null
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (IncidentesDetalle) GetAll(db *gorm.DB) []*IncidentesDetalle
```

Este método devuelve todos los detalles relacionados con incidentes registrados en la base de datos.

### Get

```go
func (this *IncidentesDetalle) Get(db *gorm.DB, id uint)
```

Este método devuelve el detalle con el identificador `id`.

### GetByIncidente

```go
func (this *IncidentesDetalle) GetByIncidente(db *gorm.DB, incidenteID int) []*IncidentesDetalle
```

Este método devuelve todos los detalles relacionados con el incidente con el identificador `incidenteID`.

## Estructura Usuario

La estructura `Usuario` representa un usuario del sistema y tiene los siguientes campos:

- `ID`: identificador único del usuario.
- `Nombre`: nombre del usuario (no puede ser nulo).
- `Apellido`: apellido del usuario (no puede ser nulo).
- `Email`: dirección de correo electrónico del usuario (no puede ser nulo).
- `Password`: contraseña del usuario (no puede ser nulo).
- `Roles`: lista de roles asignados al usuario.
- `Procesos`: lista de procesos asignados al usuario.

```json
{
    "ID": 1,
    "Nombre": "Juan",
    "Apellido": "Perez",
    "Email": "juan.perez@example.com",
    "Roles": [
        {
            "ID": 1,
            "Nombre": "Administrador",
            "Usuarios": [...]
        }
    ],
    "Procesos": [
        {
            "ID": 1,
            "Nombre": "Proceso de Ejemplo",
            "Folderid": 123456,
            "Organizacion": {...},
            "IncidentesProceso": [...],
            "Clientes": [...],
            "Usuarios": [...]
        }
    ]
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (Usuario) GetAll(db *gorm.DB) []*Usuario
```

Este método devuelve todos los usuarios registrados en la base de datos, incluyendo los roles y procesos asignados.

### Get

```go
func (this *Usuario) Get(db *gorm.DB, id uint)
```

Este método devuelve el usuario con el identificador `id`, incluyendo los roles y procesos asignados.

### GetByProcess

```go
func (Usuario) GetByProcess(db *gorm.DB, procesoID uint) []*Usuario
```

Este método devuelve todos los usuarios asignados al proceso con el identificador `procesoID`, incluyendo los roles y procesos asignados.

## Estructura Rol

La estructura `Rol` representa un rol del sistema y tiene los siguientes campos:

- `ID`: identificador único del rol.
- `Nombre`: nombre del rol (no puede ser nulo).
- `Usuarios`: lista de usuarios asignados al rol.

```json
{
  "ID": 1,
  "Nombre": "Administrador",
  "Usuarios": [...]
}
```

Además, la estructura cuenta con los siguientes métodos:

### GetAll

```go
func (Rol) GetAll(db *gorm.DB) []*Rol
```

Este método devuelve todos los roles registrados en la base de datos, incluyendo los usuarios asignados.

### Get

```go
func (this *Rol) Get(db *gorm.DB, id uint)
```

Este método devuelve el rol con el identificador `id`, incluyendo los usuarios asignados.

