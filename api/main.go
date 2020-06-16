package main

import (
    "fmt"

    "net/http"
    "github.com/gorilla/mux"

    "github.com/jasonlvhit/gocron"
)



func main() {
    // Initialize the database
    err := InitializeDB("database.db")
    if err != nil {
        fmt.Println(err)
        return
    }

    // Start the controllers pinger
    go func() {
        pinger := gocron.NewScheduler()
        pinger.Every(30).Minutes().Do(PingControllers)
        <- pinger.Start()
    }()

    // Create the main router
    router := mux.NewRouter()
    router.Use(SetContentType)

    // Register GET endpoints
    GET := router.Methods("GET").Subrouter()
    GET.HandleFunc("/", Ping)
    GET.HandleFunc("/controllers", ListControllers)
    GET.HandleFunc("/controllers/{name}", GetController)
    GET.HandleFunc("/controllers/{name}/devices", GetControllerDevices)
    GET.HandleFunc("/devices", ListDevices)

    // Register POST endpoints
    POST := router.Methods("POST").Subrouter()
    POST.Use(CheckContentType)
    POST.HandleFunc("/controllers", CreateController)

    // Register PUT endpoints
    PUT := router.Methods("PUT").Subrouter()
    PUT.Use(CheckContentType)
    PUT.HandleFunc("/controllers/{name}/wakeup/{port}", WakeUpController)

    // Register DELETE endpoints
    DELETE := router.Methods("DELETE").Subrouter()
    DELETE.Use(CheckContentType)
    DELETE.HandleFunc("/controllers/{name}", DeleteController)

    // Start the server
    fmt.Println("Starting AUDP API on http://localhost:8080...")
    fmt.Println(http.ListenAndServe(":8080", router))
}
