package main

import (
	"os"
	"sync"
	"time"

	"github.com/prigio/splunk-go-sdk/v2/modinputs"
)

// The main function of the modular input, actually processing the data for all the stanzas.
func streamEvents(mi *modinputs.ModularInput, stanzas []modinputs.Stanza) error {
	var wg sync.WaitGroup

	for i, s := range stanzas {
		wg.Add(1)
		// attention to capturing of loop variables within loops and goroutines!
		// do not pass s as reference!
		go func(seq int, st modinputs.Stanza) {
			if seq > 0 {
				time.Sleep(5 * time.Second)
			}
			// add one log for administrators, within _internal index
			mi.Log("INFO", "'Hello' modular input internal logging: starting 'streamEvents' for seq=%d stanza=%s", seq, st.Name)
			ev := modinputs.NewEvent(&st)
			ev.Time = time.Now()
			ev.Data = "Hello " + st.Param("text")
			// emit a proper "collected" event to splunk
			mi.WriteToSplunk(ev)
			wg.Done()
		}(i, s)
	}

	wg.Wait()
	mi.Log("INFO", "All stanzas processed in parallel")
	return nil
}

// Function to validate the parameters of a stanza.
// Return a non-nil error to communicate to splunk a failure in the validation.
func validate(mi *modinputs.ModularInput, stanza modinputs.Stanza) error {
	mi.Log("INFO", "Within custom validation function for stanza %s", stanza.Name)
	for _, p := range stanza.Params {
		mi.Log("INFO", "Param '%s' has value %s", p.Name, p.Value)
	}
	//return fmt.Errorf("error")
	return nil
}

func main() {
	// Prepare the script
	script, _ := modinputs.New("helloWorld", "Hello World input", "This is a sample description for the test input")
	script.EnableDebug()
	script.SetDefaultSourcetype("helloworld")

	script.RegisterNewParam("text", "Text to input", "Description of text input", "", "string", true, false)

	script.RegisterValidationFunc(validate)
	script.RegisterStreamingFuncSingleInstance(streamEvents)

	// Start actual execution
	err := script.Run(os.Args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}
}
