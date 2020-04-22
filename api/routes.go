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
    // Query controllers
    rows, _ := DB.Query(`SELECT id, ip, mac, port, name, sleeping FROM controllers`)
    defer rows.Close()

    // Create a map of controllers (id: controller)
    controllers := make(map[int64]Controller)
    for rows.Next() {
        var c Controller
        rows.Scan(&c.ID, &c.IP, &c.MAC, &c.Port, &c.Name, &c.Sleeping)
        controllers[c.ID] = c
    }

    // Query devices
    rows, _ = DB.Query(`SELECT id, cid, name, status FROM devices`)
    defer rows.Close()

    // Add devices to the controllers' devices list
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.Name, &d.Status)

        c := controllers[d.CID]
        c.Devices = append(c.Devices, d)
    }

    // Create the controllers list
    var controllers_list []Controller
    for _, c := range controllers {
        controllers_list = append(controllers_list, c)
    }

    // Write the response
    if controllers_list != nil {
        json.NewEncoder(w).Encode(controllers_list)
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
    for did := range c.Devices {
        if c.Devices[did].Name == "" {
            http.Error(w, `Missing device's name`, http.StatusBadRequest)
            return
        }
    }

    // Save the controller
    err := c.SaveAll()

    // Check for errors
    if (err != nil) { switch err.Error() {
        case "UNIQUE constraint failed: controllers.name":
            http.Error(w, "Controller's name already used", http.StatusBadRequest); return

        case "Already registered a controller with that URL":
            http.Error(w, "Already registered a controller with that URL", http.StatusBadRequest); return

        case "UNIQUE constraint failed: controllers.mac":
            http.Error(w, "Already registered a controller with that mac address", http.StatusBadRequest); return

        case "UNIQUE constraint failed: devices.name":
            http.Error(w, "Device's name already used", http.StatusBadRequest); return

        default:
            http.Error(w, err.Error(), http.StatusInternalServerError); return
    }}

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
