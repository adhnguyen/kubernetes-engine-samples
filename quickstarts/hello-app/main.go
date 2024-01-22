/**
 * Copyright 2021 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// [START gke_hello_app]
// [START container_hello_app]
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type WeatherRequest struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ApiKey string  `json:"apiKey"`
}

func main() {
	// register hello function to handle all requests
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", hello)
	// mux.Handle("/weather", http.HandlerFunc(weather)).Methods("POST")
	r := mux.NewRouter()

	r.HandleFunc("/", hello)
	r.HandleFunc("/weather", weather).Methods("POST")

	// use PORT environment variable, or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "6068"
	}

	// start the web server on port and accept requests
	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// hello responds to the request with a plain-text "Hello, world" message.
func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	host, _ := os.Hostname()
	fmt.Fprintf(w, "Hello, world!\n")
	fmt.Fprintf(w, "Version: 1.0.0\n")
	fmt.Fprintf(w, "Hostname: %s\n", host)
}

func weather(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)

	decoder := json.NewDecoder(r.Body)
	var requestData WeatherRequest

	err := decoder.Decode(&requestData)
	if err != nil {
		message := fmt.Sprintf("Error decoding JSON request: %s", err)
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}

	lat := requestData.Lat
	lon := requestData.Lon
	apiKey := requestData.ApiKey

	response, err := http.Get(fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s",
		lat, lon, apiKey))

	if err != nil {
		message := fmt.Sprintf("Failed to call API: %s", err)
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}
	defer response.Body.Close()

	var responseJson map[string]interface{}
	json.NewDecoder(response.Body).Decode(&responseJson)

	if response.StatusCode != 200 {
		message := "Failed to call API."
		if errMsg, ok := responseJson["message"].(string); ok {
			message += " Error: " + errMsg
		}
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}

	summary := responseJson["weather"].([]interface{})[0].(map[string]interface{})["description"].(string)

	// Return JSON response
	jsonResponse := map[string]interface{}{
		"sessionInfo": map[string]interface{}{
			"parameters": map[string]interface{}{
				"summary": summary,
			},
		},
	}

	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		message := fmt.Sprintf("Error encoding JSON response: %s", err)
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// [END container_hello_app]
// [END gke_hello_app]
