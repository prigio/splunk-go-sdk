package params

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/prigio/splunk-go-sdk/v2/errors"
	"github.com/prigio/splunk-go-sdk/v2/splunkd"
)

// paramOption contains admissible internal and visible values for a -dropdown or radio- parameter.
type paramOption struct {
	Value        string
	VisibleValue string
}

// Param represents a generic parameter to an alert action or a modular input, or the value of a configuration setting present in a splunk configuration file.
// It can be used in multiple ways:
//
//  1. as a regular run-time parameter, whose value is provided by Splunk via STDIN to the alert action or modular input during its startup
//  2. as a global parameter, whose value is statically configured in some configuration file (normally called after the app hosting the alert action)
//
// Initialize this struct using the NewParam functions.
type Param struct {
	// name is the internal name of the parameter, the one actually provided within splunk configurations
	name string
	// configFile is the name of the configuration file where this parameter is located.
	// This applies only to global parameters, which are not defined within alert_actions.conf
	configFile string
	// stanza is the name of the configuration stanza within ConfigFile file where this parameter is located.
	// This applies only to global parameters, which are not defined within alert_actions.conf
	stanza string
	// title is the visible name of the parameter, used within the UI
	title       string
	description string
	// defaultValue is used in case an actual value has not been set by the run-time configurations
	defaultValue string
	// availableOptions is a slice of admissible choices for the values of this parameter
	// intended to be used to represent parameters of type dropdown and radio
	availableOptions []paramOption
	// whether a value is necessary for this parameter
	required bool
	// sensitive expresses whether the parameter can or cannot be logged. If sensitive, then the actual value should be masked upon logging
	sensitive bool

	dataType string
	// customProps can be used by the user to store additional metadata for the parameter
	// like the type of UI element it should be displayed with, splunk-defined validation rules etc.
	customProps map[string]string

	//dataType    ParamDataType
	// uiType encodes how this parameter should be represented in the UI. Use the ParamTypeXXX constants for this.
	// uiType ParamType
	// placeholder is a string used within the UI to provide a sample of the value.
	// It only makes sense for parameters of type Text and TextArea
	placeholder string

	// actualValue is the actual value for the parameter provided by run-time configurations
	actualValue string
	// actualValueIsSet tracks whether a value for the parameter has been actually set.
	// if false, the DefaultValue will be returned when asking for the parameter's value
	actualValueIsSet bool

	mu sync.RWMutex
}

// NewParam instantiates a parameter, whose value is provided by splunk to the alert action when starting it up
// func NewParam(name, title, description, defaultValue, placeholder string, uiType ParamType, required bool) (*Param, error) {
// there is no stanza setting for a non-global parameter, as they might be defined in one or more stanzas.
//
//	return newParameter("alert_actions.conf", "", name, title, description, defaultValue, placeholder, uiType, required)
func NewParam(configFile, stanza, name, title, description, defaultValue string, required, sensitive bool) (*Param, error) {
	if name == "" {
		return nil, errors.NewErrInvalidParam("newParameter", nil, "'name' cannot be empty")
	}
	if title == "" {
		return nil, errors.NewErrInvalidParam("newParameter", nil, "'title' cannot be empty for '%s'", name)
	}
	if configFile == "" {
		return nil, errors.NewErrInvalidParam("newParameter", nil, "'configFile' cannot be empty for '%s'", name)
	}

	configFile = strings.TrimSuffix(configFile, ".conf")

	param := &Param{
		name:         name,
		title:        title,
		description:  description,
		configFile:   configFile,
		stanza:       stanza,
		defaultValue: defaultValue,
		required:     required,
		sensitive:    sensitive,
	}
	return param, nil
}

// AddChoice adds another valid choice to the set of acceptable ones.
// This is useful for AlertAction Dropdown/Radio parameters, where the user is expected to pick a value out of a list.
// The "value" is the actual value which would be found within splunk configurations. It empty, the function returns an error.
// The "visibleValue" is what Splunk UI should show the user in the alert configuration panel. If empty, it will get the value of "value".
// In case a choice with the same "value" already exists, the visibleValue of the existing choice will be overwritten with the one provided.
func (p *Param) AddChoice(value, visibleValue string) error {
	value = strings.TrimSpace(value)
	visibleValue = strings.TrimSpace(visibleValue)
	if value == "" {
		return errors.NewErrInvalidParam("addChoice", nil, "'value' cannot be empty for parameter '%s'", p.name)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if visibleValue == "" {
		visibleValue = value
	}
	if p.availableOptions == nil {
		p.availableOptions = make([]paramOption, 0, 1)
	}
	for i, existingChoice := range p.availableOptions {
		if existingChoice.Value == value {
			// if a choice with the same value is already existing, silently overwrite its visible value.
			p.availableOptions[i] = paramOption{Value: value, VisibleValue: visibleValue}
			return nil
		}
	}
	p.availableOptions = append(p.availableOptions, paramOption{Value: value, VisibleValue: visibleValue})
	return nil
}

// IsSensitive informs whether the parameter contains sensitive data.
// If true, the parameter value MUST be masked when being logged or printed-out.
func (p *Param) IsSensitive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sensitive
}

// IsRequired informs whether a value for the parameter is required for operation.
func (p *Param) IsRequired() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.required
}

// GetDefaultValue returns the value of the parameter to be used in case no specific configuration is available
func (p *Param) GetDefaultValue() string {
	// no locking, as this is not mutable after parameter creation
	return p.defaultValue
}

// GetConfigFile returns the name of the configuration file holding this parameter.
func (p *Param) GetConfigFile() string {
	// no locking, as this is not mutable after parameter creation
	return p.configFile
}

// GetConfigFile returns the name of the stanza within the configuration file holding this parameter.
// If empty, multiple stanzas within the configuration file can hold this parameter
func (p *Param) GetStanza() string {
	// no locking, as this is not mutable after parameter creation
	return p.stanza
}

// GetName returns the internal name of the parameter, the one used within configuration files and API calls.
func (p *Param) GetName() string {
	return p.name
}

// GetTitle returns the visible name of the parameter, the one used within the UI.
func (p *Param) GetTitle() string {
	return p.title
}

// GetDescription returns the textual description of the parameter.
func (p *Param) GetDescription() string {
	return p.description
}

// GetDataType returns the type of data for the parameter, one of:
// - "string"
// - "boolean"
// - "number"
// In case no dataType has been explicitly set, the default value is "string"
func (p *Param) GetDataType() string {
	if p.dataType == "" {
		return "string"
	}
	return p.dataType
}

// GetConfigDefinition returns a triple (configFile, stanza, param name) defining where this parameter has been defined.
func (p *Param) GetConfigDefinition() (configFile, stanza, paramName string) {
	return p.configFile, p.stanza, p.name
}

// String returns a textual representation of where the parameter is defined.
// Useful for logging purposes.
func (p *Param) String() string {
	if p.configFile != "" && p.stanza != "" {
		return fmt.Sprintf("%s[%s]/%s", p.configFile, p.stanza, p.name)
	}
	if p.configFile != "" && p.stanza == "" {
		return fmt.Sprintf("%s[*]/%s", p.configFile, p.name)
	}
	// if no information about configFile and Stanza are available, just return the name
	return p.name
}

// GetChoices returns a list of the internal values of the acceptable options for the parameter.
// If there are no acceptable choices, it returns an empty slice.
func (p *Param) GetChoices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	l := make([]string, len(p.availableOptions))
	for i, c := range p.availableOptions {
		l[i] = c.Value
	}
	return l
}

// GetCustomProperty retrieves additional properties of the parameter, such as its UI type, ad-hoc validation rules etc.
// This is the dual function to SetCustomProperty.
// This returns the value of the property, if set. It returns an empty string in case the property was not found.
func (p *Param) GetCustomProperty(name string) string {
	if name == "" {
		return ""
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.customProps == nil {
		return ""
	}
	return p.customProps[name]
}

// ForceValue forces the configuration of a value for the parameter. This can be used in cases where the parameter value comes from external sources like:
// - manual, interactive setting
// - the XML provided to a modular input at startup
// - the XML or JSON provided to an alert action at startup
// This performs validation of the value based on the parameter's configurations such as AvailableChoices and returns an error in case the validation failed.
func (p *Param) ForceValue(v string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	v = strings.TrimSpace(v)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	if len(p.availableOptions) > 0 {
		joinedChoices := new(strings.Builder)
		joinedChoices.Grow(100)
		for _, c := range p.availableOptions {
			fmt.Fprintf(joinedChoices, `"%s"; `, c.Value)
			if c.Value == v {
				p.actualValue = v
				p.actualValueIsSet = true
				return nil
			}
		}
		return fmt.Errorf("param '%s': provided value '%s' is not included within available choices: %s", p.name, v, joinedChoices.String())
	}
	p.actualValue = v
	p.actualValueIsSet = true
	return nil
}

// SetSensitive configures the parameter to contain sensitive data.
// The parameter value will be masked when being logged or printed-out.
func (p *Param) SetSensitive() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sensitive = true
}

// SetRequired configures the parameter to be necessary for execution.
func (p *Param) SetRequired() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.required = true
}

// SetDataType configures the expected type of value for the parameter.
// This is used when generating specifications for the configuration files (README/*.conf.spec files)
// The following are the only adminissible data types. Anything else will generate an error.
// - "string"
// - "boolean"
// - "number"
// If no data type has been configured, the default is "string"
func (p *Param) SetDataType(dt string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !(dt == "string" || dt == "boolean" || dt == "number") {
		return errors.NewErrInvalidParam("setDataType["+p.name+"]", nil, `'dataType': must be one of string/boolean/number. provided="%s"`, dt)
	}
	p.dataType = dt
	return nil
}

// SetCustomProperty can be used to store additional properties of the parameter, such as its UI type, ad-hoc validation rules etc.
func (p *Param) SetCustomProperty(name, value string) {
	if name == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.customProps == nil {
		p.customProps = make(map[string]string)
	}
	p.customProps[name] = value

}

// HasForcedValue informs whether a forced value has been set for the parameter.
func (p *Param) HasForcedValue() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.actualValueIsSet
}

// GetValue returns the run-time value for this parameter, according to the following rules:
//   - if 'ForceValue' was used, such value is returned, otherwise
//   - splunkd is queried for the 'property' related to the parameter's config file, stanza and name, using the system context.
//     If an error occurs, the default value is returend, ALONG with the error.
//
// The returned value has environment variables substituted if the value contains something like '$var' or '${var}'
func (p *Param) GetValue(client *splunkd.Client) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.actualValueIsSet {
		return os.ExpandEnv(p.actualValue), nil
	}

	if client == nil {
		return os.ExpandEnv(p.defaultValue), errors.NewErrInvalidParam("getValue["+p.name+"]", nil, "'client' should not be nil as the value for the parameter must be read from the splunkd API'. Returning the default value")
	}
	val, err := splunkd.NewPropertiesCollection(client, p.configFile).GetProperty(p.stanza, p.name)
	if err != nil {
		return os.ExpandEnv(p.defaultValue), fmt.Errorf(`getValue[%s]: could not retrieve parameter value from splunkd %s. Returning the default value. %w`, p.name, p.String(), err)
	}
	return os.ExpandEnv(val), nil
}

// GetValueNS returns the run-time value for this parameter, according to the following rules:
//   - if 'ForceValue' was used, such value is returned, otherwise
//   - splunkd is queried for the 'property' related to the parameter's config file, stanza and name, using the provided owner and app context.
//     If an error occurs, the default value is returend, ALONG with the error.
//
// The returned value has environment variables substituted if the value contains something like '$var' or '${var}'
func (p *Param) GetValueNS(client *splunkd.Client, owner, app string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.actualValueIsSet {
		return os.ExpandEnv(p.actualValue), nil
	}

	if client == nil {
		return os.ExpandEnv(p.defaultValue), errors.NewErrInvalidParam("getValueNS["+p.name+"]", nil, "'client' should not be nil as the value for the parameter must be read from the splunkd API'. Returning the default value")
	}
	val, err := splunkd.NewPropertiesCollectionNS(client, p.configFile, owner, app).GetProperty(p.stanza, p.name)
	if err != nil {
		return os.ExpandEnv(p.defaultValue), fmt.Errorf(`getValueNS[%s]: could not retrieve parameter value from splunkd %s. Returning the default value. %w`, p.name, p.String(), err)
	}
	return os.ExpandEnv(val), nil
}
