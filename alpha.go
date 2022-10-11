package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

const (
	uri   = "http://api.wolframalpha.com/v1/result"
	appID = "JWEH6G-WYQPJ739RG"
)

// Function to write error responses and log errors
func handleAlphaError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	//Writes the error message to response
	w.Write([]byte(err.Error()))
	//Log shows error in terminal
	log.Print(err.Error())
}

// alpha receives and request and returns a response
func alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	// Decodes the request
	err := json.NewDecoder(r.Body).Decode(&t)
	// Error if cannot be decoded
	if err != nil {
		handleAlphaError(w, http.StatusInternalServerError, err)
		return
	}
	// Converts request to string
	if text, ok := t["text"].(string); ok {
		response, err := alphaService(text)
		// If error during request or response to microservice
		if err != nil {
			handleAlphaError(w, http.StatusInternalServerError, err)
			return
		}
		// if no errors, returns response
		u := map[string]interface{}{"text": response}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
		return
	}
	// Returns error if text field cannot be converted to string
	handleAlphaError(w, http.StatusInternalServerError, errors.New("Body type of request must be of type 'string'"))
}

// alphaService deals with the wolframealpha api
func alphaService(text string) (interface{}, error) {
	url := uri + "?appid=" + appID + "&i=" + url.QueryEscape(text)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Error sending request to Alpha Service")
	}
	// if response code correct
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		// If cannot read response body
		if err != nil {
			return nil, errors.New("Cannot read the response from AlphaAPI")
		}
		return string(body), nil
	}
	// Arrives here if response status is not 200
	return nil, errors.New("Alpha service currently unavaliable")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/alpha", alpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
