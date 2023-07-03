package main

import (
	"fmt"
	"os"
	"time"

	"github.com/prigio/splunk-go-sdk/alertactions"
)

func doAlert(aa *alertactions.AlertAction) error {

	textPar, err := aa.GetParam("text")
	if err != nil {
		return err
	}
	targetPar, err := aa.GetParam("target")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(targetPar.GetValue(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	t := time.Now().Round(time.Millisecond)
	aa.Log("INFO", "Writing to file '%s'", targetPar.GetValue())
	fmt.Fprintf(f, "[%s] - %s\n", t.Format("2006-01-02 15:04:05.000 -0700"), textPar.GetValue())
	return nil
}

func main() {
	// Prepare the script
	script, err := alertactions.New("helloWorld", "Hello world alert", "Writes the text provided as a parameter into the chosen file.", "missingIcon.png")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(1)
	}

	script.AddParam("text", "Text", "Content to be written in the target file", "Hello World!", "Your text here", alertactions.ParamTypeTextArea, true)
	script.AddParam("target", "Target file", "File into which the text should be written", "./alertAction_helloWorld.out", "Your file here", alertactions.ParamTypeText, true)

	script.RegisterAlertFunc(doAlert)

	// Start actual execution
	err = script.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		// this will NOT run deferred function, so in case we have any, need to take care about that. Simply: do NOT use such functions within the main() ;-)
		os.Exit(2)
	}
}
