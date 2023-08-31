package Mail

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
)

const PathLogo = "Mail/Templates/Assets/Logo.png"
const PathIncidentChange = "Mail/Templates/IncidentChange.html"
const PathNewIncident = "Mail/Templates/NewTicket.html"
const PathNewUser = "Mail/Templates/NewUser.html"
const PathNewPassword = "Mail/Templates/NewPassword.html"

type IncidentChange struct {
	ID            int
	NombreProceso string
	NuevoEstado   string
	LogoBase64    string
}

type NewIncident struct {
	ID            int
	NombreProceso string
	Descripcion   string
	Tipo          string
	LogoBase64    string
}

type NewUser struct {
	Nombre     string
	Email      string
	Password   string
	LogoBase64 string
}

type NewPassword struct {
	Nombre     string
	Password   string
	LogoBase64 string
}

func ConvertToBase64(pathToImage string) string {
	file, err := os.Open(pathToImage)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer file.Close()
	// Read entire JPG/PNG into byte slice.
	reader := bufio.NewReader(file)
	content, _ := io.ReadAll(reader)
	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	return encoded
}

func GetBodyNewTicket(newIncident NewIncident) string {
	newIncident.LogoBase64 = ConvertToBase64(PathLogo)
	t, _ := template.ParseFiles(PathNewIncident)
	body := new(strings.Builder)
	t.Execute(body, newIncident)
	return body.String()
}

func GetBodyTicketChange(incidentChange IncidentChange) string {
	incidentChange.LogoBase64 = ConvertToBase64(PathLogo)
	t, _ := template.ParseFiles(PathIncidentChange)
	body := new(strings.Builder)
	t.Execute(body, incidentChange)
	return body.String()
}

func GetBodyNewUser(newUser NewUser) string {
	newUser.LogoBase64 = ConvertToBase64(PathLogo)
	t, _ := template.ParseFiles(PathNewUser)
	body := new(strings.Builder)
	t.Execute(body, newUser)
	return body.String()
}

func GetBodyNewPassword(newPassword NewPassword) string {
	newPassword.LogoBase64 = ConvertToBase64(PathLogo)
	t, _ := template.ParseFiles(PathNewPassword)
	body := new(strings.Builder)
	t.Execute(body, newPassword)
	return body.String()
}
