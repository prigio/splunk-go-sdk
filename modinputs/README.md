Library `git.cocus.com/bigdata/splunk-go-sdk/modinputs`
=====================================================

Offered functionality
---------------------

A modular input based on the `modinputs.ModularInput` offers: 

- structured parameter configuration
- automated generation of the xml-based _scheme_, when the modinput is started with the `--scheme` command-line parameter (expected by Splunk)
- automated validation of the xml-based input configuration when the modinput is started with the `--validate-arguments` command-line parameter (expected by Splunk);
- automated parsing of the xml-based input configuration when the script is started (expected by Splunk);
- automated generation of sample configuration files (use `--example-config` command-line parameter (additional functionality, practical for the developer)
 
Development workflow
--------------------

1. Install the library within your _go_ environment (`go get ...`) as described in the [readme.md](../readme.md) file at the root of this repository.
2. Create a struct of type `script := &modinputs.ModularInput{}`. This will be your main script.
3. Configure the modular input properties: `script.Title = "Hello world" ...`
4. Configure arguments `script.AddArgument(&modinputs.ModInputArg{....})`
5. (Optional) define a function to validate input parameters having signature
    `    func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error`
    Register the function within the modular input:
    `    script.Validate = validate`
6. define a function to stream the actual data to Splunk, having signature
    `    func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error`
    Register the function within the modular input:
    `    script.Stream = streamEvent`
7. Run the modular input: 
    `    err := script.Run()`

Examples
--------
See folder [examples](examples/) for more info.

Usage example
-------------
Install the library within your _go_ environment as described in the [readme.md](../readme.md) file at the root of this repository.

Create a file which will be the main script of your modular input:

```go

package main

import (
	"os"
	"time"

	"git.cocus.com/bigdata/splunk-go-sdk/modinputs"
)
// the actual function used to generate Splunk events based on configurations provided within the `stanza`
func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	time.Sleep(1 * time.Second)
	mi.Log("INFO", "'Hello' modular input internal logging")
	ev := mi.NewDefaultEvent(stanza)
	ev.Time = modinputs.GetEpochNow()
	ev.Data = "Hello " + stanza.GetParam("text")
	mi.WriteToSplunk(ev)
	return nil // fmt.Errorf("TEST ERRROR")
}

func main() {
	// Create an instance of the modinputs.ModularInput Struct
	script := &modinputs.ModularInput{}
    // name of the input type, appearing on Splunk UI
    script.Title = "Hello world input"
	script.Description = "This is a sample description for the test input"
	// !lowercase! name of the stanza used within splunk's inputs.conf
    script.StanzaName = "hello"
    script.Debug = false
    // see Splunk documentation for these
    script.UseExternalValidation = false
	script.UseSingleInstance = true
    
    // these two functions are responsible to
    // validate the inputs received by Splunk 
    // (only needed if script.UseExternalValidation=true)
    script.Validate = nil
    // actually generate the data that needs to be read by splunk
	script.Stream = streamEvents

    // define the expected arguments
	argText := &modinputs.ModInputArg{
		Name:             "text",
		Title:            "Text to input",
		Description:      "Description of text input",
		DataType:         modinputs.ArgDataTypeStr,
		RequiredOnCreate: true,
		RequiredOnEdit:   true,
		//		Validation: "",
		//		CustomValidation: "",
		//		CustomValidationErrMessage: "",
	}
    // argText.SetValidation(modinputs.ArgValidationIsPort)
    // add argument to script
	script.AddArgument(argText)

    // start the script!
	err := script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
	return
}
```

Compile the file and run it. 

Development
-----------

### Testing
Navigate to this directory from the command line and issue a `go test` command. This will execute all the tests provided and report on the ones failing.
