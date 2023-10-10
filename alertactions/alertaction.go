package alertactions

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-isatty"

	"github.com/prigio/splunk-go-sdk/v2/errors"
	"github.com/prigio/splunk-go-sdk/v2/params"
	"github.com/prigio/splunk-go-sdk/v2/splunkd"
)

// AlertingFunc is the signature required for the functions responsible for:
//
//  1. Validate the run-time parameters provided to the AlertAction.
//  2. Execute the actual AlertAction based on configurations provided on STDIN
type AlertingFunc func(*AlertAction) error

// AlertAction is the main structure defining how an alert action looks like.
// It provides a way for the user to define a Splunk alert action and makes
// standardised methods available.
type AlertAction struct {
	// Name appearing within alert_actions.conf stanza as [stanzaname]
	// Must be lowercase, not contain spaces and possibly with words separated by a dash '-'
	StanzaName string
	// Label for the alert, displayed on the UI
	Label string
	// Description within the UI
	Description string
	// Markdown-formatted documentation content about the logic behind the alert execution.
	// This, along with other pre-defined contents can be printed out by starting the alert from the commandline with the propert parameter. See './<alertname> -h'
	Documentation string
	// IconPath is the name of a file within appserver/static/ to be used to represent this alert action
	IconPath string
	// params defines the acceptable parameters for the alert
	// this is a slice and not a map, as this is the easiest way to retain the order the parameters were defined
	params []*params.Param
	// globalParams is used to track the global parameters necessary for the alert.
	// "global", in that they are tracked in a dedicate configuration file and are not configured within the alert UI
	// this is a slice and not a map, as this is the easiest way to retain the order the parameters were defined
	globalParams []*params.Param

	// validateParams is an optional function which can be used to validate the run-time parameters
	validateParams AlertingFunc

	// Execute is a mandatory function used to perform actual alert tasks. This is called by the alert's "Run" method.
	execute AlertingFunc

	// This debug setting is meant for facilitating development and is not configurable by a user through splunk's inputs.conf
	debug bool

	// Unique id of this run, generated when starting the "Run" function
	runID string

	// actual run-time configurations provided by Splunk
	runtimeConfig *alertConfig

	splunkd *splunkd.Client

	// splunkdlogger is used to log message for administrators within index=_internal
	splunkdlogger *log.Logger
	// endUserLogger is used to log messages for the end user in an index preconfigured by them
	endUserLogger *log.Logger

	// isAtTerminal is a boolean which is true if the alert action is being executed on a command-line or not.
	// this is used to modify the logging format
	isAtTerminal bool

	// these are used by the Run() function and are useful for testing.
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func New(stanzaName, label, description, iconPath string) (*AlertAction, error) {
	if stanzaName == "" {
		return nil, errors.NewErrInvalidParam("alertAction.New", nil, "'stanzaName' cannot be empty")
	}
	if label == "" {
		return nil, errors.NewErrInvalidParam("alertAction.New", nil, "'label' cannot be empty")
	}

	var aa = &AlertAction{
		StanzaName:   stanzaName,
		Label:        label,
		Description:  description,
		IconPath:     iconPath,
		runID:        uuid.New().String()[0:8],
		isAtTerminal: isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()),
	}
	return aa, nil
}

func (aa *AlertAction) EnableDebug() {
	aa.debug = true
}

// IsDebug returns true if debug mode has been activated for the alert action
func (aa *AlertAction) IsDebug() bool {
	return aa.debug
}

// RegisterParam adds a given parameter to the alert action.
func (aa *AlertAction) RegisterParam(p *params.Param) error {
	// check if the parameter is already present
	// return error in case it is already there
	if _, err := aa.GetParam(p.GetName()); err == nil {
		return errors.NewErrInvalidParam("registerParam["+p.GetName()+"]", nil, "'%s' already exists", p.GetName())
	}
	if aa.params == nil {
		aa.params = make([]*params.Param, 0, 1)
	}
	aa.params = append(aa.params, p)
	return nil
}

// RegisterNewParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
//
// uiType must be one of "splunk-text-input", "splunk-text-area", "splunk-select", "splunk-radio-input", "splunk-color-picker"
func (aa *AlertAction) RegisterNewParam(name, title, description, defaultValue, placeholder, uiType string, required bool) (*params.Param, error) {
	var p *params.Param
	var err error
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = aa.GetParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam("alert_actions.conf", aa.StanzaName, name, title, description, defaultValue, required, false)
	if err != nil {
		return nil, fmt.Errorf("registerNewParam: %w", err)
	}
	p.SetCustomProperty("uiType", uiType)
	p.SetCustomProperty("placeholder", placeholder)
	if aa.params == nil {
		aa.params = make([]*params.Param, 0, 1)
	}
	aa.params = append(aa.params, p)
	return p, nil
}

// GetParam searches for the param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *AlertAction) GetParam(name string) (*params.Param, error) {
	for _, p := range aa.params {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("getParam[%s]: not found", name)
}

// GetParamNames returns a list of all the parameters defined for the alert action so far
func (aa *AlertAction) GetParamNames() []string {
	var paramsList = make([]string, len(aa.params))
	for i, p := range aa.params {
		paramsList[i] = p.GetName()
	}
	return paramsList
}

// RegisterGlobalParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (aa *AlertAction) RegisterGlobalParam(p *params.Param) error {
	// check if the parameter is already present
	// return error in case it is already there
	if _, err := aa.GetGlobalParam(p.GetName()); err == nil {
		return errors.NewErrInvalidParam("registerGlobalParam["+p.GetName()+"]", nil, "'%s' already exists", p.GetName())
	}

	if aa.globalParams == nil {
		aa.globalParams = make([]*params.Param, 0, 1)
	}

	aa.globalParams = append(aa.globalParams, p)
	return nil
}

// RegisterNewGlobalParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (aa *AlertAction) RegisterNewGlobalParam(configFile, stanza, name, title, description, defaultValue string, required bool) (*params.Param, error) {
	var p *params.Param
	var err error
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = aa.GetGlobalParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewGlobalParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam(configFile, stanza, name, title, description, defaultValue, required, false)
	if err != nil {
		return nil, fmt.Errorf("registerNewGlobalParam[%s]: %w", name, err)
	}
	if aa.globalParams == nil {
		aa.globalParams = make([]*params.Param, 0, 1)
	}
	aa.globalParams = append(aa.globalParams, p)
	return p, nil
}

// GetGlobalParam searches for the global param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *AlertAction) GetGlobalParam(name string) (*params.Param, error) {
	for _, p := range aa.globalParams {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("getGlobalParam[%s]: not found", name)
}

// GetFirstResults returns the first of the search results which the alert has been invoked on.
func (aa *AlertAction) GetFirstResult() map[string]interface{} {
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

// GetApp returns the name of the app containing the search which triggered the alert action
func (aa *AlertAction) GetApp() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetApp invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.App
}

// GetOwner returns the name of the app containing the search which triggered the alert action
func (aa *AlertAction) GetOwner() string {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetOwner invoked without a runtime-configuration having being loaded.")
		return ""
	}
	return aa.runtimeConfig.Owner
}

// GetResultsFile TBD: closing the file must be done by the user.
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

func (aa *AlertAction) GetNamespace() (*splunkd.Namespace, error) {
	if aa.runtimeConfig == nil {
		return nil, fmt.Errorf("getNamespace invoked without a runtime-configuration having being loaded")
	}
	return splunkd.NewNamespace(aa.runtimeConfig.Owner, aa.runtimeConfig.App, "")
}

// setSplunkService configures the splunkd client
// Prerequisites to execution: a runtime configuration must be already available (aa.initRuntime()) when performing this method.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the alert.
func (aa *AlertAction) setSplunkService() error {
	var (
		ss  *splunkd.Client
		err error
	)
	if aa.splunkd != nil {
		// already available
		return nil
	}
	if aa.runtimeConfig == nil {
		panic("setSplunkService: no runtime config available. Execute this method after initializing the internal data structures")
	}
	// alert actions run locally on splunk servers. It might well be that certificates are self-generated there.
	ss, err = splunkd.New(aa.runtimeConfig.ServerUri, true, "")
	if err != nil {
		return fmt.Errorf("setSplunkService: cannot create splunkd service. %w", err)
	}
	ss.SetNamespace(aa.GetOwner(), aa.GetApp(), splunkd.SplunkSharingGlobal)
	if err != nil {
		return fmt.Errorf("setSplunkService: cannot set namespace. %w", err)
	}

	if err = ss.LoginWithSessionKey(aa.runtimeConfig.SessionKey); err != nil {
		return fmt.Errorf("setSplunkService: %w", err)
	}

	aa.splunkd = ss
	return nil
}

// GetRunId return a unique string identifying the execution of the alert. This can be used to refer to internal logs, troubleshooting etc.
func (aa *AlertAction) GetRunId() string {
	return aa.runID
}

// GetSplunkService returns a client which can be used to communicate with splunkd.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the alert.
func (aa *AlertAction) GetSplunkService() (*splunkd.Client, error) {
	if aa.splunkd != nil {
		return aa.splunkd, nil
	}
	if err := aa.setSplunkService(); err != nil {
		return nil, err
	}
	return aa.splunkd, nil
}

// initRuntime is responsible to load the runtime-configuration and initialize all necessary internal data structures
// This function must be executed before the actual execution of the alerting function.
func (aa *AlertAction) initRuntime(c *alertConfig) error {
	aa.runtimeConfig = c
	// order of the following calls is important, as they are depending on runtimeConfig and splunkService
	if err := aa.setSplunkService(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	if err := aa.registerLogger(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}

	// it is important to log this after the setting of the logger, but before the configuration of the parameters.
	aa.Log("INFO", `Execution started. app="%s" owner="%s", search_name="%s", sid="%s"`, aa.GetApp(), aa.GetOwner(), aa.GetSearchName(), aa.GetSid())

	if err := aa.setGlobalParams(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	if err := aa.setParams(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	return nil
}

func (aa *AlertAction) setGlobalParams() error {
	if aa.splunkd == nil {
		panic("setGlobalParams: no splunk service available. Execute this method after setSplunkService()")
	}
	var val, loggedVal string
	var err error
	for _, param := range aa.globalParams {
		// in case the value of the parameter has been set interactively, skip looking for it within splunk
		if !param.HasForcedValue() {
			val, err = param.GetValueNS(aa.splunkd, aa.GetOwner(), aa.GetApp())
			if err != nil {
				return fmt.Errorf("setGlobalParams: cannot retrieve value of global parameter '%s' within scope user='%s' app='%s'. %w", param.String(), aa.GetOwner(), aa.GetApp(), err)
			}

			if val == "" && param.GetDefaultValue() == "" && param.IsRequired() {
				return fmt.Errorf("setGlobalParams: required parameter cannot have emtpy value '%s'", param.String())
			} else if val != "" {
				loggedVal = val
				if param.IsSensitive() {
					loggedVal = "***masked***"
				}
				aa.Log("INFO", "Setting global parameter %s=\"%s\"", param.String(), loggedVal)
				param.ForceValue(val)
			}
		}
	}
	return nil
}

// setParams stores run-time parameter within the alert action and its parameters.
// Returns an error if any of these actions failed
func (aa *AlertAction) setParams() error {
	if aa.runtimeConfig == nil {
		panic("setParams: no runtime config available. Execute this method after initializing the internal data structures")
	}

	var loggedVal string
	// assign the actual value to the parameters
	for _, param := range aa.params {
		if v, found := aa.runtimeConfig.Configuration[param.GetName()]; found {
			loggedVal = v
			if param.IsSensitive() {
				loggedVal = "***masked***"
			}
			aa.Log("INFO", "Setting parameter %s=\"%s\"", param.GetName(), loggedVal)
			if err := param.ForceValue(v); err != nil {
				return fmt.Errorf("setParams: error while applying run-time configuration: %s", err.Error())
			}
		} else {
			aa.Log("DEBUG", "Parameter '%s' uses default value \"%s\"", param.GetName(), param.GetDefaultValue())
		}
	}
	return nil
}

// RegisterValidationFunc configures a function used to validate parameters.
// Basic parameter validation is done automatically. This is needed to check for dependencies across multiple parameters.
// Providing a validation function is optional.
func (aa *AlertAction) RegisterValidationFunc(f AlertingFunc) {
	aa.Log("DEBUG", "Custom parameter validation function registered")
	aa.validateParams = f
}

// RegisterAlertFunc configures the actual alerting function to be executed by the alert action.
// Not providing a function results in a run-time error, as the alert action would not know what to actually do.
func (aa *AlertAction) RegisterAlertFunc(f AlertingFunc) {
	aa.Log("DEBUG", "Alerting function registered")
	aa.execute = f
}

// Run is the function responsible for actual execution of the alert action.
// Under normal execution (invokation by splunk), this is responsible to:
//
// - parse the configurations on STDIN,
// - assign actual values to the parameters,
// - initialize the runtime
// - execute the defined validation function, if any
// - execute the alerting function
//
// Run is additionally responsible to parse command-line arguments and provide the utilities to generate splunk configuration files, between others.
func (aa *AlertAction) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var err error
	var runTimeConfig *alertConfig
	// set interfaces to outside world
	aa.stdin = stdin
	aa.stdout = stdout
	aa.stderr = stderr

	// configure standard command line parameters
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	// defines the command-line parameters using the 'flag' module
	executePtr := flags.Bool("execute", false, "Starts execution of the alert action. A JSON-based configuration must be provided via STDIN. This is what Splunk does.")
	debugPtr := flags.Bool("debug", false, "Activates debug mode, useful only during development")
	interactivePtr := flags.Bool("interactive", false, "Interactively ask for parameter values and start a local execution. Useful for development and debugging only.")
	getRunTimeConfPtr := flags.Bool("get-runtime-conf-example", false, fmt.Sprintf("Interactively ask for parameter values and generates a JSON-based configuration, as Splunk would send to your alert. You can use this as 'cat conf.json > %s -execute'.", args[0]))
	getConfPtr := flags.Bool("get-alert-actions-conf", false, "Print out a template for default/alert_actions.conf")
	getSpecPtr := flags.Bool("get-alert-actions-spec", false, "Print out a template for README/alert_actions.conf.spec")
	getCustConfPtr := flags.Bool("get-custom-config-conf", false, "Print out a template for default/<custom-config>.conf")
	getCustSpecPtr := flags.Bool("get-custom-config-spec", false, "Print out a template for README/<custom-config>.conf.spec")
	getRestMapConfPtr := flags.Bool("get-rest-map-conf", false, "Print out a template for default/restmap.conf")
	getSSSpecPtr := flags.Bool("get-saved-searches-spec", false, "Print out a template for README/savedsearches.conf.spec")
	getDocuPtr := flags.Bool("get-documentation", false, "Print out markdown-formatted documentation for the alert")
	getUIHTML := flags.Bool("get-ui-html", false, fmt.Sprintf("Print out a template for the UI configuration to be stored at default/data/ui/alerts/%s.html", aa.StanzaName))

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *debugPtr {
		aa.EnableDebug()
	}

	if *executePtr || *interactivePtr {
		start := time.Now()

		if aa.execute == nil {
			aa.Log("FATAL", "No actual alerting function has been defined")
			return fmt.Errorf("no actual alerting function has been defined")
		}

		if *executePtr {
			aa.Log("INFO", "Parsing run-time JSON configurations from STDIN")
			runTimeConfig, err = getAlertConfigFromJSON(stdin)
			if err != nil {
				aa.Log("FATAL", "Parsing of run-time JSON configurations from STDIN failed. %s", err.Error())
				return err
			}
		} else if *interactivePtr {
			if runTimeConfig, err = aa.getAlertConfigInteractive(); err != nil {
				aa.Log("FATAL", "Error when preparing execution configuration: %s", err.Error())
				return err
			}
		}

		aa.Log("DEBUG", "Setting run-time configuration: %+v", runTimeConfig)
		// initRuntime is in charge of logging the "Execution started" message
		if err = aa.initRuntime(runTimeConfig); err != nil {
			aa.Log("FATAL", "Setting of run-time configurations failed. %s", err.Error())
			return err
		}

		// Note: initRuntime() already performs validation of individual parameters.
		// However, sometimes multiple parameters should be analyzed as a group for dependencies between them.
		// The function registered at "validateParams" is supposed to take care of that
		if aa.validateParams != nil {
			aa.Log("INFO", "Validating run-time parameters with registered function")
			if err = aa.validateParams(aa); err != nil {
				aa.Log("FATAL", "Validation of run-time parameters failed. %s", err.Error())
				return err
			}
		}
		// At last, perform actual execution of the alerting function
		aa.Log("INFO", "Executing alerting function")
		if err = aa.execute(aa); err != nil {
			aa.Log("FATAL", `Execution failed. sid="%s" duration_ms=%d. %s`, aa.GetSid(), time.Since(start).Milliseconds(), err.Error())
			return err
		}
		aa.Log("INFO", `Execution succeeded. sid="%s" duration_ms=%d`, aa.GetSid(), time.Since(start).Milliseconds())
		return nil
	}

	var actionSelected bool
	if *getConfPtr {
		fmt.Println(aa.generateAlertActionsConf())
		actionSelected = true
	}

	if *getSpecPtr {
		fmt.Println(aa.generateAlertActionsSpec())
		actionSelected = true
	}
	if *getSSSpecPtr {
		fmt.Println(aa.generateSavedSearchesSpec())
		actionSelected = true
	}

	if *getUIHTML {
		fmt.Println(aa.generateUIXML())
		actionSelected = true
	}

	if *getRunTimeConfPtr {
		fmt.Println(aa.generateRuntimeConfig(args[0]))
		actionSelected = true
	}
	if *getRestMapConfPtr {
		fmt.Println(aa.generateRestMapConf())
		actionSelected = true
	}
	if *getCustConfPtr {
		fmt.Println(aa.generateAdHocConfigConfs())
		actionSelected = true
	}
	if *getCustSpecPtr {
		fmt.Println(aa.generateAdHocConfigSpecs())
		actionSelected = true
	}
	if *getDocuPtr {
		fmt.Println(aa.generateDocumentation())
		actionSelected = true
	}
	// if no valid command-line parameters were provided
	if !actionSelected {
		printHelp(aa, flags, stderr)
	}
	return nil
}

// printHelp prints command-line usage instructions to STDOUT
func printHelp(aa *AlertAction, f *flag.FlagSet, out io.Writer) {
	if aa == nil {
		return
	}
	fmt.Fprintf(out, "Usage for custom alert action '%s'\n", aa.StanzaName)
	fmt.Fprintf(out, "Label: %s\n", aa.Label)
	fmt.Fprintf(out, "Description: %s\n", aa.Description)
	fmt.Fprintln(out, "NOTE: Splunk invokes the alert action using the '--execute' flag, unless differently specified within alert_actions.conf")
	fmt.Fprintln(out, "")
	if f != nil {
		f.SetOutput(out)
		f.PrintDefaults()
	}
}
