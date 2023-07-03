package client

import (
	"fmt"
	"net/url"
	"strings"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

// ConfigResource represents the contents of a configuration file stanza.
type ConfigResource map[string]interface{}

// ConfigsCollection represents a generic configuration file as managed by the /services/configs/conf-<confFileName> endpoint.
// You can manage config file stanzas through this endpoint.
// This also supports custom configuration files defined with a custom SPEC file within etc/apps/<someapp>/README/<somefile>.conf.spec.
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTTUT/RESTconfigurations#Updating_Configuration_Files
// See:https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTREF/RESTconf#configs.2Fconf-.7Bfile.7D
type ConfigsCollection struct {
	collection[ConfigResource]
}

func NewConfigsCollection(ss *SplunkService, configFileName string) *ConfigsCollection {
	var col = &ConfigsCollection{}
	col.name = "conf-" + strings.ToLower(configFileName)
	col.path = "configs/conf-" + strings.ToLower(configFileName)
	col.splunkd = ss
	return col
}

func (col *ConfigsCollection) CreateConfig(name string, params *url.Values) (*collectionEntry[ConfigResource], error) {
	if params == nil || len(*params) == 0 {
		return nil, fmt.Errorf("%s createConfig: params cannot be empty", col.name)
	}
	// config
	params.Set("name", name)
	return col.Create(name, params)
}
