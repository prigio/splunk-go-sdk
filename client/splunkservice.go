package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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

func NewSplunkServiceWithUsernameAndPassword(host string, port uint32, username, password, passcode string, insecureSkipVerify bool) (*SplunkService, error) {
	if (port != 443 && port != 80) && (port <= 1024 || port > 65535) {
		return nil, fmt.Errorf("Invalid port number provided. Must be either 80, 443 or between 1025 and 65535")
	}

	httpTransport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	if insecureSkipVerify {
		// Disable TLS certificate verification
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	//if hrc.Proxy != nil {
	//	httpTransport.Proxy = http.ProxyURL(hrc.Proxy)
	//	}

	ns, _ := GetNamespace("nobody", "search", SplunkSharingApp)

	ss := &SplunkService{
		authUser:  username,
		nameSpace: *ns,
	}

	if port <= 0 {
		ss.baseUrl = "https://" + host + ":" + strconv.FormatUint(uint64(DefaultPort), 10)
	} else if port != 443 {
		ss.baseUrl = "https://" + host + ":" + strconv.FormatUint(uint64(port), 10)
	} else {
		ss.baseUrl = "https://" + host
	}

	ss.initHttpClient(insecureSkipVerify)

	loginParams := url.Values{}
	loginParams.Set("username", username)
	loginParams.Set("password", password)
	if passcode != "" {
		// MFA token value
		loginParams.Set("passcode", passcode)
	}

	// Submit login form
	if resp, err := ss.httpClient.PostForm(ss.baseUrl+"/services/"+pathLogin+"?output_mode=json", loginParams); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		lr := LoginResponse{}
		respBody, _ := ioutil.ReadAll(resp.Body)
		if err = json.Unmarshal(respBody, &lr); err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			// HTTP 401
			// 	{"messages":[{"type":"WARN","code":"incorrect_username_or_password","text":"Login failed"}]}
			if lr.Messages != nil && len(lr.Messages) > 0 {
				return nil, fmt.Errorf("%s - reason=%s", lr.Messages[0].Text, lr.Messages[0].Code)
			} else {
				return nil, fmt.Errorf("%s - reason=%s", lr.Message, lr.Code)
			}
		} else {
			// All is fine, store session key
			// HTTP 200
			// {"sessionKey":"FKPT2miFNvbSStAl68_IywfGIMQSN5hre...","message":"","code":""}
			ss.sessionKey = lr.SessionKey
		}
	}
	return ss, nil
}

// initHttpClient configures the httpClient internap variable. This must be called after creation of a new SplunkService
func (ss *SplunkService) initHttpClient(insecureSkipVerify bool) {
	httpTransport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	if insecureSkipVerify {
		// Disable TLS certificate verification
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	ss.httpClient = &http.Client{
		Transport: httpTransport,
		Timeout:   httpTimeout,
	}
}

// newHttpRequest returns a new http.Request object configured according to the specific needs (headers etc	)
func (ss *SplunkService) newHttpRequest(method, url string, body io.Reader) (*http.Request, error) {
	var r *http.Request
	var err error
	if url[0] == '/' {
		r, err = http.NewRequest(method, ss.baseUrl+url+"?output_mode=json", body)
	} else {
		r, err = http.NewRequest(method, ss.baseUrl+"/services/"+url+"?output_mode=json", body)
	}
	if err != nil {
		return nil, err
	}
	// type Header map[string][]string
	r.Header["Authorization"] = append(r.Header["Authorization"], "Splunk "+ss.sessionKey)
	return r, nil
}

// Info retrieves generic information about the Splunk instance the client is connected to
func (ss *SplunkService) Info() (*InfoResponse, error) {
	r, err := ss.newHttpRequest("GET", pathInfo, nil)
	if err != nil {
		return nil, err
	}
	if resp, err := ss.httpClient.Do(r); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		ir := struct {
			Entry []struct {
				Content InfoResponse
			}
		}{}

		respBody, _ := ioutil.ReadAll(resp.Body)
		//decd := json.NewDecoder(resp.Body)
		// decd.decode(&ir)
		if err := json.Unmarshal(respBody, &ir); err != nil {
			return nil, err
		}
		return &ir.Entry[0].Content, nil

	}
}

// setNameSpace updates the NameSpace configurations for the session
func (ss *SplunkService) setNameSpace(owner, app, sharing string) error {
	return ss.nameSpace.set(owner, app, sharing)
}

func (ss *SplunkService) getStoragePassword(user, realm string) {
	//ss.newHttpRequest("GET", ss.nameSpace.getServicesNSUrl()+pathStoragePasswords, body)
}
