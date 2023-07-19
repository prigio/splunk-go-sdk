package alertactions

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
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
	// globalParams is used to track the global parameters necessary for the alert.
	// "global", in that they are tracked in a dedicate configuration file and are not configured within the alert UI
	globalParams map[string]*GlobalParam

	// validateParams is an optional function which can be used to validate the run-time parameters
	validateParams ValidateFunc

	// Execute is a mandatory function used to perform actual alert tasks. This is called by the alert's "Run" method.
	execute AlertFunc

	// This debug setting is meant for facilitating development and is not configurable by a user through splunk's inputs.conf
	debug bool

	// Unique id of this run, generated when starting the "Run" function
	runID string

	// actual run-time configurations provided by Splunk
	runtimeConfig *alertConfig

	splunkservice *client.SplunkService
	splunkdlogger *log.Logger
	// endUserLogger is used to log messages for the end user in an index preconfigured by them
	endUserLogger *log.Logger
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
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
	return nil, fmt.Errorf("alert action parameter not found. name=\"%s\"", name)
}

// GetParamNames returns a list of all the parameters defined for the alert action so far
func (aa *AlertAction) GetParamNames() []string {
	var paramsList = make([]string, len(aa.params))
	for i, p := range aa.params {
		paramsList[i] = p.Name
	}
	return paramsList
}

// RegisterGlobalParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (aa *AlertAction) RegisterGlobalParam(configFile, stanza, name, title, description, defaultValue string, required bool) (*GlobalParam, error) {
	configFile = strings.TrimSuffix(configFile, ".conf")
	if configFile == "" {
		return nil, fmt.Errorf("invalid alert action global parameter defined: 'configFile' cannot be empty")
	}
	if stanza == "" {
		return nil, fmt.Errorf("invalid alert action global parameter defined: 'stanza' cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("invalid alert action global parameter defined: 'name' cannot be empty")
	}
	if title == "" {
		return nil, fmt.Errorf("invalid alert action global parameter defined: 'title' cannot be empty")
	}

	// check if the parameter is already present
	// return error in case it is already there
	if _, err := aa.GetGlobalParam(name); err == nil {
		return nil, fmt.Errorf("global parameter already existing. It is not possible to add multiple parameters having the same name. name=\"%s\"", name)
	}

	if aa.globalParams == nil {
		aa.globalParams = make(map[string]*GlobalParam)
	}

	param := &GlobalParam{
		ConfigFile: configFile,
		Stanza:     stanza,
	}
	param.Title = title
	param.Name = name
	param.Description = description
	param.DefaultValue = defaultValue
	param.Required = required

	aa.globalParams[name] = param
	return param, nil
}

// GetGlobalParam searches for the global param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *AlertAction) GetGlobalParam(name string) (*GlobalParam, error) {
	var p *GlobalParam
	var found bool
	if aa.globalParams == nil {
		found = false
	} else {
		p, found = aa.globalParams[name]
	}

	if found {
		return p, nil
	}
	return nil, fmt.Errorf("getGlobalParam: not found. name=\"%s\"", name)
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

func (aa *AlertAction) GetNamespace() (owner, app string) {
	if aa.runtimeConfig == nil {
		aa.Log("ERROR", "GetNamespace invoked without a runtime-configuration having being loaded.")
		return "", ""
	}
	return aa.runtimeConfig.Owner, aa.runtimeConfig.App
}

// setSplunkService configures the splunkd client
// Prerequisites to execution: a runtime configuration must be already available (aa.setConfig()) when performing this method.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the alert.
func (aa *AlertAction) setSplunkService() error {
	if aa.splunkservice != nil {
		// already available
		return nil
	}
	if aa.runtimeConfig == nil {
		return fmt.Errorf("alert action setSplunkService: no runtime config available. impossible to use it to initialize splunkd client")
	}
	// alert actions run locally on splunk servers. It might well be that certificates are self-generated there.
	ss, err := client.New(aa.runtimeConfig.ServerUri, true, "")
	if err != nil {
		return fmt.Errorf("alert action setSplunkService: %w", err)
	}

	if err := ss.LoginWithSessionKey(aa.runtimeConfig.SessionKey); err != nil {
		return fmt.Errorf("alert action setSplunkService: %w", err)
	}

	aa.splunkservice = ss
	return nil
}

// GetRunId return a unique string identifying the execution of the alert. This can be used to refer to internal logs, troubleshooting etc.
func (aa *AlertAction) GetRunId() string {
	return aa.runID
}

// GetSplunkService returns a client which can be used to communicate with splunkd.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the alert.
func (aa *AlertAction) GetSplunkService() (*client.SplunkService, error) {
	if aa.splunkservice != nil {
		return aa.splunkservice, nil
	}
	if err := aa.setSplunkService(); err != nil {
		return nil, err
	}
	return aa.splunkservice, nil
}

// initRuntime is responsible to load the runtime-configuration and initialize all necessary internal data structures
// This function must be executed before the actual execution of the alerting function.
func (aa *AlertAction) initRuntime(c *alertConfig) error {
	aa.runtimeConfig = c
	// order of the following calls is important, as they are depending on runtimeConfig and splunkService
	if err := aa.setSplunkService(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	if err := aa.setLogger(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}

	// it is important to log this after the setting of the logger, but before the configuration of the parameters.
	aa.Log("INFO", "Execution started. sid=\"%s\"", aa.GetSid())

	if err := aa.setGlobalParams(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	if err := aa.setParams(); err != nil {
		return fmt.Errorf("initRuntime: %w", err)
	}
	return nil
}

func (aa *AlertAction) setGlobalParams() error {
	if aa.splunkservice == nil {
		return fmt.Errorf("setGlobalParams: no splunk service available. Execute this method after setConfigAndParams() and setSplunkService()")
	}
	var configsCollection *client.ConfigsCollection
	var stanza *client.ConfigResource
	var loggedVal string
	var err error
	for _, param := range aa.globalParams {
		configsCollection = aa.splunkservice.GetConfigs(param.ConfigFile)
		stanza, err = configsCollection.GetStanza(param.Stanza)
		if err != nil {
			return fmt.Errorf("setGlobalParams: stanza '%s' not found in config '%s'. %w", param.Stanza, param.ConfigFile, err)
		}
		if val, err := stanza.GetString(param.Name); err != nil {
			if param.Required {
				return fmt.Errorf("setGlobalParams: required parameter not found '%s:[%s]/%s'", param.ConfigFile, param.Stanza, param.Name)
			}
			aa.Log("WARN", "Global parameter %s:[%s]/%s not found. Using default value", param.ConfigFile, param.Stanza, param.Name)
		} else if val == "" && param.DefaultValue == "" && param.Required {
			return fmt.Errorf("setGlobalParams: required parameter cannot have emtpy value '%s:[%s]/%s'", param.ConfigFile, param.Stanza, param.Name)
		} else if val != "" {
			loggedVal = val
			if param.Sensitive {
				loggedVal = "***masked***"
			}
			aa.Log("INFO", "Setting global parameter %s:[%s]/%s=\"%s\"", param.ConfigFile, param.Stanza, param.Name, loggedVal)
			param.SetValue(val)
		}
	}
	return nil
}

// setParams stores run-time parameter within the alert action and its parameters.
// Returns an error if any of these actions failed
func (aa *AlertAction) setParams() error {
	if aa.runtimeConfig == nil {
		return fmt.Errorf("alert action setParams: no runtime config available")
	}
	var loggedVal string
	// assign the actual value to the parameters
	for _, param := range aa.params {
		if v, found := aa.runtimeConfig.Configuration[param.Name]; found {
			loggedVal = v
			if param.Sensitive {
				loggedVal = "***masked***"
			}
			aa.Log("INFO", "Setting parameter %s=\"%s\"", param.Name, loggedVal)
			if err := param.SetValue(v); err != nil {
				return fmt.Errorf("error while applying run-time configuration: %s", err.Error())
			}
		} else {
			aa.Log("DEBUG", "Parameter '%s' uses default value \"%s\"", param.Name, param.GetValue())
		}
	}
	return nil
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
func (aa *AlertAction) printHelp(f *flag.FlagSet) {
	fmt.Printf("Usage for custom alert action '%s'\n", aa.StanzaName)
	fmt.Printf("Label: %s\n", aa.Label)
	fmt.Printf("Description: %s\n", aa.Description)
	fmt.Println("NOTE: Splunk invokes the alert action using the '--execute' flag, unless differently specified within alert_actions.conf")
	fmt.Println("")
	f.PrintDefaults()
}

// Run is the function responsible for actual execution of the alert action.
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
	getRestMapConfPtr := flags.Bool("get-rest-map-conf", false, "Print out a template for default/restmap.conf")
	getSSSpecPtr := flags.Bool("get-saved-searches-spec", false, "Print out a template for README/savedsearches.conf.spec")
	getUIHTML := flags.Bool("get-ui-html", false, fmt.Sprintf("Print out a template for the UI configuration to be stored at default/data/ui/alerts/%s.html", aa.StanzaName))
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *debugPtr {
		aa.EnableDebug()
	}

	if *executePtr {
		start := time.Now()
		if aa.execute == nil {
			aa.Log("FATAL", "No actual alerting function has been defined")
			return fmt.Errorf("no actual alerting function has been defined")
		}

		aa.Log("DEBUG", "Parsing run-time JSON configurations from STDIN")
		runTimeConfig, err = getAlertConfigFromJSON(stdin)
		if err != nil {
			aa.Log("FATAL", "Parsing of run-time JSON configurations from STDIN failed. %s", err.Error())
			return err
		}

		aa.Log("DEBUG", "Setting run-time configuration: %+v", runTimeConfig)
		// initRuntime is in charge of logging the "Execution started" message
		err = aa.initRuntime(runTimeConfig)
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
		duration := time.Since(start)
		if err != nil {
			aa.Log("FATAL", "Execution failed. sid=\"%s\" duration_ms=%d. %s", aa.GetSid(), duration.Milliseconds(), err.Error())
			return err
		}
		aa.Log("INFO", "Execution succeeded. sid=\"%s\" duration_ms=%d", aa.GetSid(), duration.Milliseconds())
		return nil
	}

	if *interactivePtr {
		if runTimeConfig, err = aa.getAlertConfigInteractive(); err != nil {
			aa.Log("FATAL", "Errow when preparing execution configuration: %s", err.Error())
			return err
		} else {
			aa.Log("DEBUG", "Setting run-time configuration: %+v", runTimeConfig)
			err = aa.initRuntime(runTimeConfig)
			if err != nil {
				aa.Log("FATAL", "Setting of run-time configurations failed. %s", err.Error())
				return err
			}
		}
		return aa.execute(aa)
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

	if *getUIHTML {
		fmt.Println(aa.generateUIHTML())
		return nil
	}

	if *getRunTimeConfPtr {
		fmt.Println(aa.generateRuntimeConfig(args[0]))
		return nil
	}
	if *getRestMapConfPtr {
		fmt.Println(aa.generateRestMapConf())
		return nil
	}
	// if no valid command-line parameters were provided
	aa.printHelp(flags)
	return nil
}
