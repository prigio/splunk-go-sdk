package main

/*
A Splunk ModularInput performing HTTP(s) checks

Compile as
(within a GO docker container)
	GOOS=linux go build -o /usr/src/bin/linux/;
	GOOS=darwin go build -o /usr/src/bin/darwin/;
	GOOS=windows go build -o /usr/src/bin/windows/;
Copy executables within the app

cp bin/linux/* apps/ta-netmon-modinput/linux_x86_64/bin/
cp bin/darwin/* apps/ta-netmon-modinput/darwin_x86_64/bin/
cp bin/windows/* apps/ta-netmon-modinput/windows_x86_64/bin/

*/
import (
	"os"
	"time"

	"github.com/prigio/splunk-go-sdk/modinputs"
)

func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	//return fmt.Errorf("Some parameter was not validated")
	return nil
}

// streamEvents takes care of performing the desired checks
// and sends related data to Splunk
func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {

	ev := mi.NewDefaultEvent(stanza)
	ev.Time = time.Now()
	ev.Data = stanza.KVString()
	mi.WriteToSplunk(ev)

	return nil
}

func main() {
	script := &modinputs.ModularInput{
		Title:                 "Go SDK test",
		Description:           "Tests if an the golang modular inputs sdk works properly",
		StanzaName:            "gosdkcheck",
		UseExternalValidation: true,
		UseSingleInstance:     false,
		Stream:                streamEvents,
		Validate:              validate,
	}

	script.EnableDebug()
	script.AddArgument("param1", "Param1", "A string parameter", modinputs.ArgDataTypeStr, "", true, true)
	script.AddArgument("debug", "Debug", "", modinputs.ArgDataTypeBool, "", true, true)
	// Start actual execution
	err := script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
	return
}
