package main

import (
    "fmt"

    "database/sql"
    _ "github.com/mattn/go-sqlite3"

    "net/http"
)


// Global Variables
var DB *sql.DB


// Data Types
type Controller struct {
    ID         int64      `json:"id"`
    IP         string     `json:"ip"`
    Port       uint64     `json:"port"`
    Name       string     `json:"name"`
    Devices    []Device   `json:"devices"`
    Sleeping   bool       `json:"sleeping"`
}

func PingControllers() {
    // Query all the awake controllers
    rows, _ := DB.Query(`SELECT id, ip, port FROM controllers WHERE sleeping=0`)
    defer rows.Close()

    for rows.Next() {
        // Try to connect to them
        var c Controller
        rows.Scan(&c.ID, &c.IP, &c.Port)
        url := fmt.Sprintf("http://%s:%d/", c.IP, c.Port)

        // If the server can't connect to them, set them as sleeping
        if _, err := http.Get(url); err != nil {
            DB.Exec(`UPDATE controllers SET sleeping=true WHERE id=?`, c.ID)
        }
    }
}

type Device struct {
    ID         int64      `json:"id"`
    CID        uint64     `json:"controller_id"`
    GPIO       *uint64    `json:"GPIO"`
    Name       string     `json:"name"`
    Status     bool       `json:"status"`
}


// Database Initialization
func InitializeDB(path string) (error) {
    // Connect to the database
    var err error
    DB, err = sql.Open("sqlite3", path)

    // Look for errors
    if err != nil { return err }
    if err := DB.Ping(); err != nil { return err }

    // Create controllers table if it doesn't exist
    statement, _ := DB.Prepare(`CREATE TABLE IF NOT EXISTS controllers (
                               id INTEGER PRIMARY KEY AUTOINCREMENT,
                               ip STRING UNIQUE NOT NULL,
                               port INTEGER NOT NULL,
                               name STRING UNIQUE NOT NULL,
                               sleeping BOOLEAN);`)
    statement.Exec()

    // Create devices table if it doesn't exist
    statement, _ = DB.Prepare(`CREATE TABLE IF NOT EXISTS devices (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                cid INTEGER NOT NULL,
                                gpio INTEGER NOT NULL,
                                name STRING UNIQUE NOT NULL,
                                status BOOLEAN NOT NULL);`)
    statement.Exec()

    return nil
}
