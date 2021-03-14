package main

import (
	"os"
	"time"

	"git.cocus.com/bigdata/splunk-go-sdk/modinputs"
)

func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	time.Sleep(1 * time.Second)
	mi.Log("INFO", "'Hello' modular input internal logging")
	ev := mi.NewDefaultEvent(stanza)
	ev.Time = modinputs.GetEpochNow()
	ev.Data = "Hello " + stanza.Param("text")
	mi.WriteToSplunk(ev)
	return nil // fmt.Errorf("TEST ERRROR")
}

func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	mi.Log("INFO", "Within custom validation function")
	return nil
}

func main() {
	// Prepare the script
	script := &modinputs.ModularInput{}
	script.Title = "Hello world input"
	script.Description = "This is a sample description for the test input"
	script.StanzaName = "hello"
	script.UseExternalValidation = true
	script.UseSingleInstance = true
	script.EnableDebug()
	script.Stream = streamEvents
	script.Validate = validate

	script.AddArgument("text", "Text to input", "Description of text input", modinputs.ArgDataTypeStr, "", true, true)

	err := script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
	return
}
