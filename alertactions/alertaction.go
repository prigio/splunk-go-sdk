package alertactions

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prigio/splunk-go-sdk/client"

	"github.com/mattn/go-isatty"
)

// isAtTerminal is a boolean which is true if the alert action is being executed on a command-line or not.
// this is used to modify the logging format
var isAtTerminal = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

// AlertFunc is the signature of the function used to execute the AlertAction based on configurations provided on STDIN
type AlertFunc func(*AlertAction) error

// ValidateFunc is the signature of the function used to validate the run-time parameters provided to the AlertAction.
type ValidateFunc func(*AlertAction) error

// AlertAction is the main structure defining how an alert action looks like.
// It provides a way for the user to define a Splunk alert action and makes
// standardised methods available.
type AlertAction struct {
	// Name appearing within alert_actions.conf stanza as [stanzaname]
	// Must be lowercase
	StanzaName string
	// Label is displayed on the UI
	Label string
	// Description within the UI
	Description string
	// IconPath is the name of a file within appserver/static/ to be used to represent this alert action
	IconPath string
	// params defines the acceptable parameters for the alert
	params []*Param

	// validateParams is an optional function which can be used to validate the run-time parameters
	validateParams ValidateFunc

	// Execute is a mandatory function used to perform actual alert tasks. This is called by the alert's "Run" method.
	execute AlertFunc

	// This debug setting is meant for facilitating development and is not configurable by a user through splunk's inputs.conf
	debug bool

	// Unique id of this run, generated when starting the "Run" function
	runID string

	// actual run-time configurations provided by Splunk
	runtimeConfig *AlertConfig

	splunkservice *client.SplunkService
}

func New(stanzaName, label, description, iconPath string) (*AlertAction, error) {
	if stanzaName == "" {
		return nil, fmt.Errorf("alertAction.New: 'stanzaName' cannot be empty")
	}

	if label == "" {
		return nil, fmt.Errorf("alertAction.New: 'label' cannot be empty")
	}

	var aa = &AlertAction{
		StanzaName:  stanzaName,
		Label:       label,
		Description: description,
		IconPath:    iconPath,
		runID:       uuid.New().String()[0:8],
	}
	return aa, nil
}

func (aa *AlertAction) EnableDebug() {
	aa.debug = true
}

// AddParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (aa *AlertAction) AddParam(name, title, description, defaultValue, placeholder string, uiType ParamType, required bool) (*Param, error) {
	if name == "" {
		return nil, fmt.Errorf("invalid alert action parameter defined: 'name' cannot be empty")
	}
	if title == "" {
		return nil, fmt.Errorf("invalid alert action parameter defined: 'title' cannot be empty")
	}
	if !(uiType == 0 || uiType == ParamTypeText || uiType == ParamTypeTextArea || uiType == ParamTypeSearchDropdown || uiType == ParamTypeRadio || uiType == ParamTypeDropdown || uiType == ParamTypeColorPicker) {
		return nil, fmt.Errorf("invalid alert action parameter defined: when the uiType should either be 0, or one of the allowed ParamTypes")
	}

	// check if the parameter is already present
	// return error in case it is already there
	if _, err := aa.GetParam(name); err == nil {
		return nil, fmt.Errorf("parameter already existing. It is not possible to add multiple parameters having the same name. name=\"%s\"", name)
	}

	if aa.params == nil {
		aa.params = make([]*Param, 0, 1)
	}
	param := &Param{
		Title:        title,
		Name:         name,
		UIType:       uiType,
		Description:  description,
		Placeholder:  placeholder,
		DefaultValue: defaultValue,
		Required:     required,
	}

	aa.params = append(aa.params, param)
	return param, nil
}

// GetParam searches for the param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *AlertAction) GetParam(name string) (*Param, error) {
	for _, p := range aa.params {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("parameter not found having name=\"%s\"", name)
}

// GetParamNames returns a list of all the parameters defined for the alert action so far
func (aa *AlertAction) GetParamNames() []string {
	var paramsList = make([]string, len(aa.params))
	for i, p := range aa.params {
		paramsList[i] = p.Name
	}
	return paramsList
}

// GetFirstResults returns the first of the search results which the alert has been invoked on.
func (aa *AlertAction) GetFirstResult() map[string]string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetFirstResult invoked without a runtime-configuration having being loaded.")
		return nil
	}
	return aa.runtimeConfig.Result
}

// GetSearchUri returns the URI of the search object on the spluknd service API
func (aa *AlertAction) GetSearchUri() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetSearchUri invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.SearchUri
}

// GetSearchName returns the name of the scheduled search which triggered the alert action
func (aa *AlertAction) GetSearchName() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetSearchName invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.SearchName
}

// TBD: closing the file must be done by the user.
func (aa *AlertAction) GetResultsFile() (*os.File, error) {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetResultsFile invoked without a runtime-configuration having being loaded.")
		return nil, fmt.Errorf("missing runtime-configuration: impossible to locate the 'ResultsFile'")
	}
	return os.Open(aa.runtimeConfig.ResultsFile)
}

func (aa *AlertAction) GetResultsFileReader(f *os.File) (*csv.Reader, error) {
	if f == nil {
		aa.Log("ERROR", "GetResultsFileReader invoked without a proper file pointer")
		return nil, fmt.Errorf("invalid parameter, f is nil")
	}
	gzReader, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return csv.NewReader(gzReader), nil
}

func (aa *AlertAction) GetResultsLink() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetResultsLink invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.ResultsLink
}

// GetSid returns the search id (sid) of the actual execution of the scheduled search which triggered the alert action
func (aa *AlertAction) GetSid() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetSid invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.Sid
}

func (aa *AlertAction) GetNamespace() (owner, app string) {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetNamespace invoked without a runtime-configuration having being loaded.")
		return "", ""
	}
	return aa.runtimeConfig.Owner, aa.runtimeConfig.App
}

func (aa *AlertAction) GetSplunkService() (*client.SplunkService, error) {
	if aa.splunkservice != nil {
		return aa.splunkservice, nil
	}
	if aa.runtimeConfig == nil {
		return nil, fmt.Errorf("alert action config is nil: impossible to use it to retrieve connection information to Splunk")
	}
	// alert actions run locally on splunk servers. It might well be that certificates are self-generated there.
	ss, err := client.New(aa.runtimeConfig.ServerUri, true, "")
	if err != nil {
		return nil, fmt.Errorf("alert action - get splunk service: %w", err)
	}

	if err := ss.LoginWithSessionKey(aa.runtimeConfig.SessionKey); err != nil {
		return nil, fmt.Errorf("alert action - get splunk service: %w", err)
	}

	aa.splunkservice = ss
	return ss, nil
}

// Log writes a log so that it can be read by Splunk.
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (aa *AlertAction) Log(level string, message string, a ...interface{}) (err error) {
	level = strings.ToUpper(level)
	if level == "WARNING" {
		// Typical error, just manage it...
		level = "WARN"
	}
	if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" && level != "FATAL" {
		fmt.Fprintf(os.Stderr, "ERROR - Log() function invoked with invalid level parameter. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'\n", level)
		return fmt.Errorf("invalid value of 'level' provided. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'", level)
	}
	if level != "DEBUG" || (level == "DEBUG" && aa.debug) {
		// do not do anything if debug is not enabled

		message = "[" + aa.StanzaName + "] " + level + " runId=\"" + string(aa.runID) + "\" - " + message + "\n"

		// isAtTerminal is global, set at the beginning of this file, in order to only do this once per execution
		if isAtTerminal {
			// prefix the message with timestamp and log_level
			message = time.Now().Round(time.Millisecond).Format("2006-01-02 15:04:05.000 -0700") + " " + message
		}

		_, err = fmt.Fprintf(os.Stderr, message, a...)
	}

	return err
}

// setConfig stores run-time parameter within the alert action and its parameters.
// Returns an error if any of these actions failed
func (aa *AlertAction) setConfig(c *AlertConfig) error {
	aa.runtimeConfig = c

	// assign the actual value to the parameters
	for _, param := range aa.params {
		if v, found := c.Configuration[param.Name]; found {
			aa.Log("DEBUG", "Setting parameter '%s' to \"%s\"", param.Name, v)
			if err := param.SetValue(v); err != nil {
				return fmt.Errorf("error while applying run-time configuration: %s", err.Error())
			}
		} else {
			aa.Log("DEBUG", "Parameter '%s' uses default value \"%s\"", param.Name, param.GetValue())
		}
	}
	return nil
}

// generateAlertActionsConf returns a string which can be used to define the alert action within the splunk configuration file default/alert_actions.conf
func (aa *AlertAction) generateAlertActionsConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `## Configurations for custom alert action '%s' within default/alert_actions.conf
## These configurations have been auto-generated
## See: https://docs.splunk.com/Documentation/Splunk/latest/Admin/Alertactionsconf

[%s]
label = %s
description = %s
icon_path = %s
is_custom = 1

# This alert action only supports JSON. Do not change this!
payload_format = json

##
## Discretionary settings
##   the developer of the custom aler action might want to set these as needed.
## 

ttl = 14400
# The minimum time to live, in seconds, of the search artifacts, if this action is triggered.
# If 'p' follows '<integer>', then '<integer>' is the number of scheduled periods.
# If no actions are triggered, the ttl for the artifacts are determined
# by the 'dispatch.ttl' setting in the savedsearches.conf file.
# Default: 10p
# Time To Live = 4 Std = 14400 seconds

forceCsvResults = true
# If set to "true", any saved search that includes this action
#  always stores results in CSV format, instead of the internal SRS format.
# If set to "false", results are always serialized using the internal SRS format.
# If set to "auto", results are serialized as CSV if the 'command' setting
# in this stanza starts with "sendalert" or contains the string "$results.file$".
* Default: auto

maxtime = 1h
# The maximum amount of time that the execution of an action is allowed to take before the action is aborted.

#alert.execute.cmd=%s
#alert.execute.cmd.arg.<n> = # Change the command line arguments passed to the script when it is invoked. 

#maxresults = <integer>
#* Set the global maximum number of search results sent through alerts.
#* Default: 100

#maxtime = <integer>[m|s|h|d]
#* The maximum amount of time that the execution of an action is allowed to
#  take before the action is aborted.

##
## Parameters specific for this alert
##   these can be autogenerated by starting the alert from the command line.
##
`, aa.Label, aa.StanzaName, aa.Label, aa.Description, aa.IconPath, os.Args[0])

	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getAlertActionsConf())
	}

	return buf.String()
}

// generateAlertActionsSpec returns a string which can be used to define the alert action within the splunk configuration file README/alert_actions.conf.spec
func (aa *AlertAction) generateAlertActionsSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `** Configurations for custom alert action '%s' within README/alert_actions.conf.spec
** These configurations have been auto-generated
[%s]
`, aa.Label, aa.StanzaName)

	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getAlertActionsSpec())
	}
	return buf.String()
}

// generateSavedSearchesSpec returns a string which can be used to define the alert action within the splunk configuration file README/savedsearches.conf.spec
func (aa *AlertAction) generateSavedSearchesSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `**
** Configurations custom alert action '%s' within README/savedsearches.conf.spec
** These configurations have been auto-generated
**
action.%s = [0|1]
`, aa.Label, aa.StanzaName)

	for _, par := range aa.params {
		fmt.Fprint(buf, par.getSavedSearchesSpec(aa.StanzaName))
	}
	return buf.String()
}

// RegisterValidationFunc configures a function used to validate parameters.
// Basic parameter validation is done automatically. This is needed to check for dependencies across multiple parameters.
// Providing a validation function is optional.
func (aa *AlertAction) RegisterValidationFunc(f ValidateFunc) {
	aa.Log("DEBUG", "Custom parameter validation function registered")
	aa.validateParams = f
}

// RegisterAlertFunc configures the actual alerting function to be executed by the alert action.
// Not providing a function results in a run-time error, as the alert action would not know what to actually do.
func (aa *AlertAction) RegisterAlertFunc(f AlertFunc) {
	aa.Log("DEBUG", "Alerting function registered")
	aa.execute = f
}

// printHelp prints command-line usage instructions to STDOUT
func (aa *AlertAction) printHelp() {
	fmt.Printf("Usage for custom alert action '%s'\n", aa.StanzaName)
	fmt.Printf("Label: %s\n", aa.Label)
	fmt.Printf("Description: %s\n", aa.Description)
	fmt.Println("NOTE: Splunk invokes the alert action using the '--execute' flag, unless differently specified within alert_actions.conf")
	fmt.Println("")
	flag.PrintDefaults()
}

func (aa *AlertAction) Run() error {
	var err error

	// defines the command-line parameters using the 'flag' module
	executePtr := flag.Bool("execute", false, "Starts execution of the alert action. A JSON-based configuration must be provided via STDIN. This is what Splunk does.")
	debugPtr := flag.Bool("debug", false, "Activates debug mode, useful only during development")
	getConfPtr := flag.Bool("get-alert-actions-conf", false, "Print out a template for default/alert_actions.conf")
	getSpecPtr := flag.Bool("get-alert-actions-spec", false, "Print out a template for README/alert_actions.conf.spec")
	getSSSpecPtr := flag.Bool("get-saved-searches-spec", false, "Print out a template for README/savedsearches.conf.spec")
	flag.Parse()

	//if len(os.Args) == 1 {
	//	return fmt.Errorf("invoke the alert with the '--execute' command-line parameter. Exiting")
	//}
	if *debugPtr {
		aa.EnableDebug()
	}

	if *executePtr {
		var runTimeConfig *AlertConfig

		aa.Log("INFO", "Execution started")

		if aa.execute == nil {
			aa.Log("FATAL", "No actual alerting function has been defined")
			return fmt.Errorf("no actual alerting function has been defined")
		}

		aa.Log("DEBUG", "Parsing run-time JSON configurations from STDIN")
		runTimeConfig, err = getAlertConfigFromJSON(os.Stdin)
		if err != nil {
			aa.Log("FATAL", "Parsing of run-time JSON configurations from STDIN failed. %s", err.Error())
			return err
		}

		aa.Log("DEBUG", "Setting run-time configuration: %+v", runTimeConfig)
		err = aa.setConfig(runTimeConfig)
		if err != nil {
			aa.Log("FATAL", "Setting of run-time configurations failed. %s", err.Error())
			return err
		}

		// Note: setConfig() already performs validation of individual parameters.
		// However, sometimes multiple parameters should be analyzed as a group for dependencies between them.
		// The function registered at "validateParams" is supposed to take care of that
		if aa.validateParams != nil {
			aa.Log("INFO", "Validating run-time parameters with registered function")
			err = aa.validateParams(aa)
			if err != nil {
				aa.Log("FATAL", "Validation of run-time parameters failed. %s", err.Error())
				return err
			}
		}
		// At last, perform actual execution of the alerting function
		aa.Log("INFO", "Executing alerting function")
		err = aa.execute(aa)
		if err != nil {
			aa.Log("FATAL", "Execution of alerting function failed. %s", err.Error())
			return err
		}
		aa.Log("INFO", "Execution of alerting function completed")
		return nil
	}

	if *getConfPtr {
		fmt.Println(aa.generateAlertActionsConf())
		return nil
	}

	if *getSpecPtr {
		fmt.Println(aa.generateAlertActionsSpec())
		return nil
	}
	if *getSSSpecPtr {
		fmt.Println(aa.generateSavedSearchesSpec())
		return nil
	}

	// if no valid command-line parameters were provided
	aa.printHelp()
	return nil
}
