package main

import (
    "encoding/json"
    "net/http"
    "strings"
)


// Middleware
func Middleware(next http.Handler) http.Handler {
    // Set application/json as content type for all the routes
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}


// Ping
func Ping(w http.ResponseWriter, r *http.Request) {
    // Return a simple {"msg": "AUDP APIs working"}
    response := map[string]string{"msg": "AUDP APIs working"}

    json.NewEncoder(w).Encode(response)
}


// Controllers
func ListControllers(w http.ResponseWriter, r *http.Request) {
    // Query controllers
    rows, _ := DB.Query(`SELECT id, url, name FROM controllers`)
    defer rows.Close()

    // Create the controllers list
    var controllers  []Controller
    for rows.Next() {
        var c Controller
        rows.Scan(&c.ID, &c.URL, &c.Name)
        controllers = append(controllers, c)
    }

    // Query devices
    rows, _ = DB.Query(`SELECT id, cid, name, status FROM devices`)
    defer rows.Close()

    // Add devices to the controllers' devices list
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.Name, &d.Status)

        c := controllers[d.CID - 1]
        c.Devices = append(c.Devices, d)
    }

    // Write the response
    if controllers != nil {
        json.NewEncoder(w).Encode(controllers)
    } else {
        json.NewEncoder(w).Encode([]Controller{})
    }
}

func AddController(w http.ResponseWriter, r *http.Request) {
    // Check Content-Type
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, `Body "Content-Type" must be "application/json"`, http.StatusUnsupportedMediaType)
        return
    }

    // Parse the controller
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Check controller's name
    if c.Name == "" {
        http.Error(w, `Missing controller's name`, http.StatusBadRequest)
        return
    }

    // Set controller's URL
    if ip := r.Header.Get("X-FORWARDED-FOR"); ip != "" {
        c.URL = "http://" + strings.Split(ip, ":")[0]
    } else {
        c.URL = "http://" + strings.Split(r.RemoteAddr, ":")[0]
    }

    // Check if every device is valid
    for did := range c.Devices {
        if c.Devices[did].Name == "" {
            http.Error(w, `Missing device's name`, http.StatusBadRequest)
            return
        }
    }

    // Save the controller
    err := c.SaveAll()

    // Check for errors
    switch err.Error() {
        case "UNIQUE constraint failed: controllers.name":
            http.Error(w, "Controller's name already used", http.StatusBadRequest); return

        case "UNIQUE constraint failed: controllers.url":
            http.Error(w, "Already registered a controller from that ip", http.StatusBadRequest); return

        case "UNIQUE constraint failed: devices.name":
            http.Error(w, "Device's name already used", http.StatusBadRequest); return
    }

    // If there have been no errors return the saved controller
    json.NewEncoder(w).Encode(c)
}
