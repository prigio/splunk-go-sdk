package client

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// pathInfo represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTintrospect#server.2Finfo
const pathInfo = "/services/server/info"

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

// pathLogin represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTaccess#auth.2Flogin
const pathLogin = "/services/auth/login"

// LoginResponse is the structure containing the data returned by the pathLogin URL
type LoginResponse struct {
	// LoginResponse manages the results of a login attempt
	// Splunk responds either
	// With HTTP 401
	// 	{"messages":[{"type":"WARN","code":"incorrect_username_or_password","text":"Login failed"}]}
	// or HTTP 200
	//	{"sessionKey":"FKPT2miFNvbSStAl68_IywfGIMQSN5hreU^ss....",
	//	"message":"","code":""}
	SessionKey string `json:"sessionKey"`
	Message    string `json:"message"`
	Code       string `json:"code"`
}

// pathStoragePasswords  represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTaccess#storage.2Fpasswords
const pathStoragePasswords = "/services/storage/passwords/"
const pathStoragePasswordsNS = "/servicesNS/%s/%s/storage/passwords/"

type CredentialResource struct {
	Realm         string `json:"realm"`
	Username      string `json:"username"`
	ClearPassword string `json:"clear_password"`
	EncrPassword  string `json:"encr_password"`
}
