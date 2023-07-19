package client

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

type SplunkService struct {
	// url of splunkd. Generally https://localhost:8089
	baseUrl string
	// used for token-based authentication
	authToken string
	// session key is provided by the login method, or when splunk executes a modular input/alert action
	sessionKey  string
	nameSpace   NameSpace
	httpClient  *http.Client
	credentials *CredentialsCollection
	kvstore     *KVStoreCollCollection
	// context of the current authenticated session. Provides info about the logged-in username, roles, etc
	authContext *ContextResource
	configs     map[string]*ConfigsCollection
	// information about the splunk version, server where splunk is deployed, ...
	info *InfoResource
}

func New(splunkdUrl string, insecureSkipVerify bool, proxy string) (*SplunkService, error) {
	if splunkdUrl == "" || (!strings.HasPrefix(splunkdUrl, "https://") && !strings.HasPrefix(splunkdUrl, "http://")) {
		return nil, fmt.Errorf("splunk service new: invalid service URL provided; must be in format http(s)://host:port")
	}
	ns, _ := NewNamespace("nobody", "search", SplunkSharingApp)

	httpClient, err := utils.NewHTTPClient(10*time.Second, insecureSkipVerify, proxy, "", "", "")

	if err != nil {
		return nil, fmt.Errorf("splunk service new: cannot create http client. %w", err)
	}

	if proxy == "" {
		splunkdUrl, err := url.Parse(splunkdUrl)
		if err != nil {
			return nil, fmt.Errorf("splunk service new: invalid splunkd URL '%s'. %w", splunkdUrl, err)
		}
		if err := utils.IsReachable(*splunkdUrl); err != nil {
			return nil, fmt.Errorf("splunk service new: unreachable splunkd URL '%s'. %w", splunkdUrl, err)
		}
	}

	ss := &SplunkService{
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

func NewFromDefaults() (*SplunkService, error) {
	return New(fmt.Sprintf("%s://%s:%d", defaultScheme, defaultHost, defaultPort), true, "")
}

func (ss *SplunkService) GetSessionKey() string {
	return ss.sessionKey
}

//func (ss *SplunkService) getCollection(method, urlPath string, body io.Reader) (httpCode int, respBody []byte, err error) {

// SetNameSpace updates the NameSpace configurations for the session
func (ss *SplunkService) SetNameSpace(owner, app string, sharing SplunkSharing) error {
	ns, err := NewNamespace(owner, app, sharing)
	if err != nil {
		return fmt.Errorf("splunk service setnamespace: %w", err)
	}
	ss.nameSpace = *ns
	return nil
}

func (ss *SplunkService) GetCredentials() *CredentialsCollection {
	if ss.credentials == nil {
		ss.credentials = NewCredentialsCollection(ss)
	}
	return ss.credentials
}

func (ss *SplunkService) GetKVStore() *KVStoreCollCollection {
	if ss.kvstore == nil {
		ss.kvstore = NewKVStoreCollCollection(ss)
	}
	return ss.kvstore
}

func (ss *SplunkService) GetConfigs(filename string) *ConfigsCollection {
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
	ss.configs[filename] = NewConfigsCollection(ss, filename)
	return ss.configs[filename]
}
