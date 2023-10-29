# Guía Rápida para Poner en Marcha el servidor API REST

## Cosas que Necesitarás Antes de Empezar
Antes de entrar en materia, asegúrate de tener instalado lo siguiente:
- [Go (Golang)](https://golang.org/dl/): Versión 1.18 o algo más reciente.
- [PostgreSQL](https://www.postgresql.org/download/): Nuestra base de datos para guardar toda la info importante.

## Bajando el Código
1. Abre tu terminal y clona el repositorio en tu máquina:
   ```
   git clone https://github.com/oxakromax/Backend_UipathMonitor.git
   cd Backend_UipathMonitor
   ```

## Ajustando las Variables de Entorno
2. Ahora, crea un archivo `.env` en la carpeta principal del proyecto. Aquí vamos a poner todas las configuraciones necesarias:
   ```
   PORT=8080
   PGHOST=localhost
   PGDATABASE=nombre_basededatos
   PGPASSWORD=contraseña_basededatos
   PGPORT=5432
   PGUSER=usuario_basededatos
   PGSSLMODE=disable
   DB_KEY=clave_secreta_bd
   MONITOR_PASS=contraseña_monitor
   MONITOR_USER=monitor@dominio.com
   MAIL_ADRESS=correo_monitor@dominio.com
   MAIL_PASSWORD=contraseña_correo
   MAIL_SMTP_SERVER=smtp.dominio.com
   MAIL_SMTP_PORT=587
   TOKEN_KEY=clave_secreta_token
   SSL_CERT=Path/To/Cert
   SSL_KEY=Path/To/Key
   ```

   **Ojo aquí**: Cambia los valores de ejemplo por los tuyos propios.

## Instalando las Dependencias
3. Ahora, instalemos las dependencias del proyecto:
   ```
   go mod tidy
   ```

## Compilación y Puesta en Marcha
4. Vamos a compilar y poner en marcha la aplicación:
   ```
   go build -o uipathmonitor
   ./uipathmonitor
   ```

   Si todo sale bien, deberías ver un mensaje diciendo que la aplicación está corriendo y escuchando en el puerto que configuraste.

## ¿Y Ahora? ¡A Usar la App!
Recuerda que este es el servidor API Rest de un conjunto de servicios, para usar la App deberás de ver el siguiente repositorio: [Frontend](https://github.com/oxakromax/Frontend_UipathMonitor)

Además para disfrutar de todas las cualidades del sistema, el servicio de monitoreo se encuentra en el siguiente respositorio: [Servicio Monitor](https://github.com/oxakromax/Monitor_UipathMonitor)

De todas maneras puedes probar cada una de las rutas de la app con Postman, aquí tienes un repositorio de ejemplos de peticiones para que puedas explorar las funcionalidades 🙌: [Postman Repository](https://www.postman.com/altimetry-candidate-39737582/workspace/api-central-backend/collection/26219135-e7851605-7c71-45f3-a48b-de4d8c9f185e?action=share&creator=26219135)

---

## Algunas Cositas sobre las Variables de Entorno

Las variables de entorno son como el corazón del servidor, así que vamos a darles un repaso:

- `PORT`: Aquí decides en qué puerto va a correr la aplicación.
- `PG*`: Todo lo que necesitas para conectar con PostgreSQL.
- `DB_KEY`: Una llavecita secreta para encriptar la comunicación con la base de datos. Si no sabes generar una llave AES por ti mismo, no te preocupes, solo haz que inicie el servidor sin este parametro y te otorgará una nueva... Aunque no iniciará en absoluto si no le das este parametro
- `MONITOR_PASS` y `MONITOR_USER`: Lo que necesitarás para que el servicio de monitorización hable con la API.
- `MAIL*`: Todo lo necesario para enviar correos electrónicos a traves de SMTP.
- `TOKEN_KEY`: Otra llave secreta, esta vez para generar los tokens JWT.
- `SSL_CERT` y `SSL_KEY`: Las rutas a los archivos de certificado SSL y la llave privada. Si los dejas en blanco, la aplicación usará HTTP sin encriptar. ¡Pero ojo!, esto no es lo más seguro del mundo.

Aunque estas variables las configuramos en el archivo `.env`, también puedes cambiarlas usando herramientas de contenerización si necesitas algo más personalizado. Lo importante es mantener el formato que te mostramos para que todo funcione como debe.

---


## ¿Y si Quiero Usar Docker?

Si prefieres usar Docker, puedes hacerlo. Aquí te dejamos una guía rápida para que te pongas en marcha:

### Dockerfile

```Dockerfile
# Usamos la imagen oficial de Go como base
FROM golang:1.21 as builder

# Instalamos git (necesario para go mod)
RUN apk update && apk add --no-cache git

# Establecemos el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiamos los archivos del proyecto al contenedor
COPY . .

# Instalamos las dependencias
RUN go mod tidy

# Compilamos la aplicación
RUN go build -o uipathmonitor

# Creamos una imagen ligera usando alpine
FROM alpine:latest

# Copiamos el binario compilado desde la imagen builder
COPY --from=builder /app/uipathmonitor /app/

# Exponemos el puerto 8080 para acceder a la aplicación
EXPOSE 8080

# Establecemos el comando por defecto para ejecutar la aplicación
CMD ["/app/uipathmonitor"]
```

---

### Pasos para Dockerizar la App:

1. Asegúrate de tener instalado [Docker](https://www.docker.com/products/docker-desktop) en tu máquina.
2. En la raíz del proyecto, donde se encuentra el `Dockerfile`, construye la imagen de Docker:
   ```bash
   docker build -t backend_uipathmonitor .
   ```
3. Una vez construida la imagen, levanta un contenedor:
   ```bash
   docker run -p 8080:8080 --env-file .env backend_uipathmonitor
   ```

¡Y listo! El servidor ahora está corriendo dentro de un contenedor Docker en el puerto 8080. Fácil, ¿verdad? 😊


