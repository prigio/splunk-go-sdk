package splunkd

import (
	"fmt"
	"strings"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// ConfigResource represents the contents of a configuration file stanza.
type PropertyResource string

/*
func (pr *PropertyResource) GetString(key string) (val string, err error) {
	tmp, exists := (*pr)[key]
	if !exists {
		return "", fmt.Errorf("not found: '%s'", key)
	}
	switch v := tmp.(type) {
	case string:
		return v, nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case float32:
		return fmt.Sprint(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (cr *PropertyResource) GetInt(key string) (val int, err error) {
	tmp, exists := (*cr)[key]
	if !exists {
		return 0, fmt.Errorf("not found: '%s'", key)
	}
	switch v := tmp.(type) {
	case string:
		intv, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return 0, fmt.Errorf("conversion error of '%s'. Expected an int value, found '%v'. %s", key, tmp, err.Error())
		}
		return int(intv), nil
	case int:
		return v, nil
	case float32:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported type for '%s'. found '%t'", key, v)
	}
}

func (cr *PropertyResource) GetFloat(key string) (val float32, err error) {
	tmp, exists := (*cr)[key]
	if !exists {
		return 0, fmt.Errorf("not found: '%s'", key)
	}
	switch v := tmp.(type) {
	case string:
		floatv, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, fmt.Errorf("conversion error of '%s'. Expected a float value, found '%v''. %s", key, tmp, err.Error())
		}
		return float32(floatv), nil
	case int:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	default:
		return 0, fmt.Errorf("unsupported type for '%s'. found '%t'", key, v)
	}
}
*/

// ConfigsCollection represents a generic configuration file as managed by the /services/configs/conf-<confFileName> endpoint.
// You can manage config file stanzas through this endpoint.
// This also supports custom configuration files defined with a custom SPEC file within etc/apps/<someapp>/README/<somefile>.conf.spec.
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTTUT/RESTconfigurations#Updating_Configuration_Files
// See:https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTREF/RESTconf#configs.2Fconf-.7Bfile.7D
type PropertiesCollection struct {
	collection[PropertyResource]
}

func NewPropertiesCollection(ss *Client, configFileName string, stanzaName string) *PropertiesCollection {
	var col = &PropertiesCollection{}
	configFileName = strings.ToLower(configFileName)
	// remove .conf from filename, if present
	configFileName = strings.TrimSuffix(configFileName, ".conf")

	col.name = "properties/" + configFileName + "/" + stanzaName
	col.path = "properties/" + configFileName + "/" + stanzaName
	col.splunkd = ss
	return col
}

func NewPropertiesCollectionNS(ss *Client, configFileName, stanzaName string, owner, app string) *PropertiesCollection {
	var col = &PropertiesCollection{}
	configFileName = strings.ToLower(configFileName)
	// remove .conf from filename, if present
	configFileName = strings.TrimSuffix(configFileName, ".conf")
	ns, _ := NewNamespace(owner, app, SplunkSharingApp)

	col.name = "properties/" + configFileName + "/" + stanzaName
	col.path = ns.GetServicesNSUrl() + "properties/" + configFileName + "/" + stanzaName
	col.splunkd = ss
	return col
}

/*
	func (col *PropertiesCollection) CreateStanza(name string, params *url.Values) (*entry[PropertyResource], error) {
		if params == nil || len(*params) == 0 {
			return nil, fmt.Errorf("%s createConfig: params cannot be empty", col.name)
		}
		// config
		params.Set("name", name)
		return col.Create(name, params)
	}

	func (col *PropertiesCollection) GetStanza(name string) (*PropertyResource, error) {
		entry, err := col.Get(name)
		if err != nil {
			return nil, err
		}
		return &entry.Content, nil
	}
*/
func (col *PropertiesCollection) GetProperty(propertyName string) (string, error) {
	entries, err := col.List()
	if err != nil {
		return "", fmt.Errorf("getProperty: %w", err)
	}
	for _, e := range entries {
		if e.Name == propertyName {
			return string(e.Content), nil
		}
	}
	return "", fmt.Errorf("getProperty: not found '%s' in '%s'", propertyName, col.path)
}

/*
// GetConfigAsString retrieves the value of configuration configName of the selected stanza
func (col *PropertiesCollection) GetConfigAsString(stanza, configName string) (string, error) {
	stanzaConf, err := col.GetStanza(stanza)
	if err != nil {
		return "", err
	}

	return stanzaConf.GetString(configName)
}

// GetConfigAsInt retrieves the value of configuration configName of the selected stanza
func (col *PropertiesCollection) GetConfigAsInt(stanza, configName string) (int, error) {
	stanzaConf, err := col.GetStanza(stanza)
	if err != nil {
		return 0, err
	}
	return stanzaConf.GetInt(configName)
}

// GetConfigAsFloat retrieves the value of configuration configName of the selected stanza
func (col *PropertiesCollection) GetConfigAsFloat(stanza, configName string) (float32, error) {
	stanzaConf, err := col.GetStanza(stanza)
	if err != nil {
		return 0, err
	}
	return stanzaConf.GetFloat(configName)
}
*/
