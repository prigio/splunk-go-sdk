package modinputs

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-isatty"
	"github.com/prigio/splunk-go-sdk/v2/errors"
	"github.com/prigio/splunk-go-sdk/v2/params"
	"github.com/prigio/splunk-go-sdk/v2/splunkd"
)

// isAtTerminal is a boolean which is true if the alert action is being executed on a command-line or not.
// this is used to modify the logging format
var isAtTerminal = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

// StreamingFunc is the signature of the function used to generate the data for the modular input
type StreamingFunc func(*ModularInput, Stanza) error

// StreamingFuncSingleInstance is the signature of the function used to generate the data for the modular input when running in single instance mode
type StreamingFuncSingleInstance func(*ModularInput, []Stanza) error

// ValidationFun is the signature of the function used to validate the parameters received from Splunk
// (only used if the mod input is configured to use external validation)
type ValidationFunc func(*ModularInput, Stanza) error

// ModularInput is the main structure defining how a modular input looks like.
// It provides a way for the user to define a Splunk modular input and makes
// standardised methods available.
type ModularInput struct {
	// Name appearing within inputs.conf stanza as [stanzaname://...].
	// Must be lowercase
	StanzaName string
	// Title displayed on the UI
	Title string
	// Description within the UI
	Description string
	// Markdown-formatted documentation content about the logic behind the modular input execution.
	// This, along with other pre-defined contents can be printed out by starting the modular input from the commandline with the proper parameter. See './<modular-input-name> -h'
	Documentation string

	useExternalValidation bool
	useSingleInstance     bool
	params                []*params.Param
	// globalParams is used to track the global parameters necessary for the alert.
	// "global", in that they are tracked in a dedicate configuration file and are not configured within the alert UI
	globalParams []*params.Param

	// (optional) function used to validate data. Expected only if the modular input is configured to use "external validation"
	validate ValidationFunc
	// function used to stream generated data when the modular input is executed once per each configuration stanza
	stream StreamingFunc

	// function used to stream generated data when the modular input is executed in single-instance mode: once for all configuration stanzas
	streamSingleInstance StreamingFuncSingleInstance

	// This debug setting is meant for facilitating development and is not configurable by a user through splunk's inputs.conf
	debug bool

	// This is used in case no sourcetype has been set within local/inputs.conf
	defaultSourcetype string
	// This is used in case no index has been configured within local/inputs.conf
	defaultIndex string

	/*
		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer
	*/
	// These parameters are read-in from the XML-based configurations provided on stdin by splunk upon execution
	splunkd       *splunkd.Client
	hostname      string
	uri           string
	sessionKey    string
	checkpointDir string
	stanzas       []Stanza
	// Unique id of this run, generated when starting the "Run" function
	runID string
	// private variables
	internalLogEvent               *SplunkEvent //this is used to setup a standardized event using for logging to index=_internal. If this is not nil, internal loggin is performed through SplunkEvents written on Stdout instead of plain output on Stderr
	cntDataEventsGeneratedbyStanza int64        // counter of data events emitted by the stanza being currently processed (internal loggin is excluded)
	cntDataEventsGeneratedTotal    int64        // counter of data events emitted in total (internal loggin is excluded)
}

func New(stanzaName, label, description string) (*ModularInput, error) {
	if stanzaName == "" {
		return nil, errors.NewErrInvalidParam("modularInput.New", nil, "'stanzaName' cannot be empty")
	}
	if label == "" {
		return nil, errors.NewErrInvalidParam("modularInput.New", nil, "'label' cannot be empty")
	}

	var mi = &ModularInput{
		StanzaName:        stanzaName,
		Title:             label,
		Description:       description,
		defaultSourcetype: stanzaName,
		runID:             uuid.New().String()[0:8],
	}
	return mi, nil
}

// SetSingleInstanceExecution configures the Modular Input to execute on all configuration stanzas being provided by splunk at once.
func (mi *ModularInput) SetSingleInstanceExecution() error {
	if mi.stream != nil {
		return fmt.Errorf("SetSingleInstanceExecution: cannot activate single instance execution mode when a streaming function for single stanzas has been configured")
	}
	mi.useSingleInstance = true
	return nil
}

// RegisterStreamingFunc registera a streaming function to be executed on one configuration stanza which is provided by Splunk at run time
func (mi *ModularInput) RegisterStreamingFunc(f StreamingFunc) error {
	if mi.useSingleInstance {
		return fmt.Errorf(("registerStreamingFunc: cannot register a streaming function when SingleInstanceMode is activated"))
	}
	mi.stream = f
	return nil
}

// RegisterStreamingFunc registera a streaming function to be executed on one configuration stanza which is provided by Splunk at run time
func (mi *ModularInput) RegisterStreamingFuncSingleInstance(f StreamingFuncSingleInstance) error {
	if !mi.useSingleInstance {
		return fmt.Errorf("registerStreamingFuncSingleInstance: cannot register a single-instance streaming function when SingleInstanceMode is not activated")
	}
	mi.streamSingleInstance = f
	return nil
}

// EnableDebug sets debug mode for the modular input
func (mi *ModularInput) EnableDebug() {
	mi.debug = true
}

// IsDebug returns true if debug mode has been activated for the modular input
func (mi *ModularInput) IsDebug() bool {
	return mi.debug
}

func (mi *ModularInput) GetRunId() string {
	if mi.runID == "" {
		mi.runID = uuid.New().String()[0:8]
	}
	return mi.runID
}

// GetSplunkService returns a client which can be used to communicate with splunkd.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the modular input.
func (mi *ModularInput) GetSplunkService() (*splunkd.Client, error) {
	if mi.splunkd != nil {
		return mi.splunkd, nil
	}
	if err := mi.setSplunkService(); err != nil {
		return nil, err
	}
	return mi.splunkd, nil
}

// setSplunkService configures the splunkd client
// Prerequisites to execution: a runtime configuration (sessionkey + splunkd URI) must be already available when performing this method.
// The client has already been authenticated using the sessionKey which Splunk provides when starting the modular input.
func (mi *ModularInput) setSplunkService() error {
	var (
		ss  *splunkd.Client
		err error
	)
	if mi.splunkd != nil {
		// already available
		return nil
	}
	if mi.sessionKey == "" || mi.uri == "" {
		return fmt.Errorf("setSplunkService: cannot instantiate a splunkd client as the necessary sessionKey and Uri have not been initialized")
	}
	// alert actions run locally on splunk servers. It might well be that certificates are self-generated there.
	ss, err = splunkd.New(mi.uri, true, "")
	if err != nil {
		return fmt.Errorf("setSplunkService: %w", err)
	}
	//ss.SetNamespace()
	if err = ss.LoginWithSessionKey(mi.sessionKey); err != nil {
		return fmt.Errorf("setSplunkService: %w", err)
	}
	mi.splunkd = ss
	return nil
}

// RegisterNewParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
//
// The following are the only adminissible values for the dataType. Anything else will generate an error.
// - "string"
// - "boolean"
// - "number"
func (mi *ModularInput) RegisterNewParam(name, title, description, defaultValue, dataType, validation string, required, sensitive bool) (*params.Param, error) {
	var (
		p   *params.Param
		err error
	)
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = mi.GetParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam("inputs.conf", mi.StanzaName, name, title, description, defaultValue, required, sensitive)
	if err != nil {
		return nil, fmt.Errorf("registerNewParam: %w", err)
	}
	err = p.SetDataType(dataType)
	if err != nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", err, `'dataType' provided="%s"`, dataType)
	}
	if validation != "" {
		p.SetCustomProperty("validation", validation)
	}
	if mi.params == nil {
		mi.params = make([]*params.Param, 0, 1)
	}
	mi.params = append(mi.params, p)
	return p, nil
}

// GetParam searches for the param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (mi *ModularInput) GetParam(name string) (*params.Param, error) {
	for _, p := range mi.params {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("getParam[%s]: not found", name)
}

// RegisterNewGlobalParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (mi *ModularInput) RegisterNewGlobalParam(configFile, stanza, name, title, description, defaultValue string, required, sensitive bool) (*params.Param, error) {
	var p *params.Param
	var err error
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = mi.GetGlobalParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewGlobalParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam(configFile, stanza, name, title, description, defaultValue, required, sensitive)
	if err != nil {
		return nil, fmt.Errorf("registerNewGlobalParam[%s]: %w", name, err)
	}
	if mi.globalParams == nil {
		mi.globalParams = make([]*params.Param, 0, 1)
	}
	mi.globalParams = append(mi.globalParams, p)
	return p, nil
}

// GetGlobalParam searches for the global param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *ModularInput) GetGlobalParam(name string) (*params.Param, error) {
	for _, p := range aa.globalParams {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf(`getGlobalParam[%s]: parameter not found`, name)
}

func (mi *ModularInput) RegisterValidationFunc(f ValidationFunc) {
	mi.useExternalValidation = true
	mi.validate = f
}

// Log writes a log so that it can be read by Splunk without being interpreted as an actual event generated by the script
// Argument 'message' can use formatting markers as fmt.Sprintf. Aditional arguments 'a' will be provided to fmt.Sprintf
func (mi *ModularInput) Log(level string, message string, a ...interface{}) (err error) {
	level = strings.ToUpper(level)
	if level == "WARNING" {
		// Typical error, just manage it...
		level = "WARN"
	}
	if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" && level != "FATAL" {
		fmt.Fprintf(os.Stderr, "ERROR - ModularInput.Log invoked with invalid level parameter. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'\n", level)
		return fmt.Errorf("ModularInput.Log: invalid value of 'level' provided. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'", level)
	}
	if level != "DEBUG" || (level == "DEBUG" && mi.debug) {
		// do not do anything if debug is not enabled
		t := time.Now().Round(time.Millisecond)
		if mi.internalLogEvent != nil {
			mi.internalLogEvent.Time = t
			// prefix the message with timestamp and log_level
			message = "[" + t.Format("2006-01-02 15:04:05.000 -0700") + "] " + level + " run_id=" + mi.runID + " - " + message
			//time.Format uses a string with such parameters to define the output format: Mon Jan 2 15:04:05 -0700 MST 2006
			mi.internalLogEvent.Data = fmt.Sprintf(message, a...)
			// using writeOut() to skip counting the events, as we do not want to count the internal logs...
			mi.internalLogEvent.writeOut()
		} else {
			// XML-based logging has not yet been activated: using STDERR instead
			message = "ModularInput " + mi.StanzaName + ": " + level + " run_id=" + mi.runID + " - " + message + "\n"
			_, err = fmt.Fprintf(os.Stderr, message, a...)
		}
	}
	return err
}

// logPlain forces a plain-text write to STDERR. This is useful to force the log to appear within splunk's splunkd.log,
// same as the ones indicating the start of the run.
// For info related to the arguments, see Log()
func (mi *ModularInput) logPlain(level string, message string, a ...interface{}) (err error) {
	level = strings.ToUpper(level)
	if level == "WARNING" {
		// Typical error, just manage it...
		level = "WARN"
	}
	if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" && level != "FATAL" {
		fmt.Fprintf(os.Stderr, "ERROR - ModularInput.Log invoked with invalid level parameter. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'\n", level)
		return fmt.Errorf("ModularInput.Log: invalid value of 'level' provided. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'", level)
	}
	// do not do anything if debug is not enabled
	if level != "DEBUG" || (level == "DEBUG" && mi.debug) {
		message = "ModularInput " + mi.StanzaName + ": " + level + " run_id=" + mi.runID + " - " + message + "\n"
		_, err = fmt.Fprintf(os.Stderr, message, a...)
	}
	return err
}

// WriteToSplunk outputs a generated event in the format accepted by Splunk
// Returns the number of bytes written, an error if anything went wrong
// This function IS NOT concurrency safe!
func (mi *ModularInput) WriteToSplunk(se *SplunkEvent) error {
	if xmlStr, err := se.xml(); err != nil {
		return err
	} else {
		// increase the counter of the generated events
		mi.cntDataEventsGeneratedbyStanza++
		mi.cntDataEventsGeneratedTotal++
		_, err = os.Stdout.WriteString(xmlStr)
		return err
	}
}

// SetDefaultSourcetype configures a sourcetype to be used if none has been received from the run-time configurations.
// Additionally, the default sourcetype is used when generating the template for default/inputs.conf
func (mi *ModularInput) SetDefaultSourcetype(st string) {
	mi.defaultSourcetype = st
}

// GetDefaultSourcetype returns the sourcetype used by the modular input to collect data
// in case there are no specific run-time configurations
func (mi *ModularInput) GetDefaultSourcetype() string {
	return mi.defaultSourcetype
}

// SetDefaultIndex configures an index to be used if none has been received from the run-time configurations.
// Additionally, the default index is used when generating the template for default/inputs.conf
func (mi *ModularInput) SetDefaultIndex(idx string) {
	mi.defaultIndex = idx
}

// GetDefaultIndex returns the index used by the modular input to collect data
// in case there are no specific run-time configurations
func (mi *ModularInput) GetDefaultIndex() string {
	return mi.defaultIndex
}

// NewDefaultEvent provides a template for the SplunkEvent to be used to log actual data to be imported to Splunk
func (mi *ModularInput) NewDefaultEvent(stanza *Stanza) (ev *SplunkEvent) {
	if stanza != nil {
		//mi.Log("DEBUG", fmt.Sprintf("NewDefaultEvent: stanza is NOT nil. %v", stanza.Params))
		// NOT specifying Data intentionally

		ev = &SplunkEvent{
			// Anything can be overridded by the actual script
			Time:       time.Now(),
			Stanza:     stanza.Name,
			SourceType: stanza.Sourcetype(),
			Index:      stanza.Index(),
			Host:       stanza.Host(),
			Source:     stanza.Source(),
			Unbroken:   false,
			Done:       false,
		}
		if stanza.Sourcetype() == "" {
			ev.SourceType = mi.defaultSourcetype
		}
	} else {
		// If no configurations are present, we basically just return a generic event
		ev = &SplunkEvent{
			Time:       time.Now(),
			SourceType: mi.defaultSourcetype,
			Unbroken:   false,
			Done:       false,
		}
	}
	return ev
}

// printHelp prints command-line usage instructions to STDOUT
func (mi *ModularInput) printHelp() {
	fmt.Printf("Usage for custom modular input '%s'\n", mi.StanzaName)
	fmt.Printf("Label: %s\n", mi.Title)
	fmt.Printf("Description: %s\n", mi.Description)
	fmt.Println("NOTE: Splunk invokes the modular-input three times with flags '--scheme', '--valildate-arguments' and with no flags at all to start actual execution")
	fmt.Println("")
	flag.PrintDefaults()
}

// Run is the main function that starts the actual processing.
// It reads the command-line parameters and performs the correct actions.
func (mi *ModularInput) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	mi.Log("DEBUG", "ModularInput.Run started. Cmd-line parameters: '%s'", strings.Join(args, " "))

	// configure standard command line parameters
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	schemePtr := flags.Bool("scheme", false, "Prints out the XML scheme definition. This is what Splunk does when starting up. See Splunk documentation.")
	validatePtr := flags.Bool("validate-arguments", false, "Validates the parameters provided on STDIN in XML format. This is what Splunk does when starting the modular input if 'external-validation' is set to true in 'inputs.conf'. See Splunk documentation")
	interactivePtr := flags.Bool("interactive", false, "Interactively ask for parameter values and start a local execution. Useful for development and debugging only.")
	getConfPtr := flags.Bool("get-inputs-conf", false, "Print out a template for default/inputs.conf")
	getSpecPtr := flags.Bool("get-inputs-spec", false, "Print out a template for README/inputs.conf.spec")
	getDocuPtr := flags.Bool("get-documentation", false, "Print out markdown-formatted documentation for the alert")
	getExamplePtr := flags.Bool("get-example", false, "Print an example of inputs.conf configuration for this modular input and exit")
	getRunTimeConfPtr := flags.Bool("get-runtime-conf-example", false, fmt.Sprintf("Interactively ask for parameter values and generates an XML-based configuration, as Splunk would send to your modular input. You can use this as 'cat conf.json > %s'.", args[0]))

	//debugPtr := flags.Bool("debug", false, "Activates debug mode, useful only during development")
	//getRunTimeConfPtr := flags.Bool("get-runtime-conf-example", false, fmt.Sprintf("Interactively ask for parameter values and generates a JSON-based configuration, as Splunk would send to your alert. You can use this as 'cat conf.json > %s -execute'.", args[0]))
	getCustConfPtr := flags.Bool("get-custom-config-conf", false, "Print out a template for default/<custom-config>.conf")
	getCustSpecPtr := flags.Bool("get-custom-config-spec", false, "Print out a template for README/<custom-config>.conf.spec")
	//getDocuPtr := flags.Bool("get-documentation", false, "Print out markdown-formatted documentation for the alert")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if len(args) == 1 {
		// no-command line flag. This signal actual execution of the modular input

		// Read XML configs from STDIN
		// Populates infos about the configuration Stanzas
		if ic, err := getInputConfigFromXML(stdin); err != nil {
			mi.Log("FATAL", "Errow when loading execution configuration XML from StdIn: %s", err.Error())
			return err
		} else {
			mi.Log("DEBUG", "Loaded input configurations: %+v", ic)
			mi.hostname = ic.Hostname
			mi.uri = ic.URI
			mi.sessionKey = ic.SessionKey
			mi.checkpointDir = ic.CheckpointDir
			mi.stanzas = ic.Stanzas
		}
		return mi.runStreaming()
	} else if *schemePtr {
		// print a XML definition of the parameters accepted by this modular input
		mi.Log("DEBUG", "starting --scheme action")
		if schemeXml, err := mi.generateXMLScheme(); err != nil {
			mi.Log("FATAL", "Error during scheme generation. %s", err.Error())
			return err
		} else {
			fmt.Println(schemeXml)
			return nil
		}
	} else if *validatePtr {
		// Read XML configs
		if vc, err := getValidationConfigFromXML(stdin); err != nil {
			mi.Log("FATAL", "Errow when loading parameters validation XML from StdIn: %s", err.Error())
			return err
		} else {
			// Assign the loaded configuration to the private vars
			// of the modularinput itself
			mi.Log("DEBUG", "Loaded validation configurations: %+v", vc)
			mi.hostname = vc.Hostname
			mi.uri = vc.URI
			mi.sessionKey = vc.SessionKey
			mi.checkpointDir = vc.CheckpointDir
			mi.stanzas = []Stanza{vc.Item}
		}
		return mi.runValidation()
	} else if *interactivePtr || *getRunTimeConfPtr {
		var ic *inputConfig
		var conf []byte
		var err error

		if ic, err = getInputConfigInteractive(mi); err != nil {
			mi.Log("FATAL", "Errow when preparing execution configuration interactively: %s", err.Error())
			return err
		} else {
			mi.Log("DEBUG", "Provided input configurations: %+v", ic)
			mi.hostname = ic.Hostname
			mi.uri = ic.URI
			mi.sessionKey = ic.SessionKey
			mi.checkpointDir = ic.CheckpointDir
			mi.stanzas = ic.Stanzas
		}
		if *interactivePtr {
			return mi.runStreaming()
		} else {
			if conf, err = xml.MarshalIndent(ic, "", "  "); err != nil {
				mi.Log("FATAL", "Error when marshaling configuration to XML: %s", err.Error())
				return err
			}
			fmt.Printf(`
<!-- 
You can use the following configuration by:
	1) saving this into a file, e.g. conf.json
	2) executing the modularinput as:
	
		cat conf.xml > %s
-->
%s
`, args[0], conf)
			return nil
		}
	} else if *getConfPtr {
		// print out the definition of the modular input for inputs.conf
		fmt.Fprintln(stdout, mi.generateInputsConf())
	} else if *getSpecPtr {
		// print out the definition of the modular input for inputs.conf.spec
		fmt.Fprintln(stdout, mi.generateInputsSpec())
	} else if *getExamplePtr {
		fmt.Fprintln(stdout, mi.generateExampleConf())
	} else if *getDocuPtr {
		fmt.Fprintln(stdout, mi.generateDocumentation())
	} else if *getCustConfPtr {
		fmt.Fprintln(stdout, mi.generateAdHocConfigConfs())
	} else if *getCustSpecPtr {
		fmt.Fprintln(stdout, mi.generateAdHocConfigSpecs())
	} else {
		mi.printHelp()
	}
	return nil
}

// runStreaming executes the data generation function configured within ModularInput mi
// on the input configurations provided as XML on stdin
func (mi *ModularInput) runStreaming() (err error) {
	// these two vars are used to track the duration of the overall streaming function
	var duration time.Duration
	mi.Log("DEBUG", "Starting 'runStreaming' function")
	if !mi.useSingleInstance && mi.stream == nil {
		mi.Log("FATAL", "No streaming function specified")
		return fmt.Errorf("FATAL: no streaming function specified")
	}
	if mi.useSingleInstance && mi.streamSingleInstance == nil {
		mi.Log("FATAL", "No streaming function specified for single-instance mode")
		return fmt.Errorf("FATAL: no streaming function specified for single-instance mode")
	}

	streamingStartTime := time.Now()

	fmt.Println("<stream>")        // Setup the XML streaming mode
	defer fmt.Println("</stream>") // close XML streaming mode when returning

	if mi.useSingleInstance {
		mi.setupEventBasedInternalLoggingSingleInstance()
		mi.Log("INFO", "Starting single-instance streaming for %d stanzas", len(mi.stanzas))
		startTime := time.Now()

		if len(mi.stanzas) > 0 {
			err = mi.streamSingleInstance(mi, mi.stanzas)
		}

		duration = time.Since(startTime)
		if err != nil {
			mi.Log("ERROR", `Execution status=failed. duration_s=%.03f cnt_events=%d error="%s"`, duration.Seconds(), mi.cntDataEventsGeneratedTotal, err.Error())
		} else {
			mi.Log("INFO", `Execution status=succeeded. duration_s=%.03f cnt_events=%d`, duration.Seconds(), mi.cntDataEventsGeneratedTotal)
		}

	} else {

		// when using single_instance_mode=false, only 1 stanza is provided as configuration.
		if len(mi.stanzas) == 0 {
			return fmt.Errorf("no configurazion stanzas are present within input configuration. Nothing to be done")
		}
		stanza := mi.stanzas[0]
		//Start logging internal messages as SplunkEvents instead of using plaintext on Stderror
		mi.setupEventBasedInternalLogging(&stanza)
		mi.Log("INFO", `Starting streaming for stanza="%s"`, stanza.Name)

		err = mi.stream(mi, stanza)

		duration = time.Since(streamingStartTime)
		if err != nil {
			mi.Log("ERROR", `Execution status=failed for stanza="%s" duration_s=%.03f cnt_events=%d error="%s"`, stanza.Name, duration.Seconds(), mi.cntDataEventsGeneratedbyStanza, err.Error())
		} else {
			mi.Log("INFO", `Execution status=succeeded for stanza="%s" duration_s=%.03f cnt_events=%d`, stanza.Name, duration.Seconds(), mi.cntDataEventsGeneratedbyStanza)
		}

	}

	return err
}

// runValidation executes the validation function configured within ModularInput mi
// on the validation configuration provided as XML on stdin
func (mi *ModularInput) runValidation() error {
	mi.Log("DEBUG", `Starting argument validation`)

	if !mi.useExternalValidation {
		mi.Log("WARN", "Invoked with --validate-arguments command-line arguments but configured to NOT use external validation. Skipping it.")
		return nil
	}
	if mi.useExternalValidation && mi.validate == nil {
		mi.Log("WARN", "Configured to use external validation, but no validation function was specified. Skipping it.")
		return nil
	}

	if err := mi.validate(mi, mi.stanzas[0]); err != nil {
		mi.Log("ERROR", `Validation of parameters for stanza="%s" status=failed error="%s"`, mi.stanzas[0].Name, err.Error())
		// Splunk specification requires to write the validation errors on STDOUT
		// See: https://docs.splunk.com/Documentation/SplunkCloud/8.1.2011/AdvancedDev/ModInputsScripts#Create_a_modular_input_script
		fmt.Printf("%s\n", err.Error())
		return err
	}

	mi.Log("INFO", `Validation of input parameters for stanza="%s" status=succeeded`, mi.stanzas[0].Name)
	return nil
}

// setupEventBasedInternalLogging configures logging to be performed through SplunkEvent events written to index=_internal instead of using plain text on standard-err.
// Before activating this, the user is informed with a INFO message on StdErr saying which source/sourcetype is being used for logging purposes from now on
// This function can only be invoked when an active configuration has been provided in input, so, when we start streaming events.
// If stanza==nil, this functions panics.
func (mi *ModularInput) setupEventBasedInternalLogging(stanza *Stanza) {
	if stanza != nil {
		inputSourcetype := "modinput:" + stanza.Scheme()
		mi.logPlain("INFO", `Starting execution of stanza="%s". Logging related internal data as 'index=_internal sourcetype="%s" source="%s"'`, stanza.Name, inputSourcetype, stanza.Name)
		mi.internalLogEvent = &SplunkEvent{
			// NOT specifying Data and Host intentionally
			Time:       time.Now(),
			Stanza:     stanza.Name,
			SourceType: inputSourcetype,
			Index:      "_internal",
			Source:     stanza.Name,
			Unbroken:   false,
			Done:       false,
		}
	} else {
		mi.logPlain("FATAL", "Function setupEventBasedInternalLogging() called without a stanza being specified, this is an error within the library. Interrupting execution.")
		panic("Library error: function setupEventBasedInternalLogging() called without a stanza being specified.")
	}
}

// setupEventBasedInternalLoggingSingleInstance configures logging to be performed through SplunkEvent events written to index=_internal instead of using plain text on standard-err.
// Before activating this, the user is informed with a INFO message on StdErr saying which source/sourcetype is being used for logging purposes from now on
func (mi *ModularInput) setupEventBasedInternalLoggingSingleInstance() {
	inputSourcetype := "modinput:" + mi.StanzaName
	mi.logPlain("INFO", `Starting single-instance execution. Logging internal data as 'index=_internal sourcetype="%s"'`, inputSourcetype)
	mi.internalLogEvent = &SplunkEvent{
		// NOT specifying Data and Host intentionally
		Time:       time.Now(),
		Stanza:     "single-instance-exec",
		SourceType: inputSourcetype,
		Index:      "_internal",
		Source:     "single-instance-exec",
		Unbroken:   false,
		Done:       false,
	}
}

func (mi *ModularInput) getLoggingSourcetype() string {
	if mi.internalLogEvent != nil {
		return mi.internalLogEvent.SourceType
	}
	return "modinput:" + mi.defaultSourcetype
}
