package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	stt   = "http://localhost:3002/stt"
	alpha = "http://localhost:3001/alpha"
	tts   = "http://localhost:3003/tts"
)

// Function to write error responses and log errors
func handleAlexaError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	//Writes the error message to response
	w.Write([]byte(err.Error()))
	//Log shows error in terminal
	log.Print(err.Error())
}

// AlexaService deals sends a request to a microservice and returns a response to
func alexaService(request *http.Request, service string) (io.Reader, error) {
	client := &http.Client{}
	// Creates a request
	if response, err := client.Do(request); err == nil {
		// Return the response body if http status is 200
		if response.StatusCode == http.StatusOK {
			return response.Body, nil
		}
		// Status code to be logged the error
		statusString := strconv.Itoa(response.StatusCode)
		return nil, errors.New("Error with " + service + " (Status code: " + statusString + ")")
	}
	// Arrives here if an error with the request
	return nil, errors.New("Unable to send request to: " + service)
}

// requestHandler sends request and returns responses
func requestHandler(w http.ResponseWriter, apiURL string, service string, request io.Reader) (io.Reader, error) {
	req, err := http.NewRequest("POST", apiURL, request)
	// If request error, prepare header and log
	if err != nil {
		handleAlexaError(w, http.StatusInternalServerError, err)
		return nil, err
	}
	res, err := alexaService(req, service)
	// If response error, prepare header and log
	if err != nil {
		handleAlexaError(w, http.StatusInternalServerError, err)
		return nil, err
	}
	// If no errors, will return response
	return res, err
}

// Alexa passes request and recieves responses from all microservices
func alexa(w http.ResponseWriter, r *http.Request) {
	//Assuming the request sent is correct: (r.Body type: base64Encoded wav)
	input, err := requestHandler(w, stt, "Speech-to-text Service", r.Body)
	// If error occured within microservice
	if err != nil {
		return
	}
	// Pass previous input into Alpha
	input, err = requestHandler(w, alpha, "Alpha Service", input)
	if err != nil {
		return
	}
	// Pass previous input into TTS
	input, err = requestHandler(w, tts, "Text-to-Speech Service", input)
	if err != nil {
		return
	}
	// Decodes response from TTS
	response := map[string]interface{}{}
	// If cannot decode
	if err := json.NewDecoder(input).Decode(&response); err != nil {
		handleAlexaError(w, http.StatusInternalServerError, err)
		return
	}
	// Returns response if no issues
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/alexa", alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
}
