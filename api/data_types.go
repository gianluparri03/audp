package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)


type Device struct {
    ID      int    `json:"id"`
    CID     int    `json:"controller_id"`
    Name    string `json:"name"`
    Status  bool   `json:"status"`
}

type Controller struct {
    ID      int      `json:"id"`
    Url     string   `json:"url"`
    Name    string   `json:"name"`
    Devices []Device `json:"devices"`
}


func Initialize_DB() (*sql.DB, error) {
    // Connect to the database
    db, err := sql.Open("sqlite3", "database.db")

    // Look for errors
    if err != nil { return nil, err }
    if err := db.Ping(); err != nil { return nil, err }

    // Create devices table if it doesn't exist
    statement, _ := db.Prepare(`CREATE TABLE IF NOT EXISTS devices (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                controller_id INTEGER NOT NULL,
                                name STRING UNIQUE NOT NULL,
                                status INTEGER NOT NULL);`)
    statement.Exec()

    // Create controllers table if it doesn't exist
    statement, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS controllers (
                                id INTEGER PRIMARY KEY AUTOINCREMENT,
                                url STRING UNIQUE NOT NULL,
                                name STRING NOT NULL);`)
    statement.Exec()

    return db, nil
}
