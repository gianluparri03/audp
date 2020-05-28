package main

import (
	"net/http"
    "github.com/jasonlvhit/gocron"
)


func CreateRouter() mux.Router {
    // Create the main router
    router := mux.NewRouter()
    router.Use(Middleware)

    // Create subrouters
    get_apis := router.Methods("GET").Subrouter()   
    post_apis := router.Methods("POST").Subrouter()
    post_apis.Use(PostMiddleware)

    // Add APIs
    get_apis.HandleFunc("/", Ping)
    get_apis.HandleFunc("/controllers", ListControllers)
    post_apis.HandleFunc("/controllers/add", AddController)
    post_apis.HandleFunc("/controllers/wakeup", WakeUpController)
    post_apis.HandleFunc("/controllers/delete", DeleteController)

    return router
}

func CronJobs() {
    // Run c.Check() for each controller every 10 seconds
    gocron.Every(10).Second().Do(func() {
        for _, controller := range GetControllersList(false) {
            if !controller.Sleeping {
                controller.Check()
            }
        }
    })

    // Start the cronjob
    <- gocron.Start()
}

func Middleware(next http.Handler) http.Handler {
    // Set application/json as content type for all the routes
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

func PostMiddleware(next http.Handler) http.Handler {
	// Return a 415 error if requet's Content-Type isn't application/json
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Content-Type") == "application/json" {
            next.ServeHTTP(w, r)
        } else {
            http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
        }
    })
}
