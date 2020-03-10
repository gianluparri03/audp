package main

import (
	"fmt"

	// WebServer
	"net/http"
	"github.com/gorilla/mux"

	// Database
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)


func main() {
	// Prepare the database
	db, _ := sql.Open("sqlite3", "database.db")
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS controllers (id INTEGER PRIMARY KEY)")
    statement.Exec()

    // Initialize the router
	r := mux.NewRouter()

	// Add ping route
	r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "working")
	})

	// Start the server
	fmt.Println("Starting AUDP API on http://localhost:8080...")
	http.ListenAndServe(":8080", r)
}
