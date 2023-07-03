package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

func (ss *SplunkService) Login(username, password, passcode2FA string) error {
	if username == "" {
		return fmt.Errorf("login: username cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("login: password cannot be empty")
	}

	var resp *http.Response
	var err error

	loginParams := url.Values{}
	loginParams.Set("username", username)
	loginParams.Set("password", password)
	if passcode2FA != "" {
		// MFA token value
		loginParams.Set("passcode", passcode2FA)
	}

	// Submit login form
	if resp, err = ss.httpClient.PostForm(buildSplunkdUrl(ss.baseUrl, pathLogin, nil), loginParams); err != nil {
		return fmt.Errorf("splunkd login: %w", err)
	}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// HTTP 401
		// 	{"messages":[{"type":"WARN","code":"incorrect_username_or_password","text":"Login failed"}]}
		return fmt.Errorf("splunkd login: HTTP %v - %s", resp.StatusCode, respBody)
	}

	lr := LoginResponse{}
	if err = json.Unmarshal(respBody, &lr); err != nil {
		return fmt.Errorf("splunkd login: %w", err)
	}

	// All is fine, store session key
	// HTTP 200
	// {"sessionKey":"FKPT2.......","message":"","code":""}
	ss.sessionKey = lr.SessionKey

	// retrieve authentication context information
	ss.AuthContext()

	return nil
}

func (ss *SplunkService) LoginWithToken(authToken string) error {
	ss.authToken = authToken

	if _, err := ss.AuthContext(); err != nil {
		return fmt.Errorf("splunkd token login: %w", err)
	}

	return nil
}

func (ss *SplunkService) LoginWithSessionKey(sessionKey string) error {
	ss.sessionKey = sessionKey
	if _, err := ss.AuthContext(); err != nil {
		return fmt.Errorf("splunkd sessionKey login: %w", err)
	}
	return nil
}
