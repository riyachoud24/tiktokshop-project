package main

import (
	"database/sql"
	//import mysql driver
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Define routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/submit", submitContent)

	// Start server on port 8080
	log.Println("Server starting on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// homePage serves the HTML form
func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

// submitContent handles the form submission and saves data to MySQL
func submitContent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Connect to the MySQL database
		db, err := sql.Open("mysql", "riyachoud24:OpenPass@tcp(127.0.0.1:3306)/tiktok_entry")
		if err != nil {
			log.Fatal("Failed to connect to the database:", err)
		}
		defer db.Close()

		// Parse the form data
		content := r.FormValue("content")

		// Insert the content into the database
		insertQuery := "INSERT INTO content_table (content) VALUES (?)"
		_, err = db.Exec(insertQuery, content)
		if err != nil {
			log.Fatal("Failed to insert data into the database:", err)
		}

		// Redirect back to the home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
