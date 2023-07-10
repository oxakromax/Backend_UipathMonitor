package Server

import (
	"bytes"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/oxakromax/Backend_UipathMonitor/ORM"
	"github.com/xuri/excelize/v2"
)

// Check if user is logged in, and has the correct role (organization).
// ---------------------------------------------------------
// Sheet 'Organization Data' uses the following query:
// SELECT o.nombre AS organizacion, COUNT(DISTINCT u.id) AS usuarios_asignados, COUNT(DISTINCT c.id) AS clientes_asignados
// FROM organizaciones o
// LEFT JOIN usuarios_organizaciones uo ON uo.organizacion_id = o.id
// LEFT JOIN usuarios u ON u.id = uo.usuario_id
// LEFT JOIN clientes c ON c.organizacion_id = o.id
// GROUP BY o.nombre;
// ---------------------------------------------------------
// And then generates an Excel file with the results.
// send it as a response to the client without saving it to disk.
// Func name: GetOrgData
func (H *Handler) GetOrgData(c echo.Context) error {

	UserID := uint(c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["id"].(float64))
	User := new(ORM.Usuario)
	User.Get(H.Db, UserID)

	if !User.HasRole("organization") {
		return c.JSON(403, "You don't have permission to do this")
	}

	type OrganizationData struct {
		Organizacion      string `json:"organizacion"`
		UsuariosAsignados int    `json:"usuarios_asignados"`
		ClientesAsignados int    `json:"clientes_asignados"`
	}

	var orgData []OrganizationData

	Query := H.Db.Raw(`
		SELECT o.nombre AS organizacion, COUNT(DISTINCT u.id) AS usuarios_asignados, COUNT(DISTINCT c.id) AS clientes_asignados
		FROM organizaciones o
		LEFT JOIN usuarios_organizaciones uo ON uo.organizacion_id = o.id
		LEFT JOIN usuarios u ON u.id = uo.usuario_id
		LEFT JOIN clientes c ON c.organizacion_id = o.id
		GROUP BY o.nombre;
	`).Scan(&orgData)

	if Query.Error != nil {
		return c.JSON(500, "Failed to fetch organization data")
	}
	// Generate Excel file in memory
	file := excelize.NewFile()
	defer file.Close()
	sheetName := "Organization Data"
	file.NewSheet(sheetName)
	file.DeleteSheet("Sheet1")

	file.SetCellValue(sheetName, "A1", "Organización")
	file.SetCellValue(sheetName, "B1", "Usuarios Asignados")
	file.SetCellValue(sheetName, "C1", "Clientes Asignados")

	for i, data := range orgData {
		row := i + 2
		file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), data.Organizacion)
		file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.UsuariosAsignados)
		file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), data.ClientesAsignados)
	}

	// Add a table with style
	showRowStripes := true
	tableStyle := &excelize.Table{
		Range:             fmt.Sprintf("A1:C%d", len(orgData)+1),
		StyleName:         "TableStyleMedium2",
		ShowFirstColumn:   true,
		ShowLastColumn:    true,
		ShowRowStripes:    &showRowStripes,
		ShowColumnStripes: false,
	}

	if err := file.AddTable(sheetName, tableStyle); err != nil {
		return c.JSON(500, "Failed to add table style to organization data")
	}

	// Save Excel file
	file.Save()

	// Write Excel file to memory buffer
	buf := new(bytes.Buffer)
	if err := file.Write(buf); err != nil {
		return c.JSON(500, "Failed to write organization data to memory")
	}

	// Set response headers
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=organization_data.xlsx")

	// Send Excel file buffer as response
	return c.Blob(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())

}
