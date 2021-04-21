package client

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// pathInfo represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTintrospect#server.2Finfo
const pathInfo = "server/info"

// InfoResponse is the structure containing the data returned by the pathInfo URL
type InfoResponse struct {
	Version              string
	Build                string
	CpuArch              string `json:"cpu_arch"`
	Guid                 string
	HealthInfo           string `json:"health_info"`
	NumberOfCores        int
	NumberOfVirtualCores int
	PhysicalMemoryMB     int    `json:"physicalMemoryMB"`
	OsBuild              string `json:"os_build"`
	OsName               string `json:"os_name"`
	OsVersion            string `json:"os_version"`
}

// pathLogin represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTaccess#auth.2Flogin
const pathLogin = "auth/login"

// LoginResponse is the structure containing the data returned by the pathLogin URL
type LoginResponse struct {
	// LoginResponse manages the results of a login attempt
	// Splunk responds either
	// With HTTP 401
	// 	{"messages":[{"type":"WARN","code":"incorrect_username_or_password","text":"Login failed"}]}
	// or HTTP 200
	//	{"sessionKey":"FKPT2miFNvbSStAl68_IywfGIMQSN5hreU^ss....",
	//	"message":"","code":""}
	SessionKey string
	Message    string
	Code       string
	Messages   []struct {
		Type string
		Code string
		Text string
	}
}

//pathStoragePasswords  represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTaccess#storage.2Fpasswords
const pathStoragePasswords = "storage/passwords/"
