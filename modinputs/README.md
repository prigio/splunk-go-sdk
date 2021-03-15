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
4. Configure arguments `script.AddArgument(name, Title, description, ....)`
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

func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	time.Sleep(1 * time.Second)
	mi.Log("INFO", "'Hello' modular input internal logging: starting 'streamEvents' for stanza=%s", stanza.Name)
	ev := mi.NewDefaultEvent(stanza)
	ev.Time = time.Now()
	ev.Data = "Hello " + stanza.Param("text")
	mi.WriteToSplunk(ev)
	return nil
}

func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	mi.Log("INFO", "Within custom validation function")
	return nil
}

func main() {
	// Prepare the script
	script := &modinputs.ModularInput{
		Title:                 "Hello world input",
		Description:           "This is a sample description for the test input",
		StanzaName:            "hello",
		UseExternalValidation: true,
		UseSingleInstance:     true,
		Stream:                streamEvents,
		Validate:              validate,
	}
	script.EnableDebug()
	script.AddArgument("text", "Text to input", "Description of text input", modinputs.ArgDataTypeStr, "", true, true)

	// Start actual execution
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
