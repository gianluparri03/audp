package main

import (
    "fmt"

    "database/sql"
    _ "github.com/mattn/go-sqlite3"

    "net/http"
)



var DB *sql.DB


type Controller struct {
    ID       int64    `json:"id"`
    IP       string   `json:"ip"`
    MAC      string   `json:"mac_address"`
    Port     int64    `json:"port"`
    Name     string   `json:"name"`
    Devices  []Device `json:"devices"`
    Sleeping bool     `json:"sleeping"`
}

func GetControllersList(devices bool) []Controller {
    // Query controllers
    var controllers []Controller

    // Query the controllers
    controllers_rows, _ := DB.Query(`SELECT id, ip, mac, port, name, sleeping FROM controllers`)
    defer controllers_rows.Close()

    // Create a map of controllers (id: controller)
    for controllers_rows.Next() {
        // Create the controller
        var c Controller
        controllers_rows.Scan(&c.ID, &c.IP, &c.MAC, &c.Port, &c.Name, &c.Sleeping)

        // If devices is true
        if devices {
            // Query the devices
            devices_rows, _ := DB.Query(`SELECT id, cid, gpio, name, status FROM devices WHERE cid = ?`, c.ID)
            defer devices_rows.Close()

            // Add devices to the controllers' devices list
            for devices_rows.Next() {
                var d Device
                devices_rows.Scan(&d.ID, &d.CID, &d.GPIO, &d.Name, &d.Status)
                c.Devices = append(c.Devices, d)
            }
        }

        // Add the controller to the controllers list
        controllers = append(controllers, c)
    }

    return controllers
}

func (c *Controller) Save() (error) {
    // Check URI
    query := "SELECT name FROM controllers WHERE ip = ? AND port = ?"
    var query_result string
    DB.QueryRow(query, c.IP, c.Port).Scan(&query_result)
    if query_result != "" { return fmt.Errorf("Already registered a controller with that URL") }

    // Insert controller in the database
    result, err := DB.Exec(`INSERT INTO controllers (ip, mac, port, name, sleeping) VALUES (?, ?, ?, ?, ?)`, c.IP, c.MAC, c.Port, c.Name, c.Sleeping)
    if err != nil { return err }

    // Set controller's id
    c.ID, _ = result.LastInsertId()

    // Return nil if no error have been encountered
    return nil
}

func (c *Controller) SaveAll() (error) {
    // Save controller
    if err := c.Save(); err != nil { return err }

    // Save every device
    for i, d := range c.Devices {
        d.CID = c.ID
        err := d.Save()
        if err != nil { return err }
        c.Devices[i] = d
    }

    // Return nil if no error have been encountered
    return nil
}

func (c *Controller) Delete() (error) {
    // If there isn't the id raise an error
    if c.ID == 0 {
        return fmt.Errorf("Missing controller's ID")
    }

    // Check if it existed    
    query := "SELECT name FROM controllers WHERE id = ?"
    var query_result string
    DB.QueryRow(query, c.ID).Scan(&query_result)
    if query_result == "" { return fmt.Errorf("Controller does not exists") }

    // Delete the controller
    _, err := DB.Exec(`DELETE FROM controllers WHERE id = ?`, c.ID)
    if err != nil { return err }

    // Delete all the related devices
    _, err = DB.Exec(`DELETE FROM devices WHERE cid = ?`, c.ID)
    if err != nil { return err }

    // If no error have been returned return nil
    return nil
}

func (c *Controller) Check() {
    // Build the controller's url
    url := fmt.Sprintf("http://%s:%d/ping", c.IP, c.Port)

    // If the server can't connect to it, set it as sleeping
    if _, err := http.Get(url); err != nil {
        c.Sleeping = true
        DB.Exec(`UPDATE controllers SET sleeping=true WHERE id = ?`, c.ID)
    }
}


type Device struct {
    ID       int64    `json:"id"`
    CID      int64    `json:"controller_id"`
    GPIO     *int64   `json:"GPIO"`
    Name     string   `json:"name"`
    Status   bool     `json:"status"`
}

func (d *Device) Save() (error) {
    // Check if controller exists
    var name string
    query := "SELECT name FROM controllers WHERE id = ?"
    err := DB.QueryRow(query, d.CID).Scan(&name)
    if err != nil { return fmt.Errorf("Controller does not exists") }

    // Check if GPIO is already used
    query = "SELECT name FROM devices WHERE cid = ? AND gpio = ?"
    err = DB.QueryRow(query, d.CID, d.GPIO).Scan(&name)
    if err == nil { return fmt.Errorf("Device's GPIO is already used") }

    // Insert device in the database
    result, err := DB.Exec(`INSERT INTO devices (cid, gpio, name, status) VALUES (?, ?, ?, ?)`, d.CID, d.GPIO, d.Name, d.Status)
    if err != nil { return err }

    // Set device's id
    d.ID, _ = result.LastInsertId()

    // Return nil if no error have been encountered
    return nil
}


func Initialize_DB() (error) {
    // Connect to the database
    var err error
    DB, err = sql.Open("sqlite3", "database.db")

    // Look for errors
    if err != nil { return err }
    if err := DB.Ping(); err != nil { return err }

    // Create devices table if it doesn't exist
    statement, _ := DB.Prepare(`CREATE TABLE IF NOT EXISTS devices (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                cid INTEGER NOT NULL,
                                gpio INTEGER NOT NULL,
                                name STRING UNIQUE NOT NULL,
                                status BOOLEAN NOT NULL);`)
    statement.Exec()

    // Create controllers table if it doesn't exist
    statement, _ = DB.Prepare(`CREATE TABLE IF NOT EXISTS controllers (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                ip STRING NOT NULL,
                                mac STRING UNIQUE NOT NULL,
                                port INTEGER NOT NULL,
                                name STRING UNIQUE NOT NULL,
                                sleeping BOOLEAN);`)
    statement.Exec()

    return nil
}
