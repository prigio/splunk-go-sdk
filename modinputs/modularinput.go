package modinputs

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type StreamingFunc func(*ModularInput, *Stanza) error
type ValidationFunc func(*ModularInput, *Stanza) error

// ModularInput is the main structure defining how a modular input looks like.
// Through composition, configuration and scheme elements are assembled within it
type ModularInput struct {
	ModInputConfig // Unnamed sub-structs: their fields and methods get imported directly within ModularInput
	ModInputScheme
	Debug bool // This debug setting is meant for facilitating development and is not configurable by a user through splunk's inputs.conf
	// private variables
	internalLogEvent               *SplunkEvent //this is used to setup a standardized event using for logging to index=_internal. If this is not nil, internal loggin is performed through SplunkEvents written on Stdout instead of plain output on Stderr
	cntDataEventsGeneratedbyStanza int64        // counter of data events emitted by the stanza being currently processed (internal loggin is excluded)
	cntDataEventsGeneratedTotal    int64        // counter of data events emitted in total (internal loggin is excluded)
	// function used to validate data
	Validate ValidationFunc
	// function used to stream generated data
	Stream StreamingFunc
}

// Log writes a log so that it can be read by splunk without being interpreted as an actual event generated by the script
func (mi *ModularInput) Log(level string, message string) (err error) {
	level = strings.ToUpper(level)
	if level == "WARNING" {
		// Typical error, just manage it...
		level = "WARN"
	}
	if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" && level != "FATAL" {
		fmt.Fprintf(os.Stderr, "ERROR - ModularInput.Log invoked with invalid level parameter. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'\n", level)
		return fmt.Errorf("ModularInput.Log: invalid value of 'level' provided. Accepted: DEBUG, INFO, WARN, ERROR, FATAL. Provided: '%s'", level)
	}
	if level != "DEBUG" || (level == "DEBUG" && mi.Debug) {
		// do not do anything if debug is not enabled
		t := time.Now().Round(time.Millisecond)
		if mi.internalLogEvent != nil {
			mi.internalLogEvent.Time = GetEpoch(t)
			//time.Format uses a string with such parameters to define the output format: Mon Jan 2 15:04:05 -0700 MST 2006
			mi.internalLogEvent.Data = fmt.Sprintf("[%s] %s - %s", t.Format("2006-01-02 15:04:05.000 -0700"), level, message)
			// using writeOut() to skip counting the events, as we do not want to count the internal logs...
			mi.internalLogEvent.writeOut()
		} else {
			// XML-based logging has not yet been activated: using STDERR instead
			_, err = fmt.Fprintf(os.Stderr, "%s - ModularInput: %s\n", level, message)
		}
	}
	return err
}

func (mi *ModularInput) WriteToSplunk(se *SplunkEvent) (cnt int, err error) {
	var b []byte
	if b, err = se.xml(); err != nil {
		return -1, err
	}
	// increase the counter of the generated events
	mi.cntDataEventsGeneratedbyStanza++
	mi.cntDataEventsGeneratedTotal++
	cnt, err = os.Stdout.Write(b)
	return cnt, err
}

// loadConfigFromStdin reads a XML-formatted configuration from stdin using the facilities
// provided by inputconfig.go
func (mi *ModularInput) loadConfigFromStdin() (err error) {
	var n int64
	buf := new(bytes.Buffer)
	n, err = buf.ReadFrom(os.Stdin)
	if err != nil {
		mi.Log("FATAL", "Error while reading from stdin. "+err.Error())
		return err
	}

	mi.Log("DEBUG", fmt.Sprintf("Read from Stdin %d bytes\n", n))
	mi.Log("DEBUG", "Parsing XML config")

	if mi.ParseConfig(buf.Bytes()) != nil {
		mi.Log("FATAL", "Error while parsing stdin. "+err.Error())
		return err
	}
	return nil
}

// Run is the main function that you, as a modular input script developer must invoke to start the actual processing.
func (mi *ModularInput) Run() (err error) {
	mi.Log("DEBUG", fmt.Sprintf("ModularInput.Run started. Cmd-line parameters: '%s'", strings.Join(os.Args, " ")))
	if len(os.Args) == 1 {
		// Read XML configs
		// Populates infos about the configuration Stanzas
		if err = mi.loadConfigFromStdin(); err != nil {
			return err
		}
		return mi.runStreaming()
	} else if os.Args[1] == "--scheme" {
		return mi.runScheme()
	} else if os.Args[1] == "--validate-arguments" {
		// Read XML configs
		if err := mi.loadConfigFromStdin(); err != nil {
			return err
		}
		return mi.runValidation()
	} else if os.Args[1] == "--example-config" {
		// This is not part of the standard spluk interface, but it is useful for users to print-out the expected splunk configurations
		fmt.Println(mi.ExampleConf())
		return nil
	} else if os.Args[1] == "-?" || os.Args[1] == "-h" || len(os.Args) >= 2 {
		fmt.Printf(`Usage: %s [--scheme|--validate-arguments|--example-config]
- invoked with no command-line parameters, start the data streaming. Provide configurations on STDIN using the XML format specified at:
		https://docs.splunk.com/Documentation/SplunkCloud/8.1.2011/AdvancedDev/ModInputsScripts#Read_XML_configuration_from_splunkd
- --scheme: prints out the XML scheme definition. See Splunk documentation
- --validate-arguments: validate the parameters provided on STDIN in XML format. See Splunk documentation
- --example-config: Print an example of inputs.conf configuration for this modular input and exit.
`, path.Base(os.Args[0]))
		return nil
	}

	return nil
}

func (mi *ModularInput) runStreaming() (err error) {
	// these two vars are used to track the duration of the overall streaming function
	var duration time.Duration
	mi.Log("DEBUG", "Starting 'runStreaming' function")
	if mi.Stream == nil {
		mi.Log("FATAL", "No streaming function specified")
		return fmt.Errorf("FATAL: no streaming function specified")
	}

	hadErrors := false
	var stanza Stanza
	streamingStartTime := time.Now()

	if len(mi.Stanzas) == 0 {
		mi.Log("FATAL", "No configurazion stanzas are present within input configuration. Nothing to be done.")
		return fmt.Errorf("ERROR: No input configuration stanzas")
	}

	fmt.Println("<stream>")          // Setup the XML streaming mode
	defer fmt.Println("\n</stream>") // close XML streaming mode when returning

	for _, stanza = range mi.Stanzas {
		// reset the counter of events for the new stanza
		mi.cntDataEventsGeneratedbyStanza = 0
		stanzaStartTime := time.Now()
		//Start logging intenrnal messages as SplunkEvents instead of using plaintext on Stderror
		mi.setupEventBasedInternalLogging(&stanza)
		mi.Log("INFO", fmt.Sprintf(`Starting streaming for stanza="%s"`, stanza.Name))
		err := mi.Stream(mi, &stanza)
		duration = time.Now().Sub(stanzaStartTime)
		if err != nil {
			hadErrors = true
			mi.Log("ERROR", fmt.Sprintf(`Execution failed for stanza="%s" duration_s=%.03f cnt_events=%d error="%s"`, stanza.Name, duration.Seconds(), mi.cntDataEventsGeneratedbyStanza, err.Error()))
		} else {
			mi.Log("INFO", fmt.Sprintf(`Execution succeeded for stanza="%s" duration_s=%.03f cnt_events=%d`, stanza.Name, duration.Seconds(), mi.cntDataEventsGeneratedbyStanza))
		}
	}
	// Remove the stanza-specific logging settings from the event used for internal logging
	mi.resetEventBasedInternalLogging()
	duration = time.Now().Sub(streamingStartTime)
	if hadErrors {
		err = fmt.Errorf("Execution of one or more stanzas failed")
		mi.Log("WARN", fmt.Sprintf("Script execution completed with errors duration_s=%.03f cnt_events=%d", duration.Seconds(), mi.cntDataEventsGeneratedTotal))
	} else {
		mi.Log("INFO", fmt.Sprintf("Script execution succeeded duration_s=%.03f cnt_events=%d", duration.Seconds(), mi.cntDataEventsGeneratedTotal))
	}

	return err
}

func (mi *ModularInput) runValidation() error {
	var stanza Stanza
	mi.Log("DEBUG", "Starting argument validation")

	if !mi.UseExternalValidation {
		mi.Log("WARN", "Script invoked with --validate-arguments command-line arguments but configured to NOT use external validation. Exiting")
		return nil
	}
	if mi.UseExternalValidation && mi.Validate == nil {
		mi.Log("ERROR", "Script configured to use external validation, but no validation function was specified")
		return nil
	}
	for _, stanza = range mi.Stanzas {
		err := mi.Validate(mi, &stanza)
		if err != nil {
			mi.Log("ERROR", fmt.Sprintf(`Validation of parameters for stanza="%s" failed with error="%s"`, stanza.Name, err.Error()))
			// Splunk specification requires to write the validation errors on STDOUT
			// See: https://docs.splunk.com/Documentation/SplunkCloud/8.1.2011/AdvancedDev/ModInputsScripts#Create_a_modular_input_script
			fmt.Printf("%s\n", err.Error())
			return err
		}
	}
	return nil
}

func (mi *ModularInput) runScheme() error {
	mi.Log("DEBUG", "starting --scheme action")

	scheme, err := mi.PrintXMLScheme()
	if err != nil {
		mi.Log("FATAL", "Error during scheme generation. "+err.Error())
		return err
	}
	fmt.Println(string(scheme))
	return nil
}

// NewDefaultEvent provides a template for the SplunkEvent to be used to log actual data to be imported to Splunk
func (mi *ModularInput) NewDefaultEvent(stanza *Stanza) (ev *SplunkEvent) {
	if stanza != nil {
		//mi.Log("DEBUG", fmt.Sprintf("NewDefaultEvent: stanza is NOT nil. %v", stanza.Params))
		// NOT specifying Data intentionally
		ev = &SplunkEvent{
			// Anything can be overridded by the actual script
			Time:       GetEpochNow(),
			Stanza:     stanza.Name,
			SourceType: stanza.GetSourcetype(),
			Index:      stanza.GetIndex(),
			Host:       stanza.GetHost(),
			Source:     stanza.GetSource(),
			Unbroken:   false,
			Done:       false,
		}
	} else {
		// If no configurations are present, we basically just return a generic event
		ev = &SplunkEvent{
			Time:     GetEpochNow(),
			Unbroken: false,
			Done:     false,
		}
	}
	return ev
}

// setupEventBasedInternalLogging configures logging to be performed through SplunkEvent events written to index=_internal instead of using plain text on standard-err.
// Before activating this, the user is informed with a WARN message on StdErr saying which source/sourcetype is being used for logging purposes from now on
// This function can only be invoked when an active configuration has been provided in input, so, when we start streaming events.
func (mi *ModularInput) setupEventBasedInternalLogging(stanza *Stanza) {
	if stanza != nil {
		inputSourcetype := "modinput:" + strings.Split(stanza.Name, "://")[0]
		mi.Log("INFO", fmt.Sprintf(`Logging using 'index=_internal sourcetype="%s" source="%s" for the rest of this execution`, inputSourcetype, stanza.Name))
		mi.internalLogEvent = &SplunkEvent{
			// NOT specifying Data and Host intentionally
			Time:       GetEpochNow(),
			Stanza:     stanza.Name,
			SourceType: inputSourcetype,
			Index:      "_internal",
			Source:     stanza.Name,
			Unbroken:   false,
			Done:       false,
		}
	} else {
		mi.Log("WARN", "Function setupEventBasedInternalLogging() called without a stanza being specified, this might be an error within the library")
		inputSourcetype := "modinput:" + path.Base(os.Args[0])
		mi.Log("INFO", fmt.Sprintf(`Logging using 'index=_internal sourcetype="%s" source="%s" for the rest of this execution`, inputSourcetype, os.Args[0]))
		mi.internalLogEvent = &SplunkEvent{
			// NOT specifying Data and Host intentionally
			Time:       GetEpochNow(),
			Stanza:     "",
			SourceType: inputSourcetype, // use filename within the sourcetype in absence of a better option
			Index:      "_internal",
			Source:     os.Args[0],
			Unbroken:   false,
			Done:       false,
		}
	}
}

// resetEventBasedInternalLogging removes the stanza-specific source from the internally logged events
// and resets it to the filename of the script
func (mi *ModularInput) resetEventBasedInternalLogging() {
	if mi.internalLogEvent != nil {
		mi.internalLogEvent.Source = path.Base(os.Args[0])
	}
}
