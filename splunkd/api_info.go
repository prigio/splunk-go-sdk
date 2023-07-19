package splunkd

import (
	"fmt"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// InfoResource is the structure containing the data returned by the pathInfo URL
type InfoResource struct {
	Version              string `json:"version"`
	Build                string `json:"build"`
	CpuArch              string `json:"cpu_arch"`
	Guid                 string `json:"guid"`
	HealthInfo           string `json:"health_info"`
	ServerName           string `json:"serverName"`
	NumberOfCores        int    `json:"numberOfCores"`
	NumberOfVirtualCores int    `json:"numberOfVirtualCores"`
	PhysicalMemoryMB     int    `json:"physicalMemoryMB"`
	OsBuild              string `json:"os_build"`
	OsName               string `json:"os_name"`
	OsVersion            string `json:"os_version"`
}

// Info retrieves generic information about the Splunk instance the client is connected to
// It caches such information locally, as this is not something which regularly varies
func (ss *Client) Info() (*InfoResource, error) {
	if ss.info != nil {
		return ss.info, nil
	}

	col := collection[InfoResource]{
		name: "info",
		path: "server/info",
	}

	// pathInfo represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTintrospect#server.2Finfo

	err := doSplunkdHttpRequest(ss, "GET", "/services/server/info", nil, nil, "", &col)
	if err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}

	ss.info = &col.Entries[0].Content
	return ss.info, nil
}
