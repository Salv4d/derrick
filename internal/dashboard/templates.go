package dashboard

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	pageTmpl *template.Template
	rowsTmpl *template.Template
	rowTmpl  *template.Template
)

func init() {
	pageTmpl = template.Must(template.New("page").ParseFS(templateFS, "templates/index.html", "templates/rows.html"))
	rowsTmpl = template.Must(template.New("rows").ParseFS(templateFS, "templates/rows.html"))
	rowTmpl  = template.Must(template.New("row").ParseFS(templateFS, "templates/rows.html"))
}

func renderPage(w io.Writer, views []projectView) error {
	return pageTmpl.ExecuteTemplate(w, "index.html", views)
}

func renderRows(w io.Writer, views []projectView) error {
	return rowsTmpl.ExecuteTemplate(w, "tbody", views)
}

func renderRow(w io.Writer, v projectView) error {
	return rowTmpl.ExecuteTemplate(w, "row", v)
}
