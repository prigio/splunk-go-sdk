package modinputs

import (
	"fmt"
	"strings"
	"testing"
)

func TestLoadConfigFromStdin(t *testing.T) {
	hostname := "myHost"
	uri := "https://127.0.0.1:8089"
	key := "123102983109283019283"
	checkpointDir := "/opt/splunk/var/lib/splunk/modinputs"
	inputXml := fmt.Sprintf(`<input>
  <server_host>%s</server_host>
  <server_uri>%s</server_uri>
  <session_key>%s</session_key>
  <checkpoint_dir>%s</checkpoint_dir>
  <configuration>
    <stanza name="myScheme://aaa">
        <param name="param1">value1</param>
        <param name="param2">value2</param>
        <param name="disabled">0</param>
        <param name="index">default</param>
    </stanza>
    <stanza name="myScheme://bbb">
        <param name="param1">value11</param>
        <param name="param2">value22</param>
        <param name="disabled">0</param>
        <param name="index">default</param>
    </stanza>
  </configuration>
</input>`, hostname, uri, key, checkpointDir)

	// actual start of testing
	c, err := getInputConfigFromXML(strings.NewReader(inputXml))
	if err != nil {
		t.Errorf("Testing LoadConfigFromStdin: %s", err.Error())
	}
	if c.Hostname != hostname {
		t.Errorf("Wrong hostname loaded: expected='%s', got='%s'", hostname, c.Hostname)
	}
	if c.URI != uri {
		t.Errorf("Wrong uri loaded: expected='%s', got='%s'", uri, c.URI)
	}
	if c.SessionKey != key {
		t.Errorf("Wrong SessionKey loaded: expected='%s', got='%s'", key, c.SessionKey)
	}
	if c.CheckpointDir != checkpointDir {
		t.Errorf("Wrong checkpointDir loaded: expected='%s', got='%s'", checkpointDir, c.CheckpointDir)
	}
	if len(c.Stanzas) != 2 {
		t.Errorf("Wrong number of stanzas loaded: expected=%d, got=%d", 2, len(c.Stanzas))
	}

	s0 := c.Stanzas[0]

	if s0.Index() != "default" {
		t.Errorf("Wrong index loaded: expected=%s, got=%s", "default", s0.Index())
	}
	if s0.Name != "myScheme://aaa" {
		t.Errorf("Wrong stanza name loaded: expected=%s, got=%s", "myScheme://aaa", s0.Name)
	}
	if s0.Param("param1") != "value1" {
		t.Errorf("Wrong value for parameter '%s' loaded: expected='%s', got='%s'", "param1", "value1", s0.Param("param1"))
	}
	if s0.Param("param2") != "value2" {
		t.Errorf("Wrong value for parameter '%s' loaded: expected='%s', got='%s'", "param2", "value2", s0.Param("param2"))
	}

}
