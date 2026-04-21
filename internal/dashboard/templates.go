package dashboard

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	pageTmpl   *template.Template
	rowsTmpl   *template.Template
	rowTmpl    *template.Template
	detailTmpl *template.Template
)

func init() {
	all := template.Must(template.New("").ParseFS(templateFS,
		"templates/index.html",
		"templates/rows.html",
		"templates/detail.html",
	))
	pageTmpl   = all
	rowsTmpl   = all
	rowTmpl    = all
	detailTmpl = all
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

func renderDetail(w io.Writer, dv detailView) error {
	return detailTmpl.ExecuteTemplate(w, "detail", dv)
}
