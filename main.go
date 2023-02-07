package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Content struct {
	RefType string `json:"ref_type"`
	RefName string `json:"ref_name"`
	Type    string `json:"type"`
}

type Body struct {
	Target Content `json:"target"`
}

// const serviceURl = "https://deelay.me/2000/http://pruebas.historiaenlinea.com.co/api/healthy"
func main() {
	errors := 0
	var delayMinutes int = 2
	err := godotenv.Load()

	if err != nil {
		log.Fatalln("Error loading .env file")
		return
	}

	// load environment variables
	serviceUrl := os.Getenv("SERVICE_URL")
	bearer := os.Getenv("BEARER")

	for true {
		time.Sleep(time.Duration(delayMinutes) * time.Second)

		// Validate if historiaenlinea is running
		isRunning := validateService(serviceUrl)

		if !isRunning {
			errors += 1
		}

		fmt.Println("errors", errors)

		if errors > 2 {
			triggerPipeline(bearer)
			// reset errors count
			errors = 0
			// while the server build and deploy the application there is around 35 min
			// this time is importar because without it, we could entry to a cicle
			delayMinutes = 35
			return
		}

		// restore delay
		delayMinutes = 2
	}
}

func triggerPipeline(bearer string) bool {

	data := Body{
		Target: Content{
			RefType: "branch",
			Type:    "pipeline_ref_target",
			RefName: "master",
		},
	}

	// body := []byte(`{
	// 	"target": {
	// 		"ref_type": "branch",
	// 		"type":    "pipeline_ref_target",
	// 		"ref_name": "master"
	// 	}
	// }`)

	dataJson, _ := json.Marshal(data)
	// Documentation about bitbucket-api
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/#api-repositories-workspace-repo-slug-pipelines-post

	req, e := http.NewRequest(
		"POST",
		"https://api.bitbucket.org/2.0/repositories/SaludElectronica/his/pipelines/",
		bytes.NewBuffer(dataJson),
	)
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for key, val := range via[0].Header {
			req.Header[key] = val
		}
		return e
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("we have a errro", err)
		return false
	}

	defer resp.Body.Close()

	return true
}

func validateService(serviceURl string) bool {
	// Control http request
	transport := &http.Transport{
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(serviceURl)

	if err != nil {
		println("ERROR->validateService", err)
		return false
	}

	defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return false
	}

	return true
}
