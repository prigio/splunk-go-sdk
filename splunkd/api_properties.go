package splunkd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/prigio/splunk-go-sdk/utils"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// ConfigResource represents the contents of a configuration file stanza.
type PropertyResource string

// ConfigsCollection represents a generic configuration file as managed by the /services/configs/conf-<confFileName> endpoint.
// You can manage config file stanzas through this endpoint.
// This also supports custom configuration files defined with a custom SPEC file within etc/apps/<someapp>/README/<somefile>.conf.spec.
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTTUT/RESTconfigurations#Updating_Configuration_Files
// See:https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTREF/RESTconf#configs.2Fconf-.7Bfile.7D
type PropertiesCollection struct {
	collection[PropertyResource]
}

func NewPropertiesCollection(ss *Client, configFileName string) *PropertiesCollection {
	var col = &PropertiesCollection{}
	configFileName = strings.ToLower(configFileName)
	// remove .conf from filename, if present
	configFileName = strings.TrimSuffix(configFileName, ".conf")

	col.name = "properties/" + configFileName
	col.path = "properties/" + configFileName
	col.splunkd = ss
	return col
}

func NewPropertiesCollectionNS(ss *Client, configFileName, owner, app string) *PropertiesCollection {
	var col = &PropertiesCollection{}
	configFileName = strings.ToLower(configFileName)
	// remove .conf from filename, if present
	configFileName = strings.TrimSuffix(configFileName, ".conf")
	ns, _ := NewNamespace(owner, app, SplunkSharingApp)

	col.name = "properties/" + configFileName
	col.path = ns.GetServicesNSUrl() + "properties/" + configFileName
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

func (col *PropertiesCollection) GetStanza(name string) (map[string]string, error) {
	if name == "" {
		return nil, utils.NewErrInvalidParam(col.name+" getStanza", nil, "name cannot be empty")
	}
	tmpCol := collection[PropertyResource]{splunkd: col.splunkd, name: col.name + "/" + name, path: col.path + "/" + name}
	entries, err := tmpCol.List()
	if err != nil {
		return nil, fmt.Errorf("%s getStanza: %w", col.name, err)
	}
	properties := make(map[string]string)
	for _, e := range entries {
		properties[e.Name] = string(e.Content)
	}
	return properties, nil
}

func (col *PropertiesCollection) Create(name string, properties *url.Values) error {
	return col.CreateStanza(name, properties)
}

func (col *PropertiesCollection) CreateStanza(name string, properties *url.Values) error {
	// https://docs.splunk.com/Documentation/Splunk/9.1.1/RESTREF/RESTconf#properties
	if name == "" {
		return utils.NewErrInvalidParam(col.name+" createStanza", nil, "name cannot be empty")
	}
	//var discard *discardBody
	var params url.Values = url.Values{}
	params.Set("__stanza", name)
	if err := doSplunkdHttpRequest(col.splunkd, "POST", getUrl(col.path, ""), nil, []byte(params.Encode()), "application/x-www-form-urlencoded", &discardBody{}); err != nil {
		return fmt.Errorf("%s createStanza %s: %w", col.name, name, err)
	}
	return col.SetProperties(name, properties)
}

func (col *PropertiesCollection) DeleteStanza(stanza string) error {
	// https://docs.splunk.com/Documentation/Splunk/9.1.1/RESTREF/RESTconf#properties
	if stanza == "" {
		return utils.NewErrInvalidParam(col.name+" deleteProperty", nil, "stanza cannot be empty")
	}

	if err := doSplunkdHttpRequest(col.splunkd, "DELETE", getUrl(col.path, stanza), nil, nil, "", &discardBody{}); err != nil {
		return fmt.Errorf("%s deleteStanza %s: %w", col.name, stanza, err)
	}
	return nil
}

func (col *PropertiesCollection) SetProperties(stanza string, properties *url.Values) error {
	// https://docs.splunk.com/Documentation/Splunk/9.1.1/RESTREF/RESTconf#properties
	if stanza == "" {
		return utils.NewErrInvalidParam(col.name+" setProperties", nil, "stanza cannot be empty")
	}

	if properties == nil || len(*properties) == 0 {
		return nil
	}
	if err := doSplunkdHttpRequest(col.splunkd, "POST", getUrl(col.path, stanza), nil, []byte(properties.Encode()), "application/x-www-form-urlencoded", &discardBody{}); err != nil {
		return fmt.Errorf("%s setProperties %s: %w", col.name, stanza, err)
	}
	return nil
}

func (col *PropertiesCollection) SetProperty(stanza, propertyName, value string) error {
	// https://docs.splunk.com/Documentation/Splunk/9.1.1/RESTREF/RESTconf#properties
	if stanza == "" {
		return utils.NewErrInvalidParam(col.name+" setProperty", nil, "stanza cannot be empty")
	}
	if propertyName == "" {
		return utils.NewErrInvalidParam(col.name+" setProperty", nil, "propertyName cannot be empty")
	}

	var params url.Values = url.Values{}
	params.Set(propertyName, value)
	if err := doSplunkdHttpRequest(col.splunkd, "POST", getUrl(col.path, stanza), nil, []byte(params.Encode()), "application/x-www-form-urlencoded", &discardBody{}); err != nil {
		return fmt.Errorf("%s setProperty %s/%s: %w", col.name, stanza, propertyName, err)
	}
	return nil
}

func (col *PropertiesCollection) GetProperty(stanza, propertyName string) (string, error) {
	if stanza == "" {
		return "", utils.NewErrInvalidParam(col.name+" getProperty", nil, "stanza cannot be empty")
	}
	if propertyName == "" {
		return "", utils.NewErrInvalidParam(col.name+" getProperty", nil, "propertyName cannot be empty")
	}
	props, err := col.GetStanza(stanza)
	if err != nil {
		return "", fmt.Errorf("%s getProperty: %w", col.name, err)
	}
	return props[propertyName], nil
}
