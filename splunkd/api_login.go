package splunkd

import (
	"fmt"
	"net/url"

	"github.com/prigio/splunk-go-sdk/v2/errors"
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

func (ss *Client) Login(username, password, passcode2FA string) error {
	if username == "" {
		return errors.NewErrInvalidParam("login", nil, "'username' cannot be empty")
	}
	if password == "" {
		return errors.NewErrInvalidParam("login", nil, "'password' cannot be empty")
	}

	var err error

	loginParams := url.Values{}
	loginParams.Set("username", username)
	loginParams.Set("password", password)
	if passcode2FA != "" {
		// MFA token value
		loginParams.Set("passcode", passcode2FA)
	}

	// Submit login form
	lr := LoginResponse{}
	if err = doSplunkdHttpRequest(ss, "POST", pathLogin, nil, []byte(loginParams.Encode()), "", &lr); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	// All is fine, store session key
	// HTTP 200
	// {"sessionKey":"FKPT2.......","message":"","code":""}
	ss.sessionKey = lr.SessionKey

	// retrieve authentication context information
	ss.AuthContext()

	return nil
}

func (ss *Client) LoginWithToken(authToken string) error {
	if authToken == "" {
		return errors.NewErrInvalidParam("loginWithToken", nil, "'authToken' cannot be empty")
	}
	ss.authToken = authToken
	if _, err := ss.AuthContext(); err != nil {
		return fmt.Errorf("loginWithToken: %w", err)
	}
	return nil
}

func (ss *Client) LoginWithSessionKey(sessionKey string) error {
	if sessionKey == "" {
		return errors.NewErrInvalidParam("loginWithSessionKey", nil, "'sessionKey' cannot be empty")
	}
	ss.sessionKey = sessionKey
	if _, err := ss.AuthContext(); err != nil {
		return fmt.Errorf("loginWithSessionKey: %w", err)
	}
	return nil
}
