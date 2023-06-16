package modinputs

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// GetEpochNow returns an the current time as Epoch, expressed in seconds with a decimal part
func GetEpochNow() float64 {
	return float64(time.Now().UnixNano()) / 1000000000.0
}

// GetEpoch returns an Epoch timestamp with millisecond precision starting from time t
func GetEpoch(t time.Time) float64 {
	return float64(t.UnixNano()) / 1000000000.0
}

/*
	 askForInput promts the user to provide a value via StdIn
		if isPassword=true, no local echo to the console is provided
		if isPassword=false and the provided input is empty, the default value is returned instead
*/
func askForInput(prompt string, defaultVal string, isPassword bool) string {
	prompt = strings.Trim(prompt, ": ")
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	if isPassword {
		if bytepw, err := term.ReadPassword(int(syscall.Stdin)); err != nil {
			return ""
		} else {
			return string(bytepw)
		}
	} else {
		if text, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil || text == "\n" {
			return defaultVal
		} else {
			return strings.Replace(text, "\n", "", -1)
		}
	}
}
