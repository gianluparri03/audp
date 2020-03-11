package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)


type Device struct {
    id              int    `json:"id"`
    controller_id   int    `json:"controller_id"`
    name            string `json:"name"`
    status          bool   `json:"status"`
}

type Controller struct {
    id      int     `json:"id"`
    url     string  `json:"url"`
    name    string  `json:"name"`
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
