/*package main

import (
	"html/template"
	"log"
	"net/http"
)

// homePage serves the home page with the form
func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

// submitContent handles the form submission
func submitContent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		content := r.FormValue("content")
		log.Println("Content submitted:", content)
		// Here you would normally save the content to a database
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
*/