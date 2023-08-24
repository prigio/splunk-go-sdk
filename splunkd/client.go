package splunkd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prigio/splunk-go-sdk/utils"
)

const (
	defaultScheme = "https"
	defaultHost   = "localhost"
	defaultPort   = 8089
	httpTimeout   = 10 * time.Second
)

type Client struct {
	// url of splunkd. Generally https://localhost:8089
	baseUrl string
	// used for token-based authentication
	authToken string
	// session key is provided by the login method, or when splunk executes a modular input/alert action
	sessionKey  string
	nameSpace   Namespace
	httpClient  *http.Client
	credentials *CredentialsCollection
	users       *UsersCollection
	kvstore     *KVStoreCollCollection
	// context of the current authenticated session. Provides info about the logged-in username, roles, etc
	authContext *ContextResource
	//configs     map[string]*ConfigsCollection
	// information about the splunk version, server where splunk is deployed, ...
	info *InfoResource
}

func New(splunkdUrl string, insecureSkipVerify bool, proxy string) (*Client, error) {
	if splunkdUrl == "" || (!strings.HasPrefix(splunkdUrl, "https://") && !strings.HasPrefix(splunkdUrl, "http://")) {
		return nil, &utils.ErrInvalidParam{Context: "splunk service new", Msg: "splunkdUrl must have format http(s)://host:port"}
	}
	ns, _ := NewNamespace("nobody", "search", SplunkSharingApp)

	httpClient, err := utils.NewHTTPClient(10*time.Second, insecureSkipVerify, proxy, "", "", "")

	if err != nil {
		return nil, fmt.Errorf("splunk service new: cannot create http client. %w", err)
	}

	if proxy == "" {
		splunkdUrl, err := url.Parse(splunkdUrl)
		if err != nil {
			return nil, &utils.ErrInvalidParam{Context: "splunk service new", Msg: "splunkdUrl", Err: err}
		}
		if err := utils.IsReachable(*splunkdUrl); err != nil {
			return nil, fmt.Errorf("splunk service new: unreachable splunkd URL '%s'. %w", splunkdUrl, err)
		}
	}

	ss := &Client{
		nameSpace:  *ns,
		baseUrl:    strings.TrimRight(splunkdUrl, "/"),
		httpClient: httpClient,
	}

	if proxy != "" {
		// test whether the proxy can connect to the splunk server.
		// this is done here, as we need the httpClient to have been prepared already
		req, err := http.NewRequest(http.MethodHead, ss.baseUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("splunk service new: unreachable splunkd URL '%s' via proxy '%s'. %w", ss.baseUrl, proxy, err)
		}
		if r, err := ss.httpClient.Do(req); err != nil {
			return nil, fmt.Errorf("splunk service new: unreachable splunkd URL '%s' via proxy '%s'. %w", ss.baseUrl, proxy, err)
		} else if r.StatusCode > 200 {
			// a HEAD call to a wrong URL produces 502-Bad Gateway when going through a proxy server.
			return nil, fmt.Errorf("splunk service new: unreachable splunkd URL '%s' via proxy '%s'. HTTP %d - %s", ss.baseUrl, proxy, r.StatusCode, r.Status)
		}
	}

	return ss, nil
}

func NewFromDefaults() (*Client, error) {
	return New(fmt.Sprintf("%s://%s:%d", defaultScheme, defaultHost, defaultPort), true, "")
}

// NewInteractive uses the Params[] definition of an alert action to prepare a configuration based on:
// - command line parameters
// - interactively asking the user if no command-line parameter was found for an argument
func NewInteractive() (*Client, error) {
	// first, need to get splunk endpoint, username and password to be able to login into the service if necessary.
	uri := utils.AskForInput("Splunkd URL", "https://localhost:8089", false)
	username := utils.AskForInput("Splunk username", "admin", false)
	password := utils.AskForInput("Splunk password", "", true)
	ss, err := New(uri, true, "")
	if err != nil {
		return nil, fmt.Errorf("connection failed to splunkd on '%s'. %w", uri, err)
	}
	if err = ss.Login(username, password, ""); err != nil {
		return nil, fmt.Errorf("login failed to splunkd with username '%s': %w", username, err)
	}
	return ss, nil
}

func (ss *Client) GetSessionKey() string {
	return ss.sessionKey
}

func (ss *Client) GetSplunkdURI() string {
	return ss.baseUrl
}

//func (ss *SplunkService) getCollection(method, urlPath string, body io.Reader) (httpCode int, respBody []byte, err error) {

// SetNamespace updates the NameSpace configurations for the session
func (ss *Client) SetNamespace(owner, app string, sharing SplunkSharing) error {
	ns, err := NewNamespace(owner, app, sharing)
	if err != nil {
		return fmt.Errorf("splunk service setNamespace: %w", err)
	}
	ss.nameSpace = *ns
	return nil
}

// SetNamespaceFromNS updates the NameSpace configurations for the session using an existing NameSpace
func (ss *Client) SetNamespaceFromNS(ns Namespace) {
	ss.nameSpace = ns
}

func (ss *Client) GetCredentials() *CredentialsCollection {
	if ss.credentials == nil {
		ss.credentials = NewCredentialsCollection(ss)
	}
	return ss.credentials
}

func (ss *Client) GetUsers() *UsersCollection {
	if ss.users == nil {
		ss.users = NewUsersCollection(ss)
	}
	return ss.users
}

func (ss *Client) GetKVStore() *KVStoreCollCollection {
	if ss.kvstore == nil {
		ss.kvstore = NewKVStoreCollCollection(ss)
	}
	return ss.kvstore
}

//func (ss *Client) GetConfigs(filename string) *ConfigsCollection {
//	return NewConfigsCollection(ss, filename)
//}
/*
func (ss *Client) GetConfigsNS(filename, owner, app string) *ConfigsCollection {
	if ss.configs == nil {
		ss.configs = make(map[string]*ConfigsCollection)
	}
	if filename == "" {
		return nil
	}

	c, ok := ss.configs[filename]
	if ok && c != nil {
		return c
	}
	ss.configs[filename] = NewConfigsCollectionNS(ss, filename, owner, app)
	return ss.configs[filename]
}
*/
