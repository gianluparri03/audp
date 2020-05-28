package main

import (
    "net/http"
    "strings"
    "encoding/json"
)


func getIP(r *http.Request) string {
    if ip := r.Header.Get("X-FORWARDED-FOR"); ip != "" {
        return strings.Split(ip, ":")[0]
    } else {
        return strings.Split(r.RemoteAddr, ":")[0]
    }
}


func Ping(w http.ResponseWriter, r *http.Request) {
    response := map[string]string{"msg": "AUDP APIs working!", "version": "dev"}
    json.NewEncoder(w).Encode(response)
}

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
    } else if c.Code == "" {
        http.Error(w, `Missing controller's code`, http.StatusBadRequest)
        return
    }

    // Set controller's IP
    c.IP = getIP(r)

    // Save the controller
    err := c.Save()

    // Check for errors
    if (err != nil) {
        switch err.Error() {
            case "UNIQUE constraint failed: controllers.name":
                http.Error(w, "Controller's name already used", http.StatusBadRequest)

            case "UNIQUE constraint failed: controllers.ip":
                http.Error(w, "Controller's IP already used", http.StatusBadRequest)

            case "UNIQUE constraint failed: controllers.code":
                http.Error(w, "Controller's code already used", http.StatusBadRequest)

            default:
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
        return
    }

    // If there have been no errors return the saved controller
    json.NewEncoder(w).Encode(c)
}

func WakeUpController(w http.ResponseWriter, r *http.Request) {
    // Parse the controller from the request
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Check if there is all the required parameters
    if c.Code == "" {
        http.Error(w, `Missing controller's code`, http.StatusBadRequest)
        return
    }

    // Fetch that controller from the db
    var id int64
    var sleeping bool
    query := "SELECT id, sleeping FROM controllers WHERE code=?"
    DB.QueryRow(query, c.Code).Scan(&id, &sleeping)

    // Check if it exists and if it's sleeping
    if id == 0 {
        http.Error(w, "Controller not found", http.StatusBadRequest)
        return
    } else if !sleeping {
        http.Error(w, "Controller isn't sleeping", http.StatusBadRequest)
        return
    }

    // Set sleeping to false and update port and ip
    if c.Port != 0 {
        DB.Exec(`UPDATE controllers SET sleeping=false, ip=?, port=? WHERE id = ?`, getIP(r), c.Port, c.ID)
    } else {
        DB.Exec(`UPDATE controllers SET sleeping=false, ip=? WHERE id = ?`, c.IP, c.ID)
    }

    // If there have been no errors return the saved controller
    c, _ = GetControllerFromId(id, true)
    json.NewEncoder(w).Encode(c)
}

func DeleteController(w http.ResponseWriter, r *http.Request) {
    // Parse the controller
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Delete it
    err := c.Delete()

    // Check for errors
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // If no errors have been raised return an ok
    response := map[string]string{"msg": "Done"}
    json.NewEncoder(w).Encode(response)
}
