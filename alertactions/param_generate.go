package alertactions

import (
	"fmt"
	"strings"
)

/* This file defines the "Param" methods used to generate configuration files, spec files and UI elements.
The following functions are used by the corresponding AlertAction functions in order to piece together all information for the relevant files.
*/

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

// getRestMapConf returns a string which can be used to describe the parameter within splunk's default/restmap.conf file
func (p *Param) getRestMapConf(stanzaName string) string {
	return fmt.Sprintf("#action.%s.param.%s = validate( match('action.%s.param.%s', \"^SOME REGULAR EXPRESSION HERE$\"), \"Setting '%s' is invalid, ADD SOME CUSTOM MESSAGE HERE\")\n", stanzaName, p.Name, stanzaName, p.Name, p.Title)
}

// getUIXML returns a string which can be used to build a HTML UI for the parameter
// https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa#Custom-HTML-component-reference
func (p *Param) getUIHTML(stanzaName string) string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, "<splunk-control-group label=\"%s\" help=\"%s\">\n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	if p.Required {
		fmt.Fprintln(buf, "<span style=\"color:red;margin: 0 2px 0 -5px;\">*</span>")
	}
	switch p.UIType {
	case ParamTypeText:
		fmt.Fprintf(buf, "  <splunk-text-input name=\"action.%s.param.%s\" placeholder=\"%s\" id=\"%s\"></splunk-text-input>\n", stanzaName, p.Name, p.Placeholder, p.Name)
	case ParamTypeTextArea:
		fmt.Fprintf(buf, "  <splunk-text-area name=\"action.%s.param.%s\" placeholder=\"%s\" id=\"%s\"></splunk-text-area>\n", stanzaName, p.Name, p.Placeholder, p.Name)
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

// getAlertActionsConf returns a string which can be used to describe the parameter within splunk's default/alert_actions.conf file
func (p *Param) getCustomConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)
	if p.Required {
		fmt.Fprintf(buf, "# %s: (required) %s \n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	} else {
		fmt.Fprintf(buf, "# %s: %s\n", p.Title, strings.ReplaceAll(p.Description, "\n", " "))
	}
	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "# Available choices: %s\n", strings.Join(p.GetChoices(), "; "))
	}

	fmt.Fprintf(buf, "%s = %s\n", p.Name, strings.ReplaceAll(p.DefaultValue, "\n", "\\\n"))

	return buf.String()
}

// getAlertActionsSpec returns a string which can be used to describe the parameter within splunk's README/alert_actions.conf.spec file
func (p *Param) getCustomSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)
	fmt.Fprintf(buf, `%s = <string>
*  %s: %s
*  Required: %v
*  Default value: "%s"
`, p.Name, p.Title, strings.ReplaceAll(p.Description, "\n", " "), p.Required, strings.ReplaceAll(p.DefaultValue, "\n", " "))
	if len(p.availableOptions) > 0 {
		fmt.Fprintf(buf, "*  Available choices: %s", strings.Join(p.GetChoices(), "; "))
	}
	return buf.String()
}

// getDocumentation returns a markdown-formatted list-item which describes the parameter
func (p *Param) getDocumentation() string {
	buf := new(strings.Builder)
	if p.ConfigFile != "" && p.Stanza != "" {
		// this is a global parameter
		configFile := p.ConfigFile
		if !strings.HasSuffix(configFile, ".conf") {
			configFile = configFile + ".conf"
		}
		fmt.Fprintf(buf, "- %s/[%s]: `%s` : %s - ", configFile, p.Stanza, p.Name, p.Title)
	} else {
		fmt.Fprintf(buf, "- `%s` : %s - ", p.Name, p.Title)
	}
	if p.Required {
		fmt.Fprintf(buf, "(required) ")
	}
	fmt.Fprint(buf, p.Description)

	if p.DefaultValue != "" {
		fmt.Fprintf(buf, "    Default value: `%s`", p.DefaultValue)
	}

	if len(p.availableOptions) > 0 {
		fmt.Fprintln(buf, "    Available choices:")
		for _, option := range p.availableOptions {
			fmt.Fprintf(buf, "    - `%s`: \"%s\"", option.Value, option.VisibleValue)
		}
	}
	return buf.String()
}
