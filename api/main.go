package main

import (
    "log"

    "net/http"
    "github.com/gorilla/mux"
    "github.com/jasonlvhit/gocron"
)


func main() {
    // Initialize the database
    err := Initialize_DB()
    if err != nil { log.Fatal(err) }

    // Initialize the router
    router := mux.NewRouter()
    router.Use(Middleware)

    // Add routes
    router.HandleFunc("/", Ping).Methods("GET")
    router.HandleFunc("/controllers", ListControllers).Methods("GET")
    router.HandleFunc("/controllers/add", AddController).Methods("POST")
    router.HandleFunc("/controllers/delete", DeleteController).Methods("POST")

    // Start the "pinger"
    go RunCronJobs()

    // Start the server
    log.Println("Starting AUDP API on http://localhost:8080...")
    log.Fatal(http.ListenAndServe(":8080", router))
}

func RunCronJobs() {
    // Check all the controllers every 10 seconds
    gocron.Every(10).Second().Do(func() {
        for _, controller := range GetControllersList(false) {
            controller.Check()
        }
    })

    // Start the cronjob
    <- gocron.Start()
}
