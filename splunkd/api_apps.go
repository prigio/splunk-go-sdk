package splunkd

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

type AppResource struct {
	Label                      string `json:"label"`
	Description                string `json:"description"`
	Author                     string `json:"author"`
	Version                    string `json:"username"`
	Configured                 bool   `json:"configured"`
	Core                       bool   `json:"core"`
	StateChangeRequiresRestart bool   `json:"state_change_requires_restart"`
	Disabled                   bool   `json:"disabled"`
	Visible                    bool   `json:"visible"`
	ShowInNav                  bool   `json:"show_in_nav"`
	ManagedByDeploymentClient  bool   `json:"managed_by_deployment_client"`
	CheckForUpdates            bool   `json:"check_for_updates"`
}

type AppsCollection struct {
	collection[AppResource]
}
