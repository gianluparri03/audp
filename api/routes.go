package main

import (
    "net/http"
    "strings"
    "encoding/json"
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
    response := map[string]string{"msg": "AUDP APIs working", "version": "v0.1dev"}
    json.NewEncoder(w).Encode(response)
}


// Controllers
func ListControllers(w http.ResponseWriter, r *http.Request) {
    // Get controllers (with devices)
    controllers := GetControllersList(true)

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

    // Check if there is all the required parameters
    if c.Name == "" {
        http.Error(w, `Missing controller's name`, http.StatusBadRequest)
        return
    } else if c.Port == 0 {
        http.Error(w, `Missing controller's port`, http.StatusBadRequest)
        return
    } else if c.MAC == "" {
        http.Error(w, `Missing controller's mac address`, http.StatusBadRequest)
        return
    }

    // Set controller's IP
    if ip := r.Header.Get("X-FORWARDED-FOR"); ip != "" {
        c.IP = strings.Split(ip, ":")[0]
    } else {
        c.IP = strings.Split(r.RemoteAddr, ":")[0]
    }

    // Check if every device is valid
    for _, d := range c.Devices {
        if d.Name == "" {
            http.Error(w, `Missing device's name`, http.StatusBadRequest)
            return
        } else if d.GPIO == nil {
            http.Error(w, `Missing device's gpio`, http.StatusBadRequest)
            return
        }
    }

    // Save the controller
    err := c.SaveAll()

    // Check for errors
    if (err != nil) { switch err.Error() {
        case "Already registered a controller with that URL", "Device's GPIO is already used":
            http.Error(w, err.Error(), http.StatusBadRequest); return

        case "UNIQUE constraint failed: controllers.name":
            http.Error(w, "Controller's name already used", http.StatusBadRequest); return

        case "UNIQUE constraint failed: controllers.mac":
            http.Error(w, "Already registered a controller with that mac address", http.StatusBadRequest); return

        case "UNIQUE constraint failed: devices.name":
            http.Error(w, "Device's name already used", http.StatusBadRequest); return

        default:
            http.Error(w, err.Error(), http.StatusInternalServerError); return
    }}

    // Check if controller is sleeping
    c.Check()

    // If there have been no errors return the saved controller
    json.NewEncoder(w).Encode(c)
}

func WakeUpController(w http.ResponseWriter, r *http.Request) {
    // Check Content-Type
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, `Body "Content-Type" must be "application/json"`, http.StatusUnsupportedMediaType)
        return
    }

    // Parse the controller
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Check if there is all the required parameters
    if c.MAC == "" {
        http.Error(w, `Missing controller's mac address`, http.StatusBadRequest)
        return
    }

    // Get that controller
    var id int64
    query := "SELECT id FROM controllers WHERE mac=?"
    DB.QueryRow(query, c.MAC).Scan(&id)

    // Check f it exists, otherwise return an error
    if id == 0 {
        http.Error(w, "There isn't a controller with that MAC address", http.StatusBadRequest)
        return
    }

    // Fetch the controller
    c, _ = FetchController(id, true)

    // Check if it's sleeping (if not return an error)
    if !c.Sleeping {
        http.Error(w, "The controller isn't sleeping", http.StatusBadRequest)
        return
    }

    // Set it to awake
    c.Sleeping = false
    DB.Exec(`UPDATE controllers SET sleeping=false WHERE id = ?`, c.ID)

    // If there have been no errors return the saved controller
    json.NewEncoder(w).Encode(c)
}

func DeleteController(w http.ResponseWriter, r *http.Request) {
    // Check Content-Type
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, `Body "Content-Type" must be "application/json"`, http.StatusUnsupportedMediaType)
        return
    }

    // Parse the controller
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Delete it
    err := c.Delete()

    // Check for errors
    if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }

    // If no errors have been raised return an ok
    response := map[string]string{"msg": "Done"}
    json.NewEncoder(w).Encode(response)
}
