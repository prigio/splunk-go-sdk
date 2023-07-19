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
    "results_file": "/Applications/Splunk/var/run/splunk/dispatch/scheduler__admin__search__RMD5e7649a8a42798738_at_1689609720_2/results.csv.gz",
    "results_link": "http://MacBook-Pro.local:8000/app/search/@go?sid=scheduler__admin__search__RMD5e7649a8a42798738_at_1689609720_2",
    "search_uri": "/servicesNS/admin/search/saved/searches/jira+test",
    "server_host": "MacBook-Pro.local",
    "server_uri": "https://127.0.0.1:8089",
    "session_key": "onrmI0ToM9zNp5BLDl",
    "sid": "scheduler__admin__search__RMD5e7649a8a42798738_at_1689609720_2",
    "search_name": "jira test",
    "configuration": {
        "comment": "thanks",
        "issue_key": "PJW-2306",
        "target_status": "In Progress",
        "track_alert_into_index": "main"
    },
    "result": {
        "_bkt": "_internal~9~0132F894-2D30-4834-A898-90FF4468F923",
        "_cd": "9:114711",
        "_eventtype_color": "",
        "_indextime": "1689609697",
        "_kv": "1",
        "_raw": "07-17-2023 18:01:36.996 +0200 INFO  Metrics - group=searchscheduler, eligible=0, delayed=0, dispatched=0, skipped=0, skipped_dma_ra=0, total_lag=0, max_lag=0, window_max_lag=0, window_total_lag=0, scheduler_cycle_max=0.000, scheduler_cycle_total=0.000, scheduler_cycles=0, load_search_cycle_max=0.000, load_search_cycle_total=0.000, schedule_search_cycle_max=0.000, schedule_search_cycle_total=0.000, load_saved_searches=0.000, load_auto_summarized=0.000, load_dma=0.000, load_deferred_calculation=0.000, load_priority_calculation=0.000, load_cycles=0, disp_get_next_runtime=0.000, disp_determine_can_run=0.000, disp_window_lag_calc=0.000, disp_auto_summary_updates=0.000, disp_advance_next_runtime=0.000, disp_update_logging=0.000, disp_dispatch_job=0.000, max_running=0, actions_triggered=0, completed=0, total_runtime=0.000, max_runtime=0.000, alert_cycle_max=0.000, alert_cycle_total=0.000, alert_cycles=3",
        "_serial": "0",
        "_si": [
            "Paolos-MacBook-Pro.local",
            "_internal"
        ],
        "_sourcetype": "splunkd",
        "_subsecond": ".996",
        "_time": "1689609696.996",
        "date_hour": "18",
        "date_mday": "17",
        "date_minute": "1",
        "date_month": "july",
        "date_second": "36",
        "date_wday": "monday",
        "date_year": "2023",
        "date_zone": "120",
        "eventtype": "",
        "host": "Paolos-MacBook-Pro.local",
        "index": "_internal",
        "linecount": "1",
        "punct": "--_::._+____-_=,_=,_=,_=,_=,_=,_=,_=,_=,_=,_=.,_=.",
        "source": "/Applications/Splunk/var/log/splunk/metrics.log",
        "sourcetype": "splunkd",
        "splunk_server": "Paolos-MacBook-Pro.local",
        "splunk_server_group": "",

		"any other fields identified during the search": ""
}
*/

// alertConfig represents the parsed JSON which Splunk provides
// on STDIN when starting the execution of an alert action
type alertConfig struct {
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
	Configuration map[string]string      `json:"configuration"`
	Result        map[string]interface{} `json:"result"`
}

// getAlertConfigFromJSON reads a JSON-formatted configuration from the provided Reader,
// parses it and loads it within an alertConfig data structure
func getAlertConfigFromJSON(input io.Reader) (*alertConfig, error) {
	if input == nil {
		input = os.Stdin
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(input); err != nil {
		return nil, fmt.Errorf("getAlertConfigFromJSON: %s", err.Error())
	}
	// parse and load the XML data within the inputConfig data structure
	ac := &alertConfig{}
	if err := json.Unmarshal(buf.Bytes(), ac); err != nil {
		return nil, fmt.Errorf("getAlertConfigFromJSON: error when parsing input configuration json. %s. %s", err.Error(), strings.ReplaceAll(buf.String(), "\n", " "))
	}
	return ac, nil
}
