package main

import (
    "testing"

    "strings"
    "bytes"
    "encoding/json"

    "net/http"
    "net/http/httptest"
    "github.com/gorilla/mux"

    "os"
)



type Endpoint struct {
    Handler   func(w http.ResponseWriter, r *http.Request)

    // Request
    Method     string
    Target     string
    ReqBody    []byte
    ReqVars    map[string]string

    // Expected Response
    ResCode    int
    ResBody    []byte
}

func (e Endpoint) Test(t *testing.T) {
    // Create the request
    req, _ := http.NewRequest(e.Method, e.Target, bytes.NewReader(e.ReqBody))
    req.RemoteAddr = "127.0.0.1"
    req = mux.SetURLVars(req, e.ReqVars)

    // Initialize a recorder and make the request
    rr := httptest.NewRecorder()
    e.Handler(rr, req)

    // Compare status code
    if rr.Code != e.ResCode {
        t.Errorf("%s %s: got %d status code instead of %d", e.Method, e.Target, rr.Code, e.ResCode)
    }

    // Compare responses
    if !CompareResponses(rr.Body, e.ResBody) {
        got := strings.Trim(rr.Body.String(), "\n")
        expected := string(e.ResBody)
        t.Errorf("%s %s: got %v instead of %v", e.Method, e.Target, got, expected)
    }
}

func CompareResponses (gotBytes *bytes.Buffer, expectedBytes []byte) bool {
    // Tranform got from *bytes.Buffer to map[string]{}interface
    var got, expected map[string]interface{}
    json.Unmarshal([]byte(gotBytes.String()), &got)
    json.Unmarshal(expectedBytes, &expected)

    // Check if they're nils and if they're of the same lenght
    if (got == nil) != (expected == nil) || len(got) != len(expected) {
        return false
    }

    // Compare pairs key: value
    for k, v := range expected {
        if got[k] != v { return false }
    }

    return true
}

func TestMain(m *testing.M) {
    InitializeDB("/tmp/audp.db")
    defer os.Remove("/tmp/audp.db")
    m.Run()
}



func TestPing(t *testing.T) {
    // 200: Ok
    t.Run("200.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: Ping,
                             Method: "GET",
                             Target: "/",
                             ResCode: 200,
                             ResBody: []byte(`{"msg": "AUDP APIs working!", "version": "dev"}`)}
        endpoint.Test(t)
    })
}

func TestListControllers(t *testing.T) {
    // 200: Without controllers
    t.Run("200.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: ListControllers,
                             Method: "GET",
                             Target: "/controllers",
                             ResCode: 200, 
                             ResBody: []byte(`[]`)}
        endpoint.Test(t)
    })

    // Add a controller
    DB.Exec(`INSERT INTO controllers (name, ip, port) VALUES ("Raspberry", "192.168.1.3", 8080);`)

    // 200: Some controllers
    t.Run("200.2", func(t *testing.T) {
        endpoint := Endpoint{Handler: ListControllers,
                             Method: "GET",
                             Target: "/controllers",
                             ResCode: 200, 
                             ResBody: []byte(`[{"id": 1, "name": "Test", "ip": "192.168.1.3", "port": 8080, "sleeping": false, "devices": null}]`)}
        endpoint.Test(t)
    })
}

func TestGetController(t *testing.T) {
    // 404: Controller not found
    t.Run("404.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: GetController,
                             Method: "GET",
                             Target: "/controllers/Banana",
                             ReqVars: map[string]string{"name": "Banana"}, 
                             ResCode: 404,
                             ResBody: []byte(`{"error": "controller not found", "description": "Didn't find a controller with that name"}`)}
        endpoint.Test(t)
    })

    // 200: Controller found
    t.Run("200.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: GetController,
                             Method: "GET",
                             Target: "/controllers/Raspberry",
                             ReqVars: map[string]string{"name": "Raspberry"}, 
                             ResCode: 200,
                             ResBody: []byte(`{"id": 1, "ip": "192.168.1.3", "port": 8080, "name": "Raspberry", "devices": null, "sleeping": false}`)}
        endpoint.Test(t)
    })
}

func TestCreateController(t *testing.T) {
    // 201: Created
    t.Run("201.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: CreateController,
                             Method: "POST",
                             Target: "/controllers",
                             ReqBody: []byte(`{"name": "ESP", "port": 8080}`),
                             ResCode: 201,
                             ResBody: []byte(`{"id": 2, "ip": "127.0.0.1", "port": 8080, "name": "ESP", "devices": null, "sleeping": false}`)}
        endpoint.Test(t)
    })

    // 400: Missing controller's name
    t.Run("400.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: CreateController,
                             Method: "POST",
                             Target: "/controllers",
                             ReqBody: []byte(`{"port": 7070}`),
                             ResCode: 400,
                             ResBody: []byte(`{"error": "invalid controller", "description": "Missing controller's name"}`)}
        endpoint.Test(t)
    })

    // 400: Missing controller's port
    t.Run("400.2", func(t *testing.T) {
        endpoint := Endpoint{Handler: CreateController,
                             Method: "POST",
                             Target: "/controllers",
                             ReqBody: []byte(`{"name": "Test"}`),
                             ResCode: 400,
                             ResBody: []byte(`{"error": "invalid controller", "description": "Missing controller's port"}`)}
        endpoint.Test(t)
    })

    // 409: Controller's name already taken
    t.Run("409.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: CreateController,
                             Method: "POST",
                             Target: "/controllers",
                             ReqBody: []byte(`{"name": "Raspberry", "port": 80}`),
                             ResCode: 409,
                             ResBody: []byte(`{"error": "can't save controller", "description": "Controller's name already used"}`)}
        endpoint.Test(t)
    })

    // 409: Controller's IP already taken
    t.Run("409.2", func(t *testing.T) {
        endpoint := Endpoint{Handler: CreateController,
                             Method: "POST",
                             Target: "/controllers",
                             ReqBody: []byte(`{"name": "Test", "port": 80}`),
                             ResCode: 409,
                             ResBody: []byte(`{"error": "can't save controller", "description": "Controller's IP already used"}`)}
        endpoint.Test(t)
    })
}

func TestWakeUpController(t *testing.T) {
    // 404: Controller doesn't exist
    t.Run("404.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: WakeUpController,
                             Method: "PUT",
                             Target: "/controllers/Toast/wakeup/8080",
                             ReqVars: map[string]string{"name": "Banana", "port": "8080"},
                             ResCode: 404,
                             ResBody: []byte(`{"error": "can't wake up controller", "description": "Controller doesn't exist"}`)}
        endpoint.Test(t)
    })

    // 409: Controller isn't sleeping
    t.Run("409.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: WakeUpController,
                             Method: "PUT",
                             Target: "/controllers/Raspberry/wakeup/8080",
                             ReqVars: map[string]string{"name": "Raspberry", "Port": "8080"},
                             ResCode: 409,
                             ResBody: []byte(`{"error": "can't wake up controller", "description": "Controller isn't sleeping"}`)}
        endpoint.Test(t)
    })

    // Set the controller as sleeping
    DB.Exec(`UPDATE controllers SET sleeping=true WHERE name="Raspberry";`)

    // 409: IP already used
    t.Run("409.2", func(t *testing.T) {
        endpoint := Endpoint{Handler: WakeUpController,
                             Method: "PUT",
                             Target: "/controllers/Raspberry/wakeup/8080",
                             ReqVars: map[string]string{"name": "Raspberry", "port": "8080"},
                             ResCode: 409,
                             ResBody: []byte(`{"error": "can't wake up controller", "description": "IP already used"}`)}
        endpoint.Test(t)
    })

    // Change ESP's ip
    DB.Exec(`UPDATE controllers SET ip="192.168.1.9" WHERE name="ESP";`)

    // 200: Woken Up
    t.Run("200.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: WakeUpController,
                             Method: "PUT",
                             Target: "/controllers/Raspberry/wakeup/3030",
                             ReqVars: map[string]string{"name": "Raspberry", "port": "3030"},
                             ResCode: 200,
                             ResBody: []byte(`{"id": 1, "ip": "127.0.0.1", "port": 3030, "name": "Raspberry", "devices": null, "sleeping": false}`)}
        endpoint.Test(t)
    })
}

func TestDeleteController(t *testing.T) {
    // 404: Not Found
    t.Run("404.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: DeleteController,
                             Method: "DELETE",
                             Target: "/controllers/Banana",
                             ReqVars: map[string]string{"name": "Banana"},
                             ResCode: 404,
                             ResBody: []byte(`{"error":"Can't delete controller","description":"Controller doesn't exist"}`)}
        endpoint.Test(t)
    })

    // 204: Done
    t.Run("204.1", func(t *testing.T) {
        endpoint := Endpoint{Handler: DeleteController,
                             Method: "DELETE",
                             Target: "/controllers/Raspberry",
                             ReqVars: map[string]string{"name": "Raspberry"},
                             ResCode: 204,
                             ResBody: []byte(``)}
        endpoint.Test(t)
    })

    // 404: Not Found
    t.Run("404.2", func(t *testing.T) {
        endpoint := Endpoint{Handler: DeleteController,
                             Method: "DELETE",
                             Target: "/controllers/Raspberry",
                             ReqVars: map[string]string{"name": "Raspberry"},
                             ResCode: 404,
                             ResBody: []byte(`{"error":"Can't delete controller","description":"Controller doesn't exist"}`)}
        endpoint.Test(t)
    })
}
