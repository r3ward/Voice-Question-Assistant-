package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

const (
	sttRegion = "uksouth"
	sttURI    = "https://" + sttRegion + ".stt.speech.microsoft.com/" +
		"speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
	sttKEY = "19c1cb3c0aa848608fed5a5a8a23d640"
)

// handleSttError writes error responses and log errors
func handleSttError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	//Writes the error message to response
	w.Write([]byte(err.Error()))
	//Log shows error in terminal
	log.Print(err.Error())
}

// speechToText receives a request and returns the response
func speechToText(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	// Decode the incoming request
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		handleSttError(w, http.StatusInternalServerError, err)
		return
	}
	// Converts request to string
	if text, ok := t["speech"].(string); ok {
		sDec, _ := base64.StdEncoding.DecodeString(text)
		response, err := sttService(sDec)
		// If error during api request or response
		if err != nil {
			handleSttError(w, http.StatusInternalServerError, err)
			return
		}
		// returns the content of speech in text form
		u := map[string]interface{}{"text": response}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
		return
	}
	// Returns error if speech field cannot be converted to string
	handleSttError(w, http.StatusInternalServerError, errors.New("Cannot serve request"))
}

// Extract the needed value from api response body
func extractValue(body string, key string) string {
	keystr := "\"" + key + "\":[^,;\\]}]*"
	r, _ := regexp.Compile(keystr)
	match := r.FindString(body)
	keyValMatch := strings.Split(match, ":")
	return strings.ReplaceAll(keyValMatch[1], "\"", "")
}

// sttService deals with the speech-to-text api
func sttService(speech []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", sttURI, bytes.NewBuffer(speech))
	if err != nil {
		return "", errors.New("Cannot send Speech-to-Text request")
	}
	// Prepares the request
	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", sttKEY)
	// Sends the request
	rsp, err := client.Do(req)
	if err != nil {
		return "", errors.New("Error request to Text-to-Speech Services")
	}
	defer rsp.Body.Close()
	// If response statusCode = 200
	if rsp.StatusCode == http.StatusOK {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			value := extractValue(string(body), "DisplayText")
			return string(value), nil
		}
		// if cannot read the response body
		return "", errors.New("Cannot read response from Speech-to-Text Service")

	}
	// If response statusCode is not 200
	return "", errors.New("Speech-to-Text service currently unavaliable")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/stt", speechToText).Methods("POST")
	http.ListenAndServe(":3002", r)
}
