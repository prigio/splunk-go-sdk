package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultHost   = "localhost"
	DefaultPort   = 8089
	DefaultScheme = "https"
)

const (
	//pathDeploymentServers = "deployment/server/"
	//path_inputs = "data/inputs/"
	//pathModular_inputs = "data/modular-inputs"
	pathUsers   = "authentication/users/"
	httpTimeout = 5 * time.Second
)

const (
	SplunkSharingUser   = "user"
	SplunkSharingApp    = "app"
	SplunkSharingSystem = "system"
	SplunkSharingGlobal = "global"
)

type NameSpace struct {
	owner   string
	app     string
	sharing string
}

// GetNamespace instantiates a new Splunk namespace
func GetNamespace(owner, app, sharing string) (*NameSpace, error) {
	ns := &NameSpace{}
	if err := ns.set(owner, app, sharing); err != nil {
		return nil, err
	}
	return ns, nil
}

func (ns *NameSpace) getServicesNSUrl() string {
	return "/servicesNS/" + ns.owner + "/" + ns.app + "/"
}

func (ns *NameSpace) set(owner, app, sharing string) error {
	ns.owner = owner
	ns.app = app
	ns.sharing = sharing
	return nil
}

type SplunkService struct {
	baseUrl    string
	debug      bool
	authUser   string
	nameSpace  NameSpace
	authToken  string
	httpClient *http.Client
	sessionKey string
}

func NewSplunkServiceWithUsernameAndPassword(serviceUrl string, username, password, passcode2FA string, insecureSkipVerify bool) (*SplunkService, error) {
	var resp *http.Response
	var err error

	if serviceUrl == "" || (!strings.HasPrefix(serviceUrl, "https://") && !strings.HasPrefix(serviceUrl, "http://")) {
		return nil, fmt.Errorf("splunk service login: invalid splunk service URL provided; must be in format http(s)://host:port")
	}
	if username == "" {
		return nil, fmt.Errorf("splunk service login: username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("splunk service login: password cannot be empty")
	}

	ns, _ := GetNamespace("nobody", "search", SplunkSharingApp)

	ss := &SplunkService{
		authUser:  username,
		nameSpace: *ns,
		baseUrl:   strings.TrimRight(serviceUrl, "/"),
	}

	ss.initHttpClient(insecureSkipVerify)

	loginParams := url.Values{}
	loginParams.Set("username", username)
	loginParams.Set("password", password)
	if passcode2FA != "" {
		// MFA token value
		loginParams.Set("passcode", passcode2FA)
	}

	// Submit login form
	if resp, err = ss.httpClient.PostForm(ss.buildUrl(pathLogin), loginParams); err != nil {
		return nil, fmt.Errorf("splunk service login: %s", err.Error())
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// HTTP 401
		// 	{"messages":[{"type":"WARN","code":"incorrect_username_or_password","text":"Login failed"}]}
		return nil, fmt.Errorf("splunk service login: HTTP %v - %s", resp.StatusCode, respBody)
	}

	lr := LoginResponse{}
	if err = json.Unmarshal(respBody, &lr); err != nil {
		return nil, fmt.Errorf("splunk service login: %s", err.Error())
	}

	// All is fine, store session key
	// HTTP 200
	// {"sessionKey":"FKPT2miFNvbSStAl68_IywfGIMQSN5hre...","message":"","code":""}
	ss.sessionKey = lr.SessionKey
	return ss, nil
}

func NewSplunkServiceWithSessionKey(serviceUrl string, sessionKey string, insecureSkipVerify bool) (*SplunkService, error) {
	var err error

	if serviceUrl == "" || (!strings.HasPrefix(serviceUrl, "https://") && !strings.HasPrefix(serviceUrl, "http://")) {
		return nil, fmt.Errorf("splunk service session-key login: Invalid splunk service URL provided. Must be in format http(s)://host:port")
	}

	ns, _ := GetNamespace("nobody", "search", SplunkSharingApp)

	ss := &SplunkService{
		sessionKey: sessionKey,
		nameSpace:  *ns,
		baseUrl:    strings.TrimRight(serviceUrl, "/"),
	}
	ss.initHttpClient(insecureSkipVerify)

	_, err = ss.Info()
	if err != nil {
		return nil, fmt.Errorf("splunk service session-key login: initialization failed. %s", err.Error())
	}

	return ss, nil
}

func NewSplunkServiceWithToken(serviceUrl string, token string, insecureSkipVerify bool) (*SplunkService, error) {
	var err error

	if serviceUrl == "" || (!strings.HasPrefix(serviceUrl, "https://") && !strings.HasPrefix(serviceUrl, "http://")) {
		return nil, fmt.Errorf("splunk service token login: Invalid splunk service URL provided. Must be in format http(s)://host:port")
	}

	ns, _ := GetNamespace("nobody", "search", SplunkSharingApp)

	ss := &SplunkService{
		authToken: token,
		nameSpace: *ns,
		baseUrl:   strings.TrimRight(serviceUrl, "/"),
	}
	ss.initHttpClient(insecureSkipVerify)

	_, err = ss.Info()
	if err != nil {
		return nil, fmt.Errorf("splunk service token login: initialization failed. %s", err.Error())
	}

	return ss, nil
}

func (ss *SplunkService) GetSessionKey() string {
	return ss.sessionKey
}

// initHttpClient configures the httpClient internap variable. This must be called after creation of a new SplunkService
func (ss *SplunkService) initHttpClient(insecureSkipVerify bool) {
	httpTransport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
		// Configure TLS certificate verification
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}

	//if hrc.Proxy != nil {
	//	httpTransport.Proxy = http.ProxyURL(hrc.Proxy)
	//	}

	ss.httpClient = &http.Client{
		Transport: httpTransport,
		Timeout:   httpTimeout,
	}
}

func (ss *SplunkService) buildUrl(urlPath string) string {
	if strings.HasPrefix(urlPath, "/services") {
		return ss.baseUrl + urlPath + "?output_mode=json"
	}
	if strings.HasPrefix(urlPath, "/") {
		return ss.baseUrl + urlPath + "?output_mode=json"
	}
	return ss.baseUrl + "/" + urlPath + "?output_mode=json"
}

func (ss *SplunkService) buildUrlWithParams(urlPath string, urlParams *url.Values) string {
	if urlParams == nil {
		urlParams = &url.Values{}
	}
	urlParams.Add("output_mode", "json")
	if strings.HasPrefix(urlPath, "/services") {
		return ss.baseUrl + urlPath + "?" + urlParams.Encode()
	}
	if strings.HasPrefix(urlPath, "/") {
		return ss.baseUrl + urlPath + "?" + urlParams.Encode()
	}
	return ss.baseUrl + "/" + urlPath + "?" + urlParams.Encode()
}

/*
// newHttpRequest returns a new http.Request object configured according to the specific needs (headers etc	)
func (ss *SplunkService) newHttpRequest(method, urlPath string, body io.Reader) (*http.Request, error) {
	var r *http.Request
	var err error

	r, err = http.NewRequest(method, ss.buildUrl(urlPath), body)
	if err != nil {
		return nil, err
	}
	// type Header map[string][]string
	r.Header["Authorization"] = append(r.Header["Authorization"], "Splunk "+ss.sessionKey)
	return r, nil
}
*/

// doHttpRequest executes the specified request and returns http code, the body contents and possibly an error
func (ss *SplunkService) doHttpRequest(method, urlPath string, urlParams *url.Values, body io.Reader) (httpCode int, respBody []byte, err error) {
	var fullUrl string
	var req *http.Request
	var resp *http.Response

	if urlParams == nil {
		fullUrl = ss.buildUrl(urlPath)
	} else {
		fullUrl = ss.buildUrlWithParams(urlPath, urlParams)
	}
	req, err = http.NewRequest(method, fullUrl, body)
	if err != nil {
		return 0, nil, err
	}
	// type Header map[string][]string
	// https://docs.splunk.com/Documentation/Splunk/8.1.3/Security/UseAuthTokens
	if ss.sessionKey != "" {
		req.Header["Authorization"] = append(req.Header["Authorization"], "Splunk "+ss.sessionKey)
	} else if ss.authToken != "" {
		req.Header["Authorization"] = append(req.Header["Authorization"], "Bearer "+ss.authToken)
	}

	if resp, err = ss.httpClient.Do(req); err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ = ioutil.ReadAll(resp.Body)

	return resp.StatusCode, respBody, nil
}

//func (ss *SplunkService) getCollection(method, urlPath string, body io.Reader) (httpCode int, respBody []byte, err error) {

// Info retrieves generic information about the Splunk instance the client is connected to
func (ss *SplunkService) Info() (*InfoResource, error) {
	var httpCode int
	var respBody []byte
	var err error

	httpCode, respBody, err = ss.doHttpRequest("GET", pathInfo, nil, nil)

	if err != nil {
		return nil, fmt.Errorf("splunk service info: %s", err.Error())
	}

	if httpCode >= 400 {
		// HTTP 401
		// {"messages":[{"type":"WARN","text":"call not properly authenticated"}]}%
		return nil, fmt.Errorf("splunk service info: HTTP %v - %s", httpCode, string(respBody))
	}

	ir := apiCollection[InfoResource]{}

	if err := json.Unmarshal(respBody, &ir); err != nil {
		return nil, fmt.Errorf("splunk service info: %s", err.Error())
	}

	return &ir.Entry[0].Content, nil
}

// setNameSpace updates the NameSpace configurations for the session
func (ss *SplunkService) setNameSpace(owner, app, sharing string) error {
	return ss.nameSpace.set(owner, app, sharing)
}

func (ss *SplunkService) getCredential(user, realm string) (CredentialResource, AccessControlList, error) {
	var fullUrl string
	var httpCode int
	var respBody []byte
	var err error

	//ss.doHttpRequest("GET", ss.nameSpace.getServicesNSUrl()+pathStoragePasswords, body)
	if realm != "" {
		fullUrl = fmt.Sprintf("%s/%s:%s:", pathStoragePasswords, realm, user)
	} else {
		fullUrl = fmt.Sprintf("%s/%s", pathStoragePasswords, user)
	}
	httpCode, respBody, err = ss.doHttpRequest("GET", fullUrl, nil, nil)
	if err != nil {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("splunk service credential: %s", err.Error())
	}
	if httpCode >= 400 {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("splunk service credential: HTTP %v - %s", httpCode, respBody)
	}

	cred := apiCollection[CredentialResource]{}

	if err := json.Unmarshal(respBody, &cred); err != nil {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("splunk service credential: %s", err.Error())
	}
	if cred.Paging.Total == 0 {
		return CredentialResource{}, AccessControlList{}, nil
	}

	return cred.Entry[0].Content, cred.Entry[0].ACL, nil
}

/*
func (ss *SplunkService) setCredential(user, realm, password string) (Credential, error) {
	var fullUrl string
	var resp *http.Response
	//var httpCode int
	//var respBody []byte
	var err error
	credParams := url.Values{}

	_, err = ss.getCredential(user, realm)
	if err != nil {
		// no credential present, let's create one
		// Submit credentials form
		credParams.Set("name", user)
		credParams.Set("password", password)
		if realm != "" {
			credParams.Set("realm", realm)
		}
		if resp, err = ss.httpClient.PostForm(ss.buildUrl(pathStoragePasswords), credParams); err != nil {
			return Credential{}, fmt.Errorf("splunk service setCredential: %s", err.Error())
		} else {
			fmt.Println(resp.StatusCode)
		}
	} else {
		// credential IS present, need to update it
		if realm != "" {
			fullUrl = fmt.Sprintf("%s/%s:%s:", pathStoragePasswords, realm, user)
		} else {
			fullUrl = fmt.Sprintf("%s/%s", pathStoragePasswords, user)
		}

		credParams.Set("password", password)

		if _, err = ss.httpClient.PostForm(ss.buildUrl(fullUrl), credParams); err != nil {
			return Credential{}, fmt.Errorf("splunk service setCredential: %s", err.Error())
		}
	}
	return Credential{}, nil
}
*/
