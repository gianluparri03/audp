package main

import (
    "encoding/json"
    "net/http"
)


type Response struct {
    Msg     string      `json:"msg"`
    Data    interface{} `json:"data,omitempty"`
}


func Middleware(next http.Handler) http.Handler {
    // Set application/json as content type for all the routes
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}


func Error404() http.Handler {
    // Returns a {"msg": "not found"} with a 404 status code
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")

        response, _ := json.Marshal(Response{Msg: "not found"})
        http.Error(w, string(response), 404)
    })
}

func Error405() http.Handler {
    // Returns a {"msg": "method not allowed"} with a 405 status code
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")

        response, _ := json.Marshal(Response{Msg: "method not allowed"})
        http.Error(w, string(response), 405)
    })
}


func Ping(w http.ResponseWriter, r *http.Request) {
    // Return a simple {"msg": "working"}
    response := Response{Msg: "working"}
    json.NewEncoder(w).Encode(response)
}

func ControllersList(w http.ResponseWriter, r *http.Request) {
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

    // Create response
    response := Response{Msg: "ok"}
    if controllers != nil {
        response.Data = controllers
    } else {
        response.Data = []Controller{}
    }

    // Write it
    json.NewEncoder(w).Encode(response)
}
