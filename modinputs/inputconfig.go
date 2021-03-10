package modinputs

import (
	"bytes"
	"encoding/xml"
	"fmt"
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

type ModInputConfig struct {
	XMLName       xml.Name `xml:"input"`
	Hostname      string   `xml:"server_host"`
	URI           string   `xml:"server_uri"`
	SessionKey    string   `xml:"session_key"`
	CheckpointDir string   `xml:"checkpoint_dir"`
	// there are multiple stanzas, which are all children of element <configuration>
	Stanzas []Stanza `xml:"configuration>stanza"`
}

// LoadConfigFromStdin reads a XML-formatted configuration from stdin,
// parses it and loads it within the ModInputConfig data structure
// Returns the number of bytes read and an error
func (c *ModInputConfig) LoadConfigFromStdin() (cnt int64, err error) {
	buf := new(bytes.Buffer)
	if cnt, err = buf.ReadFrom(os.Stdin); err != nil {
		return cnt, err
	} else if cnt < 10 {
		// additionally check for data which is waaaay too small to be parsed.
		return cnt, fmt.Errorf("LoadConfigFromStdin: error xmldata too small.")
	}
	// parse and load the XML data within the ModInputConfig data structure
	if err = xml.Unmarshal(buf.Bytes(), &c); err != nil {
		return cnt, fmt.Errorf("LoadConfigFromStdin: error when parsing xml configuration. %s", err.Error())
	}
	return cnt, nil
}

type Stanza struct {
	XMLName xml.Name `xml:"stanza"`
	Name    string   `xml:"name,attr"` // name attribute of the stanza
	App     string   `xml:"app,attr"`  // application where the configuration is defined
	Params  []Param  `xml:"param"`
}

// GetParam scans the stanza s parameters and returns the param with the specified name. If not found, returns ""
func (s *Stanza) GetParam(name string) (ret string) {
	for _, p := range s.Params {
		if strings.ToLower(p.Name) == name {
			return p.Value
		}
	}
	return ""
}

// GetSourcetype returns the sourcetype configured for the stanza s
func (s *Stanza) GetSourcetype() (ret string) {
	return s.GetParam("sourcetype")
}

// GetHost returns the host configured for the stanza s
func (s *Stanza) GetHost() (ret string) {
	return s.GetParam("host")
}

// GetSource returns the sourcsourceetype configured for the stanza s
func (s *Stanza) GetSource() (ret string) {
	return s.GetParam("source")
}

// GetIndex returns the index configured for the stanza s
func (s *Stanza) GetIndex() (ret string) {
	return s.GetParam("index")
}

type Param struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"` // name attribute of the param
	Value   string   `xml:",chardata"` // access the textual data of the param value
}

/* Usage sample
s := `<input>
  <server_host>myHost</server_host>
  <server_uri>https://127.0.0.1:8089</server_uri>
  <session_key>123102983109283019283</session_key>
  ....`

config, err := modinputs.ParseInputConfig([]byte(s))
if err != nil {
	fmt.Println(err.Error())
} else {
	fmt.Printfln("SessionKey: %s\n", config.SessionKey)
	for _, stanza := range config.Configuration.Stanzas {
		fmt.Printf("Name:%s\n", stanza.Name)
		for _, param := range stanza.Params {
			fmt.Printf("Name:%s, val:%s\n", param.Name, param.Value)
		}
	}
}
*/
