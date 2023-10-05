package splunkd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/prigio/splunk-go-sdk/v2/errors"
)

// discardBody is used as a "nil" type by doSplunkdHttpRequest() for the parseJSONResultInto argument
// reason is that, being doSplunkdHttpRequest a generic function, if it receives a "nil" argument, the parametric type of the function cannot be determined by the compiler
type discardBody struct{}

// doSplunkdHttpRequest executes the specified request and returns http code, the body contents and possibly an error
func doSplunkdHttpRequest[T any](ss *Client, method, urlPath string, urlParams *url.Values, body []byte, contentType string, parseJSONResultInto *T) (err error) {
	if ss == nil {
		return errors.NewErrInvalidParam("doSplunkdHttpRequest", nil, "'splunkService' cannot be nil")
	}
	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" && method != "DELETE" && method != "PUT" && method != "HEAD" {
		return errors.NewErrInvalidParam("doSplunkdHttpRequest", nil, "'method' must be one of GET, POST, DELETE, PUT, HEAD. provided:'%s'", method)
	}
	if urlPath == "" {
		return errors.NewErrInvalidParam("doSplunkdHttpRequest", nil, "'urlPath' cannot be empty")
	}

	var fullUrl string
	var req *http.Request
	var resp *http.Response
	var bodyReader *bytes.Reader

	if !strings.HasPrefix(ss.baseUrl, "http") {
		ss.baseUrl = "https://" + ss.baseUrl
	}

	if urlParams == nil {
		urlParams = &url.Values{}
	}
	urlParams.Set("output_mode", "json")

	fullUrl, _ = url.JoinPath(ss.baseUrl, urlPath)
	// in some cases, the SDK sends absolute URLs to this function
	if strings.HasPrefix(urlPath, "http") {
		fullUrl = urlPath
	}
	fullUrl = fullUrl + "?" + urlParams.Encode()

	// this also manages case where body is nil or has len=0
	bodyReader = bytes.NewReader(body)

	if req, err = http.NewRequest(method, fullUrl, bodyReader); err != nil {
		return fmt.Errorf("doSplunkdHttpRequest: %w", err)
	}
	if contentType != "" {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
		req.Header.Set("content-type", contentType)
	}

	// type Header map[string][]string
	// https://docs.splunk.com/Documentation/Splunk/8.1.3/Security/UseAuthTokens
	if ss.sessionKey != "" {
		req.Header.Set("Authorization", "Splunk "+ss.sessionKey)
	} else if ss.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ss.authToken)
	}

	//log.Printf("DEBUG [splunk service]: performing HTTP %s %s %s\n", req.Method, req.URL.Path, string(body))
	if resp, err = ss.httpClient.Do(req); err != nil {
		//log.Debug("splunk service: HTTP %s %s: %s", req.Method, req.URL.Path, err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		// HTTP 401
		// {"messages":[{"type":"WARN","text":"call not properly authenticated"}]}%
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		//log.Printf("DEBUG [splunk service]: reply %s %s", resp.Status, respBody)
		return fmt.Errorf("HTTP %s '%s':  %s %s - %s", method, fullUrl, resp.Status, http.StatusText(resp.StatusCode), string(respBody))
	}
	//log.Printf("DBODY: %T\n", parseJSONResultInto)
	if parseJSONResultInto != nil && fmt.Sprintf("%T", parseJSONResultInto) != "*splunkd.discardBody" {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		//log.Printf("DEBUG [splunk service]: reply %s %s", resp.Status, respBody)
		return json.Unmarshal(respBody, parseJSONResultInto)
	}

	return nil
}

func interfaceToBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, _ := strconv.ParseBool(val)
		return b
	case int:
		if val > 0 {
			return true
		}
	}
	return false
}
