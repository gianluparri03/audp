package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)



var DB *sql.DB


type Device struct {
    ID      int64  `json:"id"`
    CID     int64  `json:"controller_id"`
    Name    string `json:"name"`
    Status  bool   `json:"status"`
}

type Controller struct {
    ID      int64    `json:"id"`
    URL     string   `json:"url"`
    Name    string   `json:"name"`
    Devices []Device `json:"devices"`
}


func (d *Device) Save() (error) {
    // Check if controller exists
    query := "SELECT name FROM controllers WHERE id = ?"
    var name string
    err := DB.QueryRow(query, d.CID).Scan(&name)
    if err != nil { return err }

    // Insert device in the database
    statement, _ := DB.Prepare(`INSERT INTO devices (cid, name, status) VALUES (?, ?, ?)`)
    result, err := statement.Exec(d.CID, d.Name, d.Status)

    // Check for errors
    if err != nil { return err }

    // Set device's id
    d.ID, _ = result.LastInsertId()

    // Return nil if no error have been encountered
    return nil
}

func (c *Controller) Save() (error) {
    // Insert controller in the database
    statement, _ := DB.Prepare(`INSERT INTO controllers (url, name) VALUES (?, ?)`)
    result, err := statement.Exec(c.URL, c.Name)

    // Check for errors
    if err != nil { return err }

    // Set controller's id
    c.ID, _ = result.LastInsertId()

    // Return nil if no error have been encountered
    return nil
}

func (c *Controller) SaveAll() (error) {
    // Save controller
    err := c.Save()
    if err != nil { return err }

    // Save every device
    for i, d := range c.Devices {
        d.CID = c.ID
        err = d.Save()
        if err != nil { return err }
        c.Devices[i] = d
    }

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
                                name STRING UNIQUE NOT NULL,
                                status INTEGER NOT NULL);`)
    statement.Exec()

    // Create controllers table if it doesn't exist
    statement, _ = DB.Prepare(`CREATE TABLE IF NOT EXISTS controllers (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                url STRING UNIQUE NOT NULL,
                                name STRING NOT NULL);`)
    statement.Exec()

    return nil
}
