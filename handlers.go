package main

/*
this file is to define the logic for different handling routes
*/

import (
	"html/template"
	"log"
	"net/http"
)

// this handler will serve an HTML form when the user access the ('/) route
// TODO: write a index.html in the same dir
func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func submitConent(w http.ResponseWrite, r *http.Request) {
	if r.Method == http.MethodPost {
		content := r.FormValue("content")
		log.Println("Content submitted:", content)
		//Here you would normally save the content to a database
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
