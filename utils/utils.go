package utils

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// NewHTTPClient configures a new HTTP client which can be used to issue requests to external services
func NewHTTPClient(timeout time.Duration, insecureSkipVerify bool, proxy string, tlsCAcerts, tlsClientCert, tlsClientKey string) (*http.Client, error) {
	// initialize the internal http client to communicate with splunkd
	var retries = 3
	var backoff = 30 * time.Second
	httpTransport, err := NewHTTPTransport(timeout, retries, backoff, proxy, insecureSkipVerify, tlsCAcerts, tlsClientCert, tlsClientKey)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: httpTransport, Timeout: timeout}, nil
}

// NewHTTPCNewHTTPTransport configures a new HTTP transport with the provided settings. The transport can be used to instantiate a http client
func NewHTTPTransport(timeout time.Duration, retries int, backoff time.Duration, proxy string, insecureSkipVerify bool, tlsCAcerts, tlsClientCert, tlsClientKey string) (http.RoundTripper, error) {
	// initialize the internal http client to communicate with splunkd
	httpTransport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: timeout,
	}

	tlsConfig := &tls.Config{
		// Configure TLS certificate verification
		InsecureSkipVerify: insecureSkipVerify,
	}

	if tlsClientCert != "" && tlsClientKey != "" {
		clientCert, err := tls.LoadX509KeyPair(tlsClientCert, tlsClientKey)
		if err != nil {
			return nil, fmt.Errorf("error loading tls certificate and key '%s'/'%s'. %w", tlsClientCert, tlsClientKey, err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	if tlsCAcerts != "" {
		caCert, err := os.ReadFile(tlsCAcerts)
		if err != nil {
			return nil, fmt.Errorf("error loading tls CA certificate '%s'. %w", tlsCAcerts, err)
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	}

	httpTransport.TLSClientConfig = tlsConfig

	if proxy != "" {
		// golang also uses the env variable HTTP_PROXY to automatically use proxy
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL '%s'. %w", proxy, err)
		}
		err = IsReachable(*proxyUrl)
		if err != nil {
			return nil, fmt.Errorf("unreachable proxy URL '%s'. %w", proxy, err)
		}
		httpTransport.Proxy = http.ProxyURL(proxyUrl)
	}
	return httpTransport, nil
}

// IsReachable tries to connect to the target URL and returns an error if this is not possible
func IsReachable(target url.URL) error {
	var port = target.Port()
	if target.Port() == "" {
		switch strings.ToLower(target.Scheme) {
		case "http":
			port = "80"
		case "https":
			port = "443"
		case "ftp":
			port = "20"
		case "ldap":
			port = "389"
		case "ldaps":
			port = "636"
		case "ftps":
			port = "990"
		case "ssh":
			port = "22"
		case "smtp":
			port = "25"
		case "pop3":
			port = "110"
		case "pop3s":
			port = "995"
		case "sftp":
			port = "115"
		case "sockes":
			port = "1080"
		case "imaps":
			port = "993"
		default:
			return fmt.Errorf("isReachable: invalid URL - cannot determine correct port. URL=\"%s\"", target.String())
		}
	}
	if conn, err := net.DialTimeout("tcp", target.Hostname()+":"+port, 500*time.Millisecond); err != nil {
		return err
	} else {
		conn.Close()
		return nil
	}
}

/*
	 askForInput promts the user to provide a value via StdIn
		if isPassword=true, no local echo to the console is provided
		if isPassword=false and the provided input is empty, the default value is returned instead
*/
func AskForInput(prompt string, defaultVal string, isPassword bool) string {
	prompt = strings.Trim(prompt, ": ")
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	if isPassword {
		if bytepw, err := term.ReadPassword(int(syscall.Stdin)); err != nil {
			// reading passwords does not move cursor to next line, so have to force it
			fmt.Println("")
			return ""
		} else {
			// reading passwords does not move cursor to next line, so have to force it
			fmt.Println("")
			return string(bytepw)
		}
	} else {
		if text, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil || text == "\n" {
			return defaultVal
		} else {
			return strings.Replace(text, "\n", "", -1)
		}
	}
}

// GetEpochNow returns an the current time as Epoch, expressed in seconds with a decimal part
func GetEpochNow() float64 {
	return float64(time.Now().UnixNano()) / 1000000000.0
}

// GetEpoch returns an Epoch timestamp with millisecond precision starting from time t
func GetEpoch(t time.Time) float64 {
	return float64(t.UnixNano()) / 1000000000.0
}
