package main

import (
	"os"
	"time"

	"github.com/prigio/splunk-go-sdk/modinputs"
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
	script.SetDefaultSourcetype("helloworld")
	// Start actual execution
	err := script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
}
