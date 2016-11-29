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
	errEmptyMessage         = errors.New("Ah, the message seems empty. Can you try again?")
	errUserNotFound         = errors.New("There doesn't seem to be any user by that name...")
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
		"templates/packages.tmpl",
	))
	// epackage page
	tmpl["epackage"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/package.tmpl",
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
