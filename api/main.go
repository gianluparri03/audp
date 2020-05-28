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

    // Start the cronjobs
    go CronJobs()

    // Start the server
    log.Println("Starting AUDP API on http://localhost:8080...")
    log.Fatal(http.ListenAndServe(":8080", CreateRouter()))
}
