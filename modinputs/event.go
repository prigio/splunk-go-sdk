package modinputs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

// SplunkEvent is structure used to feed log data to splunk using the XML streaming mode.
// See: https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
type SplunkEvent struct {
	Data       string
	Time       float64
	SourceType string
	Index      string
	Host       string
	Source     string
	Stanza     string
	Unbroken   bool // unbroken events are events whose data attribute spans multiple <event> xml elements
	Done       bool // the last of a series of unbroken events uses <done/> to signalize the event is now complete
}

// NewSplunkEventFromDefaults returns a new SplunkEvent with the Done and Unbroken fields pre-initialized
// to a sensible standard (events are "broken", done is false)
func NewSplunkEventFromDefaults() *SplunkEvent {
	se := &SplunkEvent{}
	se.Unbroken = false
	se.Done = false
	return se
}

// writeOut is a private function which allows the modular input to skip counting the events emitted.
// useful for internal logging, which is not counter.
func (se *SplunkEvent) writeOut() (cnt int, err error) {
	var b []byte
	if b, err = se.xml(); err != nil {
		return -1, err
	}
	cnt, err = os.Stdout.Write(b)
	return cnt, err
}

// writeOutPlain is a private function which allows the modular input to skip counting the events emitted.
// useful for internal logging, which is not counted.
func (se *SplunkEvent) writeOutPlain(prependTime bool) (cnt int, err error) {
	var b []byte
	if b, err = se.string(prependTime); err != nil {
		return -1, err
	}
	cnt, err = os.Stdout.Write(b)
	return cnt, err
}

// XML generates a Splunk ModularInput compatible XML representation of the SplunkEvent.
//    See https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
func (se *SplunkEvent) xml() ([]byte, error) {
	// It would be easy to use xml.Marshal, but tests revelaed it takes 30% time to generate events than this method
	// for the xml needed to generate the Scheme it is not important, as that is only done once per execution.
	// But the events logging is much more time-critical.
	buf := new(bytes.Buffer)
	if se.Data == "" {
		return nil, fmt.Errorf("Events must have at least the data field set to be written to XML.")
	}

	// Prints start of the event, where attributes are optional: <event stanza="" unbroken="">
	buf.WriteString(`<event`)
	if se.Stanza != "" {
		buf.WriteString(` stanza="`)
		xml.EscapeText(buf, []byte(se.Stanza))
	}
	if se.Unbroken {
		// unbroken events are events whose data attribute spans multiple <event> xml elements
		buf.WriteString(` unbroken="1"`)
	}
	buf.WriteString(`">`)

	if se.Time > 0 {
		fmt.Fprintf(buf, "<time>%.3f</time>", se.Time)
	}
	if se.SourceType != "" {
		fmt.Fprintf(buf, "<sourcetype>%v</sourcetype>", se.SourceType)
	}
	if se.Index != "" {
		fmt.Fprintf(buf, "<index>%v</index>", se.Index)
	}
	if se.Host != "" {
		fmt.Fprintf(buf, "<host>%v</host>", se.Host)
	}
	if se.Source != "" {
		fmt.Fprintf(buf, "<source>%v</source>", se.Source)
	}

	buf.WriteString("<data>")
	xml.EscapeText(buf, []byte(se.Data))
	buf.WriteString("</data>")

	if se.Done {
		// the last of a series of unbroken events uses <done/> to signalize the event is now complete
		fmt.Fprint(buf, "<done/>")
	}

	buf.WriteString("</event>\n")
	return buf.Bytes(), nil
}

// string generates a plain-text representation of the SplunkEvent.
//    See https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsStream
func (se *SplunkEvent) string(prependTime bool) ([]byte, error) {
	// It would be easy to use xml.Marshal, but tests revelaed it takes 30% time to generate events than this method
	// for the xml needed to generate the Scheme it is not important, as that is only done once per execution.
	// But the events logging is much more time-critical.
	if se.Data == "" {
		return nil, fmt.Errorf("Events must have at least the data field set to be written out.")
	}
	buf := new(bytes.Buffer)
	if prependTime && se.Time > 0 {
		buf.WriteString(time.Unix(0, int64(se.Time*1000000000.0)).Format(time.RFC3339))
		buf.WriteString(" - ")
	}
	buf.WriteString(se.Data)

	return buf.Bytes(), nil
}
