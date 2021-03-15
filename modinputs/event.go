package modinputs

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SplunkEvent is structure used to feed log data to splunk using the XML streaming mode.
// See: https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
type SplunkEvent struct {
	Unbroken       bool // unbroken events are events whose data attribute spans multiple <event> xml elements
	Done           bool // the last of a series of unbroken events uses <done/> to signalize the event is now complete
	Time           time.Time
	cachedTime     time.Time
	Data           string
	SourceType     string
	Index          string
	Host           string
	Source         string
	Stanza         string
	cachedEpochStr string
}

// EpochTime reads the Time parameters of SplunkEvent se and returns an floating point
// representation of the time expressed as Epoch with millisecond precision
func (se *SplunkEvent) epochTimeStr() string {
	// checks if the event timestamp changed wrt the cached version
	if se.Time != se.cachedTime {
		// regenerate the cached epoch representation
		se.cachedTime = se.Time
		se.cachedEpochStr = strconv.FormatFloat(float64(se.Time.Truncate(time.Millisecond).UnixNano())/1000000000.0, 'f', 3, 64)
	}
	return se.cachedEpochStr
}

// writeOut is a private function which allows the modular input to skip counting the events emitted.
// useful for internal logging, which is not counter.
func (se *SplunkEvent) writeOut() (cnt int, err error) {
	if xmlStr, err := se.xml(); err != nil {
		return -1, err
	} else {
		return os.Stdout.WriteString(xmlStr)
	}
}

// writeOutPlain is a private function which allows the modular input to skip counting the events emitted.
// useful for internal logging, which is not counted.
func (se *SplunkEvent) writeOutPlain(prependTime bool) (cnt int, err error) {
	if plainStr, err := se.string(prependTime); err != nil {
		return -1, err
	} else {
		cnt, _ = os.Stdout.WriteString(plainStr)
		os.Stdout.WriteString("\n")
		return cnt + 1, nil
	}
}

/*
// getSplunkAsterisksHeader returns a "***SPLUNK*** host=xxx sourcetype=yyy..." header
// which can be used by splunk when reading plain-text log files to change the specified
// attributes for the subsequent events.
// See: https://docs.splunk.com/Documentation/Splunk/8.0.1/Admin/Propsconf#Header_Processor_configuration
func (se *SplunkEvent) getSplunkAsterisksHeader() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(128)
	buf.WriteString("***SPLUNK***")
	if se.Host != "" {
		buf.WriteString(` host="` + se.Host + `"`)
	}
	if se.Source != "" {
		buf.WriteString(` source="` + se.Source + `"`)
	}
	if se.SourceType != "" {
		buf.WriteString(` sourcetype="` + se.SourceType + `"`)
	}
	if se.Index != "" {
		buf.WriteString(` index="` + se.Index + `"`)
	}
	return buf.String()
}
*/

// XML generates a Splunk ModularInput compatible XML representation of the SplunkEvent.
//    See https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
func (se *SplunkEvent) xml() (string, error) {
	// It would be easy to use xml.Marshal, but tests revelaed it takes 30% time to generate events than this method.
	// For the xml needed to generate the Scheme it is not important, as that is only done once per execution.
	// But the events logging is much more time-critical.
	if se.Data == "" {
		return "", fmt.Errorf("Events must have at least the data field set to be written to XML.")
	}

	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	// Prints start of the event, where attributes are optional: <event stanza="" unbroken="">
	buf.WriteString(`<event`)
	if se.Stanza != "" {
		buf.WriteString(` stanza="`)
		xml.EscapeText(buf, []byte(se.Stanza))
	}
	if se.Unbroken {
		// unbroken events are events whose data attribute spans multiple <event> xml elements
		buf.WriteString(`" unbroken="1`)
	}
	buf.WriteString(`">`)

	// NOTE OF PERFORMANCE
	// After profiling this code
	// 		go test -bench=. -cpuprofile ./cpuprof5.dat
	// 		go tool pprof -http="0.0.0.0:8081" cpuprof5.dat
	// it is faster to call buf.WriteString multiple times instead of doing buf.WriteString(<time> + se.epoch... + </time>)
	// reason is there are less string allocations in the former case
	if !se.Time.IsZero() {
		//fmt.Fprintf(buf, "<time>%.3f</time>", float64(se.Time.Truncate(time.Millisecond).UnixNano())/1000000000.0)
		//buf.WriteString("<time>" + strconv.FormatFloat(se.EpochTime(), 'f', 3, 64) + "</time>")
		//buf.WriteString("<time>" + strconv.FormatFloat(float64(se.Time.Truncate(time.Millisecond).UnixNano())/1000000000.0, 'f', 3, 64) + "</time>")
		buf.WriteString("<time>")
		buf.WriteString(se.epochTimeStr())
		buf.WriteString("</time>")
	}
	if se.SourceType != "" {
		buf.WriteString("<sourcetype>")
		buf.WriteString(se.SourceType)
		buf.WriteString("</sourcetype>")
	}
	if se.Index != "" {
		buf.WriteString("<index>")
		buf.WriteString(se.Index)
		buf.WriteString("</index>")
	}
	if se.Host != "" {
		buf.WriteString("<host>")
		buf.WriteString(se.Host)
		buf.WriteString("</host>")
	}
	if se.Source != "" {
		buf.WriteString("<source>")
		buf.WriteString(se.Source)
		buf.WriteString("</source>")
	}
	buf.WriteString("<data>")
	xml.EscapeText(buf, []byte(se.Data))
	buf.WriteString("</data>")
	if se.Done {
		// the last of a series of unbroken events uses <done/> to signalize the event is now complete
		buf.WriteString("<done/>")
	}
	buf.WriteString("</event>")

	return buf.String(), nil
}

// string generates a plain-text representation of the SplunkEvent.
//    See https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
func (se *SplunkEvent) string(prependTime bool) (string, error) {
	// It would be easy to use xml.Marshal, but tests revelaed it takes 30% time to generate events than this method
	// for the xml needed to generate the Scheme it is not important, as that is only done once per execution.
	// But the events logging is much more time-critical.
	if se.Data == "" {
		return "", fmt.Errorf("Events must have at least the data field set to be written out.")
	}

	if prependTime && !se.Time.IsZero() {
		return se.Time.Format(time.RFC3339) + " - " + se.Data, nil
	} else {
		return se.Data, nil
	}
}
