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
	aa.splunkdlogger = aa.splunkd.NewLogger("runId="+aa.runID, 0, "_internal", "", fmt.Sprintf("Alert [%s] %s", aa.GetApp(), aa.GetSearchName()), aa.getLoggingSourcetype())
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

	if !aa.isAtTerminal && aa.splunkdlogger != nil {
		aa.splunkdlogger.Printf(message, a...)
	} else {
		fmt.Fprintf(os.Stderr, message, a...)
	}
}

// RegisterEndUserLogger configures logging to report to the end-user the results of the alert execution.
// Messages will be logged into the specified index and can have a custom prefix added to them.
//
// This method can only be executed after the run-time has been initialized, namely after [Run] as read runtime configurations from STDIN.
// This method panics if executed before proper initialization.
func (aa *AlertAction) RegisterEndUserLogger(index, messagePrefix string) error {
	if aa.splunkd == nil {
		panic("registerEndUserLogger: no runtime config available. This operation must be performed when a runtime config is available")
	}
	if index == "" {
		return fmt.Errorf("alert action setEndUserLogger: index parameter cannot be emtpy")
	}
	sourcetype := "alertaction:" + aa.StanzaName
	source := fmt.Sprintf("Alert [%s] %s", aa.GetApp(), aa.GetSearchName())
	// initialize a logger to perform logging visible by the end user
	aa.Log("INFO", `Will be logging results of execution for the end-user as index="%s" sourcetype="%s" source="%s"`, index, sourcetype, source)
	aa.endUserLogger = aa.splunkd.NewLogger(messagePrefix, 0, index, "", source, sourcetype)
	return nil
}

// LogForEndUser writes a log to an index visible for the end-user of the alert in order to report on
// the alert execution.
// It is MANDATORY to first initialize the logger using [RegisterEndUserLogger].This method panic if that was not done.
//
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (aa *AlertAction) LogForEndUser(level string, message string, a ...interface{}) {
	if aa.endUserLogger == nil {
		panic("logForEndUser: logger available. Use RegisterEndUserLogger to initialize a logger when a runtime config is available")
	}
	level = strings.ToUpper(level)
	message = fmt.Sprintf("%s %s - %s\n",
		time.Now().Round(time.Millisecond).Format("2006-01-02T15:04:05.000-0700"),
		level,
		message)

	aa.endUserLogger.Printf(message, a...)
}

// LogForEndUserIfEnabled writes a log to an index visible for the end-user of the alert in order to report on
// the alert execution.
// In case [RegisterEndUserLogger] had not been used to configure the logging output, this method silently does nothing.
//
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (aa *AlertAction) LogForEndUserIfEnabled(level string, message string, a ...interface{}) {
	if aa.endUserLogger == nil {
		return
	}
	level = strings.ToUpper(level)
	message = fmt.Sprintf("%s %s - %s\n",
		time.Now().Round(time.Millisecond).Format("2006-01-02T15:04:05.000-0700"),
		level,
		message)

	aa.endUserLogger.Printf(message, a...)
}
