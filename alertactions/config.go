package alertactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
This file contains structs and utilities to read the JSON-based configuration
provided by Splunk to an alert action via STDIN
*/

/* This is the most generic form of the JSON coming from splunkd via stdinput
   {
       "app": "search",
       "owner": "admin",
       "results_file": "raw_results.csv.gz",
       "results_link": "ignored",
       "search_uri": "/servicesNS/nobody/search/saved/searches/Test+alert+CSV+raw",
       "server_host": "MacBook.local",
       "server_uri": "https://127.0.0.1:8089",
       "session_key": "ignored",
       "sid": "SID-Manual-Test-CSV-raw-results",
       "search_name": "Test alert CSV raw",
       "configuration": {
           "description": "some description provided by the user",
           "token": "some-auth-token",
           "endpoint": "https://host.docker.internal:7001/post-csv-raw",
           "url_params": "k1=v1, k2=v2, k3=v3",
           "format": "CSV",
           "max_count": "2000",
           "debug": "Y"
       },
       "result": {
           "_time": "1574860920"
       }
   }
*/

// AlertConfig represents the parsed JSON which Splunk provides
// on STDIN when starting the execution of an alert action
type AlertConfig struct {
	App         string `json:"app"`
	Owner       string `json:"owner"`
	ResultsFile string `json:"results_file"`
	ResultsLink string `json:"results_link"`
	SearchUri   string `json:"search_uri"`
	ServerHost  string `json:"server_host"`
	ServerUri   string `json:"server_uri"`
	SessionKey  string `json:"session_key"`
	Sid         string `json:"sid"`
	SearchName  string `json:"search_name"`
	// Configuration is the collection of actual parameters provided by the user when invoking this alert action
	Configuration map[string]string `json:"configuration"`
	Result        map[string]string `json:"result"`
}

// getAlertConfigFromJSON reads a JSON-formatted configuration from the provided Reader,
// parses it and loads it within an alertConfig data structure
func getAlertConfigFromJSON(input io.Reader) (*AlertConfig, error) {
	if input == nil {
		input = os.Stdin
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(input); err != nil {
		return nil, fmt.Errorf("getAlertConfigFromJSON: %s", err.Error())
	}
	// parse and load the XML data within the inputConfig data structure
	ac := &AlertConfig{}
	if err := json.Unmarshal(buf.Bytes(), ac); err != nil {
		return nil, fmt.Errorf("getAlertConfigFromJSON: error when parsing input configuration json. %s. %s", err.Error(), strings.ReplaceAll(buf.String(), "\n", " "))
	}
	return ac, nil
}

/*
// getAlertConfigInteractive uses the Params[] definition of an alert action to prepare a configuration based on:
// - command line parameters
// - interactively asking the user if no command-line parameter was found for an argument
func getAlertConfigInteractive(mi *ModularInput) (*alertConfig, error) {
	// first, need to get splunk endpoint, username and password to be able to login into the service if necessary.
	ic := &inputConfig{}
	ic.CheckpointDir = filepath.Join(os.TempDir(), mi.runID)
	fmt.Printf("CheckPointDir set to '%s'", ic.CheckpointDir)

	fmt.Println("Interactively provide information to access local splunkd service.")
	ic.URI = askForInput("Splunkd URL", "https://localhost:8089", false)
	username := askForInput("Splunk username", "admin", false)
	password := askForInput("Splunk password", "", true)

	ss, err := client.NewSplunkServiceWithUsernameAndPassword(ic.URI, username, password, "", true)
	if err != nil {
		return nil, fmt.Errorf("connection failed to splunkd on '%s' with username '%s': %s", ic.URI, username, err.Error())
	}
	ic.SessionKey = ss.GetSessionKey()

	// Stanzas hosts the configurations provided to the modular input
	ic.Stanzas = make([]Stanza, 1)
	stanza := Stanza{Name: "interactive-input"}
	stanza.Params = make([]Param, len(mi.Args))

	fmt.Println("Interactively provide values for modular input parameters.")
	var prompt, val string
	for seq, arg := range mi.Args {
		prompt = fmt.Sprintf("Provide parameter %s (%s, '%s')", arg.Title, arg.DataType, arg.Name)
		if arg.Description != "" {
			prompt = fmt.Sprintf("%s\n    %s\n", prompt, arg.Description)
		}
		val = askForInput(prompt, arg.DefaultValue, false)
		stanza.Params[seq] = Param{Name: arg.Name, Value: val}
	}

	ic.Stanzas[0] = stanza

	return ic, nil
}
*/
