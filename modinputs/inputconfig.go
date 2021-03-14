package modinputs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

/* This is the most generic form of the XML coming from splunkd via stdinput
In some cases, there might be NO <stanza> elements
<input>
  <server_host>myHost</server_host>
  <server_uri>https://127.0.0.1:8089</server_uri>
  <session_key>123102983109283019283</session_key>
  <checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs</checkpoint_dir>
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
</input>
*/

// inputConfig represents the parsed XML which Splunk provides
// on STDIN when starting the execution of a modular input
type inputConfig struct {
	XMLName       xml.Name `xml:"input"`
	Hostname      string   `xml:"server_host"`
	URI           string   `xml:"server_uri"`
	SessionKey    string   `xml:"session_key"`
	CheckpointDir string   `xml:"checkpoint_dir"`
	// there are multiple stanzas, which are all children of element <configuration>
	Stanzas []Stanza `xml:"configuration>stanza"`
}

// getInputConfigFromXML reads a XML-formatted configuration from the provided Reader,
// parses it and loads it within an inputConfig data structure
func getInputConfigFromXML(input io.Reader) (*inputConfig, error) {
	if input == nil {
		input = os.Stdin
	}
	buf := new(bytes.Buffer)
	if cnt, err := buf.ReadFrom(input); err != nil {
		return nil, fmt.Errorf("getInputConfigFromXML: %s.", err.Error())
	} else if cnt < 10 {
		// additionally check for data which is waaaay too small to be parsed.
		return nil, fmt.Errorf("getInputConfigFromXML: error xmldata too small.")
	}
	// parse and load the XML data within the inputConfig data structure
	ic := &inputConfig{}
	if err := xml.Unmarshal(buf.Bytes(), ic); err != nil {
		return nil, fmt.Errorf("getInputConfigFromXML: error when parsing input configuration xml. %s. %s", err.Error(), strings.ReplaceAll(buf.String(), "\n", "\\n"))
	}
	return ic, nil
}

/* This is the most generic form of the XML coming from splunkd via stdinput
Ref: https://docs.splunk.com/Documentation/Splunk/8.1.2/AdvancedDev/ModInputsValidate
In case of VALIDATION of parameters, the XML is different from the one received when executing the mod input :-/

<items>
    <server_host>myHost</server_host>
    <server_uri>https://127.0.0.1:8089</server_uri>
    <session_key>123102983109283019283</session_key>
    <checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs</checkpoint_dir>
    <item name="myScheme">
        <param name="param1">value1</param>
		<param name="param2">value2</param>
        <param_list name="param3">
            <value>value2</value>
            <value>value3</value>
            <value>value4</value>
        </param_list>
    </item>
</items>
*/

// ValidationConfig represents the parsed XML which Splunk provides
// on STDIN when starting the parameters validation of a modular input (command-line param: --validate-arguments)
type validationConfig struct {
	XMLName       xml.Name `xml:"items"`
	Hostname      string   `xml:"server_host"`
	URI           string   `xml:"server_uri"`
	SessionKey    string   `xml:"session_key"`
	CheckpointDir string   `xml:"checkpoint_dir"`
	// there can only be one validation item
	Item Stanza `xml:"item"`
}

// getValidationConfigFromXML reads a XML-formatted configuration from the provided "Reader" object,
// It parses the xml and loads it within the ValidationConfig data structure
// The XML MUST conform to the specification https://docs.splunk.com/Documentation/Splunk/8.1.2/AdvancedDev/ModInputsValidate
func getValidationConfigFromXML(input io.Reader) (*validationConfig, error) {
	if input == nil {
		input = os.Stdin
	}
	buf := new(bytes.Buffer)
	if cnt, err := buf.ReadFrom(os.Stdin); err != nil {
		return nil, fmt.Errorf("getValidationConfigFromXML: %s.", err.Error())
	} else if cnt < 10 {
		// additionally check for data which is waaaay too small to be parsed.
		return nil, fmt.Errorf("getValidationConfigFromXML: error xmldata too small.")
	}
	// parse and load the XML data within the ModInputConfig data structure
	vc := &validationConfig{}
	if err := xml.Unmarshal(buf.Bytes(), vc); err != nil {
		return nil, fmt.Errorf("getValidationConfigFromXML: error when parsing validation xml. %s. %s", err.Error(), strings.ReplaceAll(buf.String(), "\n", " "))
	}
	return vc, nil
}
