package alertactions

/*
This file contains utility methods for the AlertAction struct to deal with logging
*/
import (
	"fmt"
	"os"
	"strings"
	"time"
)

// getLoggingSourcetype returns a string indicating the sourcetype used for the administrative logging within index=_internal
func (aa *AlertAction) getLoggingSourcetype() string {
	return "alertaction:" + aa.StanzaName
}

// registerLogger configures the splunkd-based logger
// A runtime configuration must be already available when performing this method.
func (aa *AlertAction) registerLogger() error {
	if aa.splunkdlogger != nil {
		// already available
		return nil
	}
	if aa.splunkd == nil {
		return fmt.Errorf("alert action setLogger: no splunkd client available. This operation must be performed when a runtime config is available")
	}

	// initialize a logger to perform internal logging
	aa.splunkdlogger = aa.splunkd.NewLogger("runId:"+aa.runID, 0, "_internal", "", fmt.Sprintf("Alert [%s] %s", aa.GetApp(), aa.GetSearchName()), aa.getLoggingSourcetype())
	return nil
}

// Log writes a log so that it can be read by Splunk.
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (aa *AlertAction) Log(level string, message string, a ...interface{}) {
	level = strings.ToUpper(level)
	if level == "DEBUG" && !aa.debug {
		// do not do anything if debug is not enabled
		return
	}
	if level == "WARNING" {
		// Typical error, just manage it...
		level = "WARN"
	}
	if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" && level != "FATAL" {
		level = "INFO"
	}

	message = fmt.Sprintf("%s [%s] %s - %s\n",
		time.Now().Round(time.Millisecond).Format("2006-01-02T15:04:05.000-0700"),
		aa.StanzaName,
		level,
		message)

	// isAtTerminal is global, set at the beginning of this file, in order to only do this once per execution
	if !isAtTerminal && aa.splunkdlogger != nil {
		aa.splunkdlogger.Printf(message, a...)
	} else {
		fmt.Fprintf(os.Stderr, message, a...)
	}
}

// RegisterEndUserLogger configures logging to report to the end-user the results of the alert execution.
// Messages will be logged into the specified index and can have a custom prefix added to them.
func (aa *AlertAction) RegisterEndUserLogger(index, messagePrefix string) error {
	if aa.splunkd == nil {
		// already available
		return fmt.Errorf("alert action setEndUserLogger: no splunkd client available. This operation must be performed when a runtime config is available")
	}
	if index == "" {
		return fmt.Errorf("alert action setEndUserLogger: index parameter cannot be emtpy")
	}
	// initialize a logger to perform logging visible by the end user
	aa.endUserLogger = aa.splunkd.NewLogger(messagePrefix, 0, index, "", fmt.Sprintf("Alert [%s] %s", aa.GetApp(), aa.GetSearchName()), "alertaction:"+aa.StanzaName)
	return nil
}

// Log writes a log so that it can be read by Splunk.
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (aa *AlertAction) LogForEndUser(level string, message string, a ...interface{}) {
	level = strings.ToUpper(level)
	message = fmt.Sprintf("%s %s - %s\n",
		time.Now().Round(time.Millisecond).Format("2006-01-02T15:04:05.000-0700"),
		level,
		message)

	// isAtTerminal is global, set at the beginning of this file, in order to only do this once per execution
	if isAtTerminal {
		fmt.Fprintf(os.Stderr, message, a...)
	} else if aa.endUserLogger != nil {
		aa.endUserLogger.Printf(message, a...)
	} else {
		// log an internal message if it was not possible to log as needed.
		aa.Log("ERROR", "Alert action requested logging for end-user, but this has not been initialized. Original message: %s\n", message)
	}
}
