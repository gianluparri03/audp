package main

import (
    "strings"
    "strconv"
    "encoding/json"

    "net/http"
    "github.com/gorilla/mux"
)



type APIError struct {
    Error       string   `json:"error"`
    Description string   `json:"description"`
}


// Utility functions
func getIP(r *http.Request) string {
    if ip := r.Header.Get("X-FORWARDED-FOR"); ip != "" {
        return strings.Split(ip, ":")[0]
    }

    return strings.Split(r.RemoteAddr, ":")[0]
}

func ReturnError(w http.ResponseWriter, code int, err APIError) {
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(&err)
}


// Middlewares
func SetContentType(next http.Handler) http.Handler {
    // Set application/json as content type for all the routes
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

func CheckContentType(next http.Handler) http.Handler {
    // Return a 415 error if requet's Content-Type isn't application/json
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Content-Type") == "application/json" {
            next.ServeHTTP(w, r)
        } else {
            http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
        }
    })
}


// "Stand Alone" Endpoints
func Ping(w http.ResponseWriter, r *http.Request) {
    response := map[string]string{"msg": "AUDP APIs working!", "version": "dev"}
    json.NewEncoder(w).Encode(response)
}


// Controllers' Endpoints
func ListControllers(w http.ResponseWriter, r *http.Request) {
    // Create an empty list
    var controllers []Controller

    // Query the controllers
    crows, _ := DB.Query(`SELECT id, ip, port, name, sleeping FROM controllers;`)
    defer crows.Close()

    for crows.Next() {
        // Fetch the controller
        var c Controller
        crows.Scan(&c.ID, &c.IP, &c.Port, &c.Name, &c.Sleeping)

        // Query the connected devices
        drows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices WHERE cid=?;`, c.ID)
        defer drows.Close()

        // Gotta fetch them all! (cit.)
        for drows.Next() {
            var d Device
            drows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
            c.Devices = append(c.Devices, d)
        }

        // Add the controller to the controllers' list
        controllers = append(controllers, c)
    }

    // Write the response
    json.NewEncoder(w).Encode(controllers)
}

func GetController(w http.ResponseWriter, r *http.Request) {
    // Get controller's name
    name := mux.Vars(r)["name"]

    // Try to retrieve it from the database
    var c Controller
    query := `SELECT id, ip, port, name, sleeping FROM controllers WHERE name=?;`
    DB.QueryRow(query, name).Scan(&c.ID, &c.IP, &c.Port, &c.Name, &c.Sleeping)

    // Check if it exists
    if c.ID == 0 {
        ReturnError(w, 404, APIError{"controller not found", "Didn't find a controller with that name"})
        return
    }

    // Query the connected devices
    rows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices WHERE cid=?;`, c.ID)
    defer rows.Close()

    // Fetch them
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
        c.Devices = append(c.Devices, d)
    }

    // Return the controller
    json.NewEncoder(w).Encode(c)
}

func GetControllerDevices(w http.ResponseWriter, r *http.Request) {
    // Get controller's name
    name := mux.Vars(r)["name"]

    // Check if it exists
    var id int64
    DB.QueryRow(`SELECT id FROM controllers WHERE name=?;`, name).Scan(&id)
    if id == 0 {
        ReturnError(w, 404, APIError{"controller not found", "Didn't find a controller with that name"})
        return
    }

    // Create an empty list
    var devices []Device

    // Query the devices
    rows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices WHERE cid=?;`, id)
    defer rows.Close()

    // Fetch them
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
        devices = append(devices, d)
    }

    // Write the response
    json.NewEncoder(w).Encode(devices)
}

func CreateController(w http.ResponseWriter, r *http.Request) {
    // Parse the controller
    var c Controller
    json.NewDecoder(r.Body).Decode(&c)

    // Check wether it has a name and a port
    if c.Name == "" {
        ReturnError(w, 400, APIError{"invalid controller", "Missing controller's name"})
        return
    } else if c.Port == 0 {
        ReturnError(w, 400, APIError{"invalid controller", "Missing controller's port"})
        return
    }

    // Set controller's IP
    c.IP = getIP(r)

    // Try to save the controller in the database
    query := `INSERT INTO controllers (ip, port, name) VALUES (?, ?, ?);`
    res, err := DB.Exec(query, c.IP, c.Port, c.Name)

    // If an error has been raised, return it
    if err != nil {
        apierror := APIError{Error: "can't save controller"}

        // Write a description
        if err.Error() == "UNIQUE constraint failed: controllers.name" {
            apierror.Description = "Controller's name already used"
        } else if err.Error() == "UNIQUE constraint failed: controllers.ip" {
            apierror.Description = "Controller's IP already used"
        } else {
            apierror.Description = err.Error()
        }

        ReturnError(w, 409, apierror)
        return
    }

    // Save new ID
    c.ID, _ = res.LastInsertId()

    // Return it
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(&c)
}

func WakeUpController(w http.ResponseWriter, r *http.Request) {
    // Get controller's name
    vars := mux.Vars(r)
    name := vars["name"]
    port, _ := strconv.ParseUint(vars["port"], 10, 32)
    ip := getIP(r)

    // Fetch it from the db
    var n int64
    var sleeping bool
    DB.QueryRow(`SELECT count(*), sleeping FROM controllers WHERE name=?;`, name).Scan(&n, &sleeping)

    // Check if it exists and if it's sleeping
    if n == 0 {
        ReturnError(w, 404, APIError{"can't wake up controller", "Controller doesn't exist"})
        return
    } else if !sleeping {
        ReturnError(w, 409, APIError{"can't wake up controller", "Controller isn't sleeping"})
        return
    }

    // Check if ip is already used
    DB.QueryRow(`SELECT count(*) FROM controllers WHERE ip=?;`, ip).Scan(&n)
    if n != 0 {
        ReturnError(w, 409, APIError{"can't wake up controller", "IP already used"})
        return
    }

    // Set sleeping to false and update port and ip
    if port != 0 {
        DB.Exec(`UPDATE controllers SET sleeping=false, ip=?, port=? WHERE name=?;`, ip, port, name)
    } else {
        DB.Exec(`UPDATE controllers SET sleeping=false, ip=? WHERE name=?;`, ip, name)
    }

    // Fetch the controller from the db
    var c Controller
    query := `SELECT id, ip, port, name, sleeping FROM controllers WHERE name=?;`
    DB.QueryRow(query, name).Scan(&c.ID, &c.IP, &c.Port, &c.Name, &c.Sleeping)

    // Query the connected devices
    rows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices WHERE cid=?;`, c.ID)
    defer rows.Close()

    // Fetch them
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
        c.Devices = append(c.Devices, d)
    }

    // Return the controller
    json.NewEncoder(w).Encode(c)
}

func DeleteController(w http.ResponseWriter, r *http.Request) {
    // Get controller's name
    name := mux.Vars(r)["name"]

    // Check if it exists
    var id int64
    DB.QueryRow(`SELECT id FROM controllers WHERE name=?;`, name).Scan(&id)
    if id == 0 {
        ReturnError(w, 404, APIError{"Can't delete controller", "Controller doesn't exist"})
        return
    }

    // Delete the controller and the associated devices
    DB.Exec(`DELETE FROM controllers WHERE id=?;`, id)
    DB.Exec(`DELETE FROM devices WHERE cid=?;`, id)

    // Return a 204
    w.WriteHeader(204)
}


// Devices' Endpoints
func ListDevices(w http.ResponseWriter, r *http.Request) {
    // Create an empty list
    var devices []Device

    // Query the devices
    rows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices`)
    defer rows.Close()

    // Fetch them
    for rows.Next() {
        var d Device
        rows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
        devices = append(devices, d)
    }

    // Write the response
    json.NewEncoder(w).Encode(devices)
}
