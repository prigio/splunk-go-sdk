package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	// context of the current authenticated session. Provides info about the logged-in username, roles, etc
	authContext *ContextResource
	// information about the splunk version, server where splunk is deployed, ...
	info *InfoResource
}

func isReachable(target url.URL) error {
	if conn, err := net.DialTimeout("tcp", target.Host, 500*time.Millisecond); err != nil {
		return err
	} else {
		conn.Close()
	}
	return nil
}

func New(serviceUrl string, insecureSkipVerify bool, proxy string) (*SplunkService, error) {
	if serviceUrl == "" || (!strings.HasPrefix(serviceUrl, "https://") && !strings.HasPrefix(serviceUrl, "http://")) {
		return nil, fmt.Errorf("splunk service new: invalid service URL provided; must be in format http(s)://host:port")
	}
	ns, _ := NewNamespace("nobody", "search", SplunkSharingApp)

	ss := &SplunkService{
		nameSpace: *ns,
		baseUrl:   strings.TrimRight(serviceUrl, "/"),
	}

	// initialize the internal http client to communicate with splunkd
	httpTransport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
		// Configure TLS certificate verification
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}

	if proxy != "" {
		// golang also uses the env variable HTTP_PROXY to automatically use proxy
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL '%s'. %w", proxy, err)
		}
		err = isReachable(*proxyUrl)
		if err != nil {
			return nil, fmt.Errorf("unreachable proxy URL '%s'. %w", proxy, err)
		}
		httpTransport.Proxy = http.ProxyURL(proxyUrl)
	} else {
		splunkdUrl, err := url.Parse(ss.baseUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid splunkd URL '%s'. %w", ss.baseUrl, err)
		}
		if err := isReachable(*splunkdUrl); err != nil {
			return nil, fmt.Errorf("unreachable splunkd URL '%s'. %w", ss.baseUrl, err)
		}
	}

	ss.httpClient = &http.Client{
		Transport: httpTransport,
		Timeout:   httpTimeout,
	}

	if proxy != "" {
		// test whether the proxy can connect to the splunk server.
		// this is done here, as we need the httpClient to have been prepared already
		req, err := http.NewRequest(http.MethodHead, ss.baseUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("unreachable splunkd URL '%s' via proxy '%s'. %w", ss.baseUrl, proxy, err)
		}
		if r, err := ss.httpClient.Do(req); err != nil {
			return nil, fmt.Errorf("unreachable splunkd URL '%s' via proxy '%s'. %w", ss.baseUrl, proxy, err)
		} else if r.StatusCode > 200 {
			// a HEAD call to a wrong URL produces 502-Bad Gateway when going through a proxy server.
			return nil, fmt.Errorf("unreachable splunkd URL '%s' via proxy '%s'. HTTP %d - %s", ss.baseUrl, proxy, r.StatusCode, r.Status)
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

	fullUrl = buildSplunkdUrl(ss.baseUrl, urlPath, urlParams)

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

	//log.Printf("DEBUG [splunk service]: performing HTTP %s %s", req.Method, req.URL.Path)
	if resp, err = ss.httpClient.Do(req); err != nil {
		//log.Debug("splunk service: HTTP %s %s: %s", req.Method, req.URL.Path, err.Error())
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ = ioutil.ReadAll(resp.Body)

	//log.Debug("splunk service: HTTP %s %s: %d", req.Method, req.URL.Path, resp.StatusCode)
	return resp.StatusCode, respBody, nil
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
