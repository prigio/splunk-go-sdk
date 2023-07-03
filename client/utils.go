package client

import (
	"net/url"
	"strings"
)

func buildSplunkdUrl(baseUrl, urlPath string, urlParams *url.Values) string {
	if baseUrl == "" {
		return ""
	}
	if !strings.HasPrefix(baseUrl, "http") {
		baseUrl = "https://" + baseUrl
	}

	if urlParams == nil {
		urlParams = &url.Values{}
	}
	urlParams.Set("output_mode", "json")

	fullUrl, _ := url.JoinPath(baseUrl, urlPath)

	return fullUrl + "?" + urlParams.Encode()
}
