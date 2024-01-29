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
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

type WeatherRequest struct {
	Location string `string:"location"`
	ApiKey   string `json:"apiKey"`
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

	var (
		lat float64
		lon float64
	)

	location := requestData.Location
	apiKey := requestData.ApiKey

	// Retrieve lat, lon for location
	location = url.QueryEscape(location)
	urlString := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", location, apiKey)
	target, err := url.Parse(urlString)
	if err != nil {
		fmt.Fprintf(w, `{"error": "Failed to parse URL: %s"}`, err)
		return
	}

	response, err := http.Get(target.String())
	if err != nil {
		message := fmt.Sprintf("Failed to call API: %s", err)
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		message := "Failed to call API."
		fmt.Fprintf(w, `{"error": "%s"}`, message)
		return
	}
	var responseList []interface{}
	json.NewDecoder(response.Body).Decode(&responseList)

	lat = responseList[0].(map[string]interface{})["lat"].(float64)
	lon = responseList[0].(map[string]interface{})["lon"].(float64)

	response, err = http.Get(fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
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
	temp_min := responseJson["main"].(map[string]interface{})["temp_min"].(float64)
	temp_max := responseJson["main"].(map[string]interface{})["temp_max"].(float64)

	// Return JSON response
	jsonResponse := map[string]interface{}{
		"sessionInfo": map[string]interface{}{
			"parameters": map[string]interface{}{
				"summary":  summary,
				"temp_min": temp_min,
				"temp_max": temp_max,
				"lat":      lat,
				"lon":      lon,
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
