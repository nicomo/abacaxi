package views

import (
	"errors"
	"html/template"
	"net/http"
)

// global vars
var (
	tmpl                    map[string]*template.Template // we bundle our templates in a single map of templates
	ErrTemplateDoesNotExist = errors.New("The template does not exist.")
)

// load templates on init
// base is our base template calling all other templates
func init() {

	if tmpl == nil {
		tmpl = make(map[string]*template.Template)
	}
	// home page
	tmpl["home"] = template.Must(template.ParseFiles("templates/index.tmpl",
		"templates/base.tmpl",
		"templates/head.tmpl",
	))
	// targetservice page
	tmpl["targetservice"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/package.tmpl",
	))

	// file uploaded
	tmpl["upload"] = template.Must(template.ParseFiles("templates/base.tmpl",
		"templates/head.tmpl",
		"templates/upload.tmpl",
	))

}

// RenderTmpl is a wrapper around template.ExecuteTemplate
func RenderTmpl(w http.ResponseWriter, name string, data map[string]interface{}) error {

	//make sure template actually exists
	tmpl, ok := tmpl[name]
	if !ok {
		return ErrTemplateDoesNotExist
	}
	tmpl.ExecuteTemplate(w, "base", data)
	return nil
}
