package splunkd

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

func (ss *Client) NewLogger(name string, flag int, index, host, source, sourcetype string) *log.Logger {
	if index == "" {
		index = "_internal"
	}
	if host == "" {
		host, _ = os.Hostname()
	}
	if source == "" {
		source = "splunk-go-sdk"
	}
	if sourcetype == "" {
		sourcetype = "logger:" + name
	}

	params := url.Values{}
	params.Add("index", index)
	// os.Hostname can still return an error...
	if host != "" {
		params.Add("host", host)
	}
	params.Add("source", source)
	params.Add("sourcetype", sourcetype)

	l := splunkdLogger{
		name:          name,
		splunkd:       ss,
		loggingParams: params,
	}
	prefix := ""
	if name != "" {
		prefix = fmt.Sprintf("[%s] ", name)
	}

	return log.New(l, prefix, flag)
}

// splunkdLogger is an object which can be used as a writer for GO's "log" package
// it will write logs directly into Splunk.
// This is NOT thought for high-volume logging (use HEC for that), but just a simple way for tools built
// using this SDK to provide their own internal logging.
type splunkdLogger struct {
	name          string
	loggingParams url.Values
	splunkd       *Client
}

// Write implements the io.Writer interface, necessary for the Logger struct to be usable as an output channel for GO's log module.
// Write writes len(p) bytes from p to the underlying data stream. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early. Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
// Ref: https://pkg.go.dev/io#Writer
func (l splunkdLogger) Write(p []byte) (n int, err error) {
	lr := logResult{}
	if err := doSplunkdHttpRequest(l.splunkd, "POST", "services/receivers/simple", &l.loggingParams, p, "", &lr); err != nil {
		return 0, fmt.Errorf("splunk service logger[%s]: %w", l.name, err)
	}
	if lr.Bytes < len(p) {
		return lr.Bytes, fmt.Errorf("splunk service logger[%s]: %d bytes written instead of %d", l.name, lr.Bytes, len(p))
	}
	return lr.Bytes, nil
}

// LogResult represents the data returned by the SDK's Log() method, which calls splunkd's /services/receivers/simple endpoint
// This structure can be used to confirm where splunkd is going to store the indexed event
type logResult struct {
	// {"index":"default","bytes":70,"host":"172.17.0.1","source":"www","sourcetype":"web_event"}%
	Index      string `json:"index"`
	Host       string `json:"host"`
	Source     string `json:"source"`
	Sourcetype string `json:"sourcetype"`
	Bytes      int    `json:"bytes"`
}
