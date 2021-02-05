package main

import (
	"os"
	"time"

	"git.cocus.com/bigdata/splunk-go/modinputs"
)

func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	time.Sleep(1 * time.Second)
	mi.Log("INFO", "'Hello' modular input internal logging")
	ev := mi.NewDefaultEvent(stanza)
	ev.Time = modinputs.GetEpochNow()
	ev.Data = "Hello " + stanza.GetParam("text")
	mi.WriteToSplunk(ev)
	return nil // fmt.Errorf("TEST ERRROR")
}

/*
func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	return fmt.Errorf("VALIDATION ERROR")
}
*/
func main() {
	// Prepare the script
	script := &modinputs.ModularInput{}
	script.Title = "Hello world input"
	script.Description = "This is a sample description for the test input"
	script.StantaName = "hello"
	script.UseExternalValidation = false
	script.UseSingleInstance = true
	script.Debug = false
	script.Stream = streamEvents
	script.Validate = nil

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
	script.AddArgument(argText)

	err := script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
	return
}
