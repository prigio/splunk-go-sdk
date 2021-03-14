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

	"git.cocus.com/bigdata/splunk-go-sdk/modinputs"
)

func validate(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {
	//return fmt.Errorf("Some parameter was not validated")
	return nil
}

// streamEvents takes care of performing the desired checks
// and sends related data to Splunk
func streamEvents(mi *modinputs.ModularInput, stanza *modinputs.Stanza) error {

	ev := mi.NewDefaultEvent(stanza)
	ev.Time = modinputs.GetEpoch(time.Now())
	ev.Data = stanza.KVString()
	mi.WriteToSplunk(ev)

	return nil
}

func main() {
	script := &modinputs.ModularInput{}
	script.EnableDebug()
	script.Title = "Go SDK test"
	script.Description = "Tests if an the golang modular inputs sdk works properly"
	script.StanzaName = "gosdkcheck"
	script.UseExternalValidation = true
	script.UseSingleInstance = false
	script.Stream = streamEvents
	script.Validate = validate
	var err error

	script.AddArgument("param1", "Param1", "A string parameter", modinputs.ArgDataTypeStr, "", true, true)
	script.AddArgument("debug", "Debug", "", modinputs.ArgDataTypeBool, "", true, true)

	err = script.Run()
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
	return
}
