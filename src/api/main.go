package main

import (
    "fmt"
    "log"

    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    // Connect to the database
    _, err := Initialize_DB()

    // Look for errors
    if err != nil { log.Fatal(err) }

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
