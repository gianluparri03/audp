package main

import (
    "log"

    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    // Initialize the database
    err := Initialize_DB()
    if err != nil { log.Fatal(err) }

    // Initialize the router
    router := mux.NewRouter()

    // Add routes
    router.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("working\n"))
    })

    // Start the server
    log.Println("Starting AUDP API on http://localhost:8080...")
    log.Fatal(http.ListenAndServe(":8080", router))
}
