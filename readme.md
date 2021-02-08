# Splunk-Go - A simple framework to interact with Splunk using Golang

Reason behind this: sometime you want to deploy python scripts, but there is no python available on the targets (say, a thousand windows servers using the Splunk UF). In those cases, you need to deploy something else. 

Golang is probably the best choice for these use cases as it allows to packate scripts without external library dependencies.

However, how do you make these scripts talk with Splunk?! Enters this repository. 

This repository provides:

- **a framework & library to build modular inputs** - this is already production grade
- **a (simple) splunkd client** - so that your scripts can perform basic communication with splunkd on port `8089`. **In development**.
- a custom alerts framework - if needs arises. 


## Usage
### Prepare build pipeline
Import this repository within your golang code

filename.go: 

```
import (
	"git.cocus.com/bigdata/splunk-go/modinputs"
    // if needed
    "git.cocus.com/bigdata/splunk-go/client"
)
```

Note, as this is a PRIVATE repository, you need to:

1. Get a build token from the team managing this repository. 
    The token is configured within git.cocus.com. 
2. Configure your build pipeline to use that token when downloading the library.
    Create a file `~/.netrc` with this content

    ```
        machine git.cocus.com login go-build password <build-token>
    ```
3. Configure your build pipeline with the following environment variable:
    `    GOPRIVATE=git.cocus.com/*`
4. Test your build pipeline. You can manually execute `go get git.cocus.com/bigdata/splunk-go/modinputs` and see if it worked.

### Use the library

#### modinputs

Within your main go code: `main()`: 

```go
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


## Various resources
These are links for future reference, but not directly relevant to this library.

https://eli.thegreenplace.net/2019/simple-go-project-layout-with-modules/

https://www.docker.com/blog/containerize-your-go-developer-environment-part-1/

