package alertactions

import (
	"fmt"
	"strings"
)

/* This file defines the struct describing a parameter of an alert action */

// ParamType is used to communicate how an alert action parameter should be represented within the UI
// See: https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa#Custom-HTML-elements
type ParamType int

const (
	// https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa#Custom-HTML-elements
	ParamTypeText           ParamType = 1
	ParamTypeTextArea       ParamType = 2
	ParamTypeSearchDropdown ParamType = 3
	ParamTypeDropdown       ParamType = 4
	ParamTypeRadio          ParamType = 5
	ParamTypeColorPicker    ParamType = 6
)

// paramOption contains on admissible internal and visible values for a -dropdown or radio- parameter.
type paramOption struct {
	Value        string
	VisibleValue string
}

// Parameters used by the ModularInput.
type Param struct {
	// Title is the visible name of the parameter, used within the UI
	Title string
	// Name is the internal name of the parameter, the one actually provided within splunk configurations
	Name        string
	Description string
	// UIType encodes how this parameter should be represented in the UI. Use the ParamTypeXXX constants for this.
	UIType ParamType
	// DefaultValue is used in case an actual value has not been set by the run-time configurations
	DefaultValue string
	// Placeholder is a string used within the UI to provide a sample of the value.
	// It only makes sense for parameters of type Text and TextArea
	Placeholder string
	Required    bool

	// availableOptions is a slice of admissible choices for the values of this parameter
	// intended to be used to represent parameters of type dropdown and radio
	availableOptions []paramOption
	// actualValue is the actual value for the parameter provided by run-time configurations
	actualValue string
	// actualValueIsSet tracks whether a value for the parameter has been actually set.
	// if false, the DefaultValue will be returned when asking for the parameter's value
	actualValueIsSet bool
}

// AddChoice adds another valid choice to the set of acceptable ones. This is useful for Dropdown/Radio parameters,
// where the user is expected to pick a value out of a list.
// The "value" is the actual value which would be found within splunk configurations. It empty, the function returns an error.
// The "visibleValue" is what Splunk UI should show the user in the alert configuration panel. If empty, it will get the value of "value".
// The function returns an error if multiple choices having the same "value" have been registered.
func (p *Param) AddChoice(value, visibleValue string) error {
	value = strings.TrimSpace(value)
	visibleValue = strings.TrimSpace(visibleValue)
	if value == "" {
		return fmt.Errorf("param '%v': invald parameter: 'value' cannot be empty string", p.Name)
	}
	if visibleValue == "" {
		visibleValue = value
	}
	if p.availableOptions == nil {
		p.availableOptions = make([]paramOption, 0, 1)
	}
	for _, existingChoice := range p.availableOptions {
		if existingChoice.Value == value {
			return fmt.Errorf("param '%s': duplicated parameter: a choice is already existing having value=\"%s\"", p.Name, value)
		}
	}
	p.availableOptions = append(p.availableOptions, paramOption{Value: value, VisibleValue: visibleValue})
	return nil
}

// setValue sets the run-time value of the parameter. It performs validation of the value based on the parameter's configurations such as AvailableChoices.
// Returns an error in case the validation failed
func (p *Param) SetValue(v string) error {
	v = strings.TrimSpace(v)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	if len(p.availableOptions) > 0 {
		joinedChoices := new(strings.Builder)
		joinedChoices.Grow(100)
		for _, c := range p.availableOptions {
			fmt.Fprintf(joinedChoices, "\"%s\"; ", c.Value)
			if c.Value == v {
				p.actualValue = v
				p.actualValueIsSet = true
				return nil
			}
		}
		return fmt.Errorf("param '%s': provided value '%s' is not included within available choices: %s", p.Name, v, joinedChoices.String())
	}
	p.actualValue = v
	p.actualValueIsSet = true
	return nil
}

// getValue returns the run-time value which was set for this parameter, or its DefaultValue in case no value has been set
func (p *Param) GetValue() string {
	if p.actualValueIsSet {
		return p.actualValue
	}
	return p.DefaultValue
}

// GetChoices returns a list of the internal values of the acceptable options for the parameter.
// If there are no acceptable choices, it returns an empty slice.
func (p *Param) GetChoices() []string {
	l := make([]string, len(p.availableOptions))
	for i, c := range p.availableOptions {
		l[i] = c.Value
	}
	return l
}

// getAlertActionsSpec returns a string which can be used to describe the parameter within splunk's README/alert_actions.conf.spec file
func (p *Param) getAlertActionsSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `param.%s = <string>
*  %s: %s
*  Default value: "%s"
`, p.Name, p.Title, strings.ReplaceAll(p.Description, "\n", " "), strings.ReplaceAll(p.DefaultValue, "\n", " "))

	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "* Available choices: %s", strings.Join(p.GetChoices(), "; "))
	}
	return buf.String()
}

// getAlertActionsConf returns a string which can be used to describe the parameter within splunk's default/alert_actions.conf file
func (p *Param) getAlertActionsConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, "# %s: %s\n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "# Available choices: %s\n", strings.Join(p.GetChoices(), "; "))
	}

	fmt.Fprintf(buf, "param.%s = %s\n", p.Name, strings.ReplaceAll(p.DefaultValue, "\n", "\\\n"))

	return buf.String()
}

// getSavedSearchesSpec returns a string which can be used to describe the parameter within splunk's README/savedsearches.conf.spec file
func (p *Param) getSavedSearchesSpec(stanzaName string) string {
	specVal := "<string>"
	if len(p.availableOptions) > 0 {
		specVal = fmt.Sprintf("[%s]", strings.Join(p.GetChoices(), "|"))
	}
	return fmt.Sprintf("action.%s.param.%s = %s\n", stanzaName, p.Name, specVal)
}
