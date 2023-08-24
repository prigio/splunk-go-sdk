package alertactions

import (
	"fmt"
	"strings"
)

/* This file defines the "Param" methods used to generate configuration files, spec files and UI elements.
The following functions are used by the corresponding AlertAction functions in order to piece together all information for the relevant files.
*/

// GenerateSpec returns a string which can be used to describe the parameter within the specification of a splunk configuration file
// Parameter 'namePrefix' is prefixed to the name of the parameter.
// This is needed as some configuration files track custom parameters as "param.XXX" or in other ways.
// E.g.
//
//	param name is "debug"
//	namePrefix is "config."
//	the resulting line within the .conf file will be:
//	    config.debug = ...
func (p *Param) GenerateSpec(namePrefix string) string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	if namePrefix != "" {
		namePrefix = strings.TrimRight(namePrefix, ".")
		fmt.Fprintf(buf, "%s.", namePrefix)
	}
	fmt.Fprintf(buf, `%s = <string>
*  %s: %s
*  Required: %v
*  Default value: "%s"
`, p.Name, p.Title, strings.ReplaceAll(p.Description, "\n", " "), p.required, strings.ReplaceAll(p.defaultValue, "\n", " "))

	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "* Available choices: %s", strings.Join(p.GetChoices(), "; "))
	}
	return buf.String()
}

// GenerateConf returns a string which can be used to describe the parameter within a splunk configuration file
// Parameter 'namePrefix' is prefixed to the name of the parameter.
// This is needed as some configuration files track custom parameters as "param.XXX" or in other ways.
// E.g.
//
//	param name is "debug"
//	namePrefix is "config."
//	the resulting line within the .conf file will be:
//	    config.debug = ...
func (p *Param) GenerateConf(namePrefix string) string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, "# %s: %s\n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "# Available choices: %s\n", strings.Join(p.GetChoices(), "; "))
	}

	if namePrefix == "" {
		fmt.Fprintf(buf, "%s = %s\n", p.Name, strings.ReplaceAll(p.defaultValue, "\n", "\\\n"))
	} else {
		fmt.Fprintf(buf, "%s%s = %s\n", namePrefix, p.Name, strings.ReplaceAll(p.defaultValue, "\n", "\\\n"))
	}

	return buf.String()
}

// GenerateRestMapConf returns a string which can be used to describe the parameter within splunk's default/restmap.conf file
func (p *Param) GenerateRestMapConf(stanzaName string) string {
	// this only is only needed for NON global parameters
	// global parameters get an empty string
	if p.configFile != "" && p.stanza != "" {
		return ""
	}
	return fmt.Sprintf("#action.%s.param.%s = validate( match('action.%s.param.%s', \"^SOME REGULAR EXPRESSION HERE$\"), \"Setting '%s' is invalid, ADD SOME CUSTOM MESSAGE HERE\")\n", stanzaName, p.Name, stanzaName, p.Name, p.Title)
}

// getUIXML returns a string which can be used to build a HTML UI for the parameter
// https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa#Custom-HTML-component-reference
func (p *Param) getUIHTML(stanzaName string) string {
	// this only is only needed for NON global parameters
	// global parameters get an empty string
	if p.configFile != "" && p.stanza != "" {
		return ""
	}
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, "<splunk-control-group label=\"%s\" help=\"%s\">\n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	if p.required {
		fmt.Fprintln(buf, "<span style=\"color:red;margin: 0 2px 0 -5px;\">*</span>")
	}
	switch p.uiType {
	case ParamTypeText:
		fmt.Fprintf(buf, "  <splunk-text-input name=\"action.%s.param.%s\" placeholder=\"%s\" id=\"%s\"></splunk-text-input>\n", stanzaName, p.Name, p.placeholder, p.Name)
	case ParamTypeTextArea:
		fmt.Fprintf(buf, "  <splunk-text-area name=\"action.%s.param.%s\" placeholder=\"%s\" id=\"%s\"></splunk-text-area>\n", stanzaName, p.Name, p.placeholder, p.Name)
	case ParamTypeDropdown:
		fmt.Fprintf(buf, "  <splunk-select name=\"action.%s.param.%s\" id=\"%s\">\n", stanzaName, p.Name, p.Name)
		for _, c := range p.availableOptions {
			fmt.Fprintf(buf, "    <option value=\"%s\">%s</option>\n", c.Value, c.VisibleValue)
		}
		fmt.Fprintf(buf, "  </splunk-select>\n")
	case ParamTypeRadio:
		fmt.Fprintf(buf, "  <splunk-radio-input name=\"action.%s.param.%s\" id=\"%s\">\n", stanzaName, p.Name, p.Name)
		for _, c := range p.availableOptions {
			fmt.Fprintf(buf, "    <option value=\"%s\">%s</option>\n", c.Value, c.VisibleValue)
		}
		fmt.Fprintf(buf, "\n  </splunk-radio-input>\n")
	case ParamTypeColorPicker:
		fmt.Fprintf(buf, "  <splunk-color-picker name=\"action.%s.param.%s\" id=\"%s\" palette=\"splunkSemantic\"></splunk-color-picker>\n", stanzaName, p.Name, p.Name)
	}
	fmt.Fprint(buf, "</splunk-control-group>\n")

	return buf.String()
}

// GenerateDocumentation returns a markdown-formatted list-item which describes the parameter
func (p *Param) GenerateDocumentation() string {
	buf := new(strings.Builder)
	if p.configFile != "" && p.stanza != "" {
		// this is a global parameter
		configFile := p.configFile
		if !strings.HasSuffix(configFile, ".conf") {
			configFile = configFile + ".conf"
		}
		fmt.Fprintf(buf, "- %s/[%s]: `%s` : %s - ", configFile, p.stanza, p.Name, p.Title)
	} else {
		fmt.Fprintf(buf, "- `%s` : %s - ", p.Name, p.Title)
	}
	if p.required {
		fmt.Fprintf(buf, "(required) ")
	}
	fmt.Fprint(buf, p.Description)

	if p.defaultValue != "" {
		fmt.Fprintf(buf, "    Default value: `%s`", p.defaultValue)
	}

	if len(p.availableOptions) > 0 {
		fmt.Fprintln(buf, "    Available choices:")
		for _, option := range p.availableOptions {
			fmt.Fprintf(buf, "    - `%s`: \"%s\"", option.Value, option.VisibleValue)
		}
	}
	return buf.String()
}
