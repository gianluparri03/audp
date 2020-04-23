#include <ESP8266WiFi.h>
#include <WiFiClient.h>
#include <ESP8266WebServer.h>
#include <ESP8266HTTPClient.h>


// Configs
String WIFI_SSID;
String WIFI_PASSWORD;
String AUDP_SERVER;
int PORT;

// Initialize server and http
ESP8266WebServer server(PORT);
HTTPClient http;


void setup(){
    // Initialize Serial console
    Serial.begin(115200); 
    delay(5000);

    // Connect to wifi and register into the AUDP server
    connectToWifi();
    registerIntoAUDP();

    // Add routes to server
    server.on("/", [](){
        server.send(200, "text/html", "pong!");
    });

    // Start the server
    server.begin();
    Serial.println("HTTP server started");
}

void loop(){
    server.handleClient();
}

void connectToWifi() {
    // Set wifi credentials
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    Serial.println("");

    // Wait for connection
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }
    
    // Log the connection
    Serial.println("");
    Serial.print("Connected to Wifi");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
    Serial.print("MAC address: ");
    Serial.println(WiFi.macAddress());
}

void registerIntoAUDP() {
    // Prepare the request at /controllers/wakeup
    http.begin(AUDP_SERVER + "/controllers/wakeup");
    http.addHeader("Content-Type", "application/json");
    String request_body = "{\"mac_address\": \"" + WiFi.macAddress() + "\"}";

    // Make it
    int response_code = http.POST(request_body);
    String response_content = http.getString();
    http.end();

    // If the response code is 200 stop the function right here
    if (response_code == 200) {
        return;

    // If there isn't a controller with that MAC address register a new one
    } else if (response_code == 400 && response_content == "There isn't a controller with that MAC address\n") {
        // Prepare a new request at /controllers/add
        http.begin(AUDP_SERVER + "/controllers/add");
        http.addHeader("Content-Type", "application/json");
        String request_body = "{\"name\": \"" + WiFi.macAddress() + "\", \"mac_address\": \"" + WiFi.macAddress() + "\", \"port\": " + PORT + "}";

        // Make it
        int response_code = http.POST(request_body);
        String response_content = http.getString();
        http.end();

        // Check if it worked
        if (response_code == 200) {
            return;
        } else {
            Serial.print("Error while registering into the AUDP server: ");
            Serial.println(response_content);
        }

    // otherwise return an error
    } else {
        Serial.print("Error while connecting to the AUDP api: ");
        Serial.println(response_content);
    }
}
