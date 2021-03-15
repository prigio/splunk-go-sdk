package modinputs

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestXML(t *testing.T) {
	tn := time.Now()
	se := &SplunkEvent{
		Time:       tn,
		SourceType: "testsourcetype",
		Index:      "testindex",
		Host:       "testhost",
		Source:     "testsource",
		Stanza:     "testscheme://testinput",
	}

	_, err := se.xml()
	if err == nil {
		t.Error("SplunkEvent generated XML for an event without 'Data', which must actually raise an error.")
	}

	se.Data = "some test data"

	xml, err := se.xml()
	if err != nil {
		t.Errorf("SplunkEvent raised an error for a correctly structured event. %s", err.Error())
	}

	if !strings.Contains(string(xml), strconv.FormatFloat(float64(tn.Truncate(time.Millisecond).UnixNano())/1000000000.0, 'f', 3, 64)) {
		t.Errorf("SplunkEvent's XML does not contain epoch time. XML text: %s", string(xml))
	}
	if !strings.Contains(string(xml), "<data>some test data</data>") {
		t.Errorf("SplunkEvent's XML does not contain the actual event data. XML text: %s", string(xml))
	}
	if !strings.Contains(string(xml), "<index>testindex</index>") {
		t.Errorf("SplunkEvent's XML does not contain indication of index. XML text: %s", string(xml))
	}
	if !strings.Contains(string(xml), "<sourcetype>testsourcetype</sourcetype>") {
		t.Errorf("SplunkEvent's XML does not contain indication of index. XML text: %s", string(xml))
	}
	if !strings.Contains(string(xml), "<host>testhost</host>") {
		t.Errorf("SplunkEvent's XML does not contain indication of index. XML text: %s", string(xml))
	}
	if !strings.Contains(string(xml), "<source>testsource</source>") {
		t.Errorf("SplunkEvent's XML does not contain indication of index. XML text: %s", string(xml))
	}
}

func TestXMLSpecialChars(t *testing.T) {
	tn := time.Now()
	se := &SplunkEvent{
		Time:       tn,
		SourceType: "Äpp",
		Data:       "Dieser String enthält einen Umlaut",
	}

	xml, err := se.xml()
	if err != nil {
		t.Errorf("SplunkEvent raised an error for a correctly structured event. %s", err.Error())
	}

	if !strings.Contains(xml, "Äpp") {
		t.Errorf("SplunkEvent's XML does not contain the expected 'Äpp' value within sourcetype. XML text: %s", xml)
	}
	if !strings.Contains(xml, "enthält") {
		t.Errorf("SplunkEvent's XML does not contain the expected 'enthält' word within the data. XML text: %s", xml)
	}
}

func BenchmarkXML(b *testing.B) {
	se := &SplunkEvent{
		Time:       time.Now(),
		SourceType: "testsourcetype",
		Index:      "testindex",
		Host:       "testhost",
		Source:     "testsource",
		Stanza:     "testscheme://testinput",
		Data:       "some test data",
	}

	// run the xml() function b.N times
	for n := 0; n < b.N; n++ {
		if n%3 == 0 {
			se.Time = time.Now()
		}
		se.xml()
	}

}
