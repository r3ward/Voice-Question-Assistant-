package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	ttsREGION = "uksouth"
	ttsURI    = "https://" + ttsREGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
	ttsKEY = "19c1cb3c0aa848608fed5a5a8a23d640"
	first  = "<speak version='1.0' xml:lang='en-US'> <voice xml:lang='en-US' name='en-US-JennyNeural'>"
	last   = "</voice> </speak>"
)

// handleTttsError writes error responses and log errors
func handleTtsError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	//Writes the error message to response
	w.Write([]byte(err.Error()))
	//Log shows error in terminal
	log.Print(err.Error())
}

// textToSpeech receives a request and returns the response
func textToSpeech(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	// Decode the incoming request
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		handleTtsError(w, http.StatusInternalServerError, err)
		return
	}
	// Converts request to string
	if text, ok := t["text"].(string); ok {
		finalText := first + text + last
		response, err := ttsService([]byte(finalText))
		// If error during api request or response
		if err != nil {
			handleTtsError(w, http.StatusInternalServerError, err)
			return
		}
		//returns a base64 encoded speech file
		sEnc := base64.StdEncoding.EncodeToString(response)
		u := map[string]interface{}{"speech": sEnc}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
		return
	}
	// Returns error if text field cannot be converted to string
	handleTtsError(w, http.StatusInternalServerError, errors.New("Cannot serve request"))
}

// Sends and recieves
func ttsService(text []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", ttsURI, bytes.NewBuffer(text))
	if err != nil {
		return nil, errors.New("Error with the Text-to-Speach request")
	}
	// Prepares the request
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", ttsKEY)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	// Sends the request
	rsp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Error sending request to Text-to-Speech Service")
	}
	defer rsp.Body.Close()
	// If response statusCode = 200
	if rsp.StatusCode == http.StatusOK {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			return body, nil
		}
		// if cannot read the response body
		return nil, errors.New("Cannot read Text-to-Speech Service response")
	}
	// If response statusCode is not 200
	return nil, errors.New("Speech-to-Text service currently unavaliable")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tts", textToSpeech).Methods("POST")
	http.ListenAndServe(":3003", r)
}
