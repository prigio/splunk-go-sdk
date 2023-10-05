package modinputs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Stanza represents the configuration for a modular input found in inputs.conf
type Stanza struct {
	// name attribute of the stanza, in form '<scheme>://<input name>'
	Name string `xml:"name,attr"`
	// App in which the stanza is defined
	App string `xml:"app,attr"`
	// List of parameters for the stanza
	Params []Param `xml:"param"`
	// List of list-parameters for the stanza (may only be used by the validation xml)
	ParamLists []ParamList `xml:"param_list"`
}

// KVString returns a k=v based representation of the configurations present within the Stanza
func (s *Stanza) KVString() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, `stanza="%s" app="%s"`, s.Name, s.App)
	for _, p := range s.Params {
		fmt.Fprintf(buf, ` %s="%s"`, p.Name, p.Value)
	}
	for _, p := range s.ParamLists {
		fmt.Fprintf(buf, ` %s="%s"`, p.Name, strings.Join(p.Values, ";"))
	}

	return buf.String()
}

// Scheme returns the configured scheme name without the separator and actual input name
// <scheme>://<inputname>
func (s *Stanza) Scheme() string {
	return strings.Split(s.Name, "://")[0]
}

// InputName returns the configured name of the stanza without the scheme and separator
// <scheme>://<inputname>
// If the stanza is called simply "<scheme>", an empty string is returned
func (s *Stanza) InputName() string {
	p := strings.Split(s.Name, "://")
	if len(p) > 1 {
		return p[1]
	} else {
		return ""
	}
}

// Param scans the stanza s parameters and returns the param with the specified name. If not found, returns ""
func (s *Stanza) Param(name string) string {
	for _, p := range s.Params {
		if strings.ToLower(p.Name) == name {
			return p.Value
		}
	}
	return ""
}

// ParamAsCSVList scans the stanza s parameters to find the param with the specified name; it then:
// - splits its value on commas ','
// - trims emtpy spaces from the resulting values
// - returns a slice of said values
// If the parameter was not found, or it was found and is empty: nil is returned
func (s *Stanza) ParamAsCSVList(name string) []string {
	for _, p := range s.Params {
		if strings.ToLower(p.Name) == name {
			if p.Value == "" {
				return nil
			}
			var ret = make([]string, 0, 10)
			for _, v := range strings.Split(p.Value, ",") {
				ret = append(ret, strings.Trim(v, " "))
			}
			return ret
		}
	}
	return nil
}

// ParamAsList scans the stanza s parameters to find the param with the specified name; it then:
// - splits its value occurrences of the 'sep' parameter
// - trims emtpy spaces from the resulting values
// - returns a slice of said values
// If the parameter was not found, or it was found and is empty: nil is returned
func (s *Stanza) ParamAsList(name string, sep string) []string {
	for _, p := range s.Params {
		if strings.ToLower(p.Name) == name {
			if p.Value == "" {
				return nil
			}
			var ret = make([]string, 0, 10)
			for _, v := range strings.Split(p.Value, sep) {
				ret = append(ret, strings.Trim(v, " "))
			}
			return ret
		}
	}
	return nil
}

// ParamList scans the ValidationItem vi "list" parameters and returns the values of the param_list with the specified name.
// If not found, returns an empty list of strings
func (vi *Stanza) ParamList(name string) []string {
	for _, p := range vi.ParamLists {
		if strings.ToLower(p.Name) == name {
			return p.Values
		}
	}
	return []string{}
}

// Host returns the host configured for the stanza s
func (s *Stanza) Host() (ret string) {
	return s.Param("host")
}

// Index returns the index configured for the stanza s
func (s *Stanza) Index() (ret string) {
	return s.Param("index")
}

// Source returns the sourcsourceetype configured for the stanza s
func (s *Stanza) Source() (ret string) {
	return s.Param("source")
}

// Sourcetype returns the sourcetype configured for the stanza s
func (s *Stanza) Sourcetype() (ret string) {
	return s.Param("sourcetype")
}

// Interval returns the execution interval configured for the stanza s
func (s *Stanza) Interval() (ret string) {
	return s.Param("interval")
}

// Param represents a single configuration parameter: a key-value pair as found within inputs.conf
type Param struct {
	/*
	   <param name="param">val</param>
	*/
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"` // name attribute of the param
	Value   string   `xml:",chardata"` // access the textual data of the param value
}

// ParamList stores the list-values of the param_list element within the validation XML
type ParamList struct {
	/*
	   <param_list name="param2">
	       <value>value2</value>
	       <value>value3</value>
	       <value>value4</value>
	   </param_list>
	*/
	XMLName xml.Name `xml:"param_list"`
	Name    string   `xml:"name,attr"` // name attribute of the param
	Values  []string `xml:"value"`     // access the textual data of the param value
}
