package client

import (
	"fmt"
	"time"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API for this endpoint:
// https://splunkd/services/authentication/current-context?output_mode=json

// ContextResource is the structure containing the data returned by the current-contex URL
type ContextResource struct {
	Capabilities   []string `json:"capabilities"`
	Version        string   `json:"version"`
	DefaultApp     string   `json:"defaultApp"`
	Email          string   `json:"email"`
	LockedOut      bool     `json:"locked-out"`
	Realname       string   `json:"realname"`
	Roles          []string `json:"roles"`
	Username       string   `json:"username"`
	LastLoginEpoch int64    `json:"last_successful_login"`
	LastLogin      time.Time
}

// Info retrieves generic information about the Splunk instance the client is connected to
// It caches such information locally, as this is not something which regularly varies
func (ss *SplunkService) AuthContext() (*ContextResource, error) {
	if ss.authContext != nil {
		return ss.authContext, nil
	}

	col := collection[ContextResource]{
		name: "auth-context",
		path: "authentication/current-context",
	}

	// pathInfo represents this enpoint https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTintrospect#server.2Finfo
	httpCode, respBody, err := ss.doHttpRequest("GET", "/services/authentication/current-context", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}

	err = col.parseResponse(httpCode, respBody)
	if err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}

	ss.authContext = &col.Entries[0].Content
	ss.authContext.LastLogin = time.Unix(ss.authContext.LastLoginEpoch, 0)

	return ss.authContext, nil
}

// Can checks whether the logged-in user has the specified capability
func (ss *SplunkService) Can(capability string) (bool, error) {
	if capability == "" {
		return false, fmt.Errorf("can capability: parameter 'capability' cannot be emtpy")
	}
	var cr *ContextResource
	var err error

	cr, err = ss.AuthContext()
	if err != nil {
		return false, err
	}

	for _, usercapab := range cr.Capabilities {
		if capability == usercapab {
			return true, nil
		}
	}
	return false, nil
}

// Has checks whether the logged-in user has the specified role assigned
func (ss *SplunkService) Has(role string) (bool, error) {
	if role == "" {
		return false, fmt.Errorf("has role: parameter 'role' cannot be emtpy")
	}
	var cr *ContextResource
	var err error

	cr, err = ss.AuthContext()
	if err != nil {
		return false, err
	}

	for _, userrole := range cr.Roles {
		if role == userrole {
			return true, nil
		}
	}
	return false, nil
}

func (ss *SplunkService) Username() (string, error) {
	var cr *ContextResource
	var err error

	cr, err = ss.AuthContext()
	if err != nil {
		return "", err
	}
	return cr.Username, nil
}
