package alertactions

/*

This file contain assistance method to the AlertAction struct used to
generate splunk configuration files needed to configure an alert action
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/prigio/splunk-go-sdk/client"
	"github.com/prigio/splunk-go-sdk/utils"
)

func (aa *AlertAction) generateRuntimeConfig(filename string) string {
	buf := new(strings.Builder)
	var conf []byte
	conf, err := aa.generateAlertConfigJson()
	if err != nil {
		return err.Error()
	}
	fmt.Fprintf(buf, `
You can use the following configuration by:
	1) saving this into a file, e.g. conf.json
	2) executing the alert as:
	
		cat conf.json > %s --execute

%s
`, filename, conf)
	return buf.String()
}

// generateSavedSearchesSpec returns a string which can be used to define the alert action within the splunk configuration file README/savedsearches.conf.spec
func (aa *AlertAction) generateSavedSearchesSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `**
** Configurations custom alert action '%s' within README/savedsearches.conf.spec
** These configurations have been auto-generated
**
action.%s = [0|1]
`, aa.Label, aa.StanzaName)

	for _, par := range aa.params {
		fmt.Fprint(buf, par.getSavedSearchesSpec(aa.StanzaName))
	}
	return buf.String()
}

// generateUIHTML returns a string which can be used to configure the UI for the alert action within the splunk configuration file default/data/ui/alerts/<stanzaname>.html
func (aa *AlertAction) generateUIHTML() string {
	buf := new(strings.Builder)
	buf.Grow(512)
	fmt.Fprintf(buf, `<!-- 
Template for UI configuration. This has been automatically generated
Store this XML content at: default/data/ui/alerts/%s.html -->\n", aa.StanzaName)
Documentation for this file is at: 
	- https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa/
	- https://docs.splunk.com/Documentation/SplunkCloud/9.0.2305/AdvancedDev/CustomVizFormatterApiRef
-->
<form>
`, aa.StanzaName)

	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getUIHTML(aa.StanzaName))
	}

	fmt.Fprintln(buf, "</form>")
	return buf.String()
}

// generateAlertActionsSpec returns a string which can be used to define the alert action within the splunk configuration file README/alert_actions.conf.spec
func (aa *AlertAction) generateAlertActionsSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `** Configurations for custom alert action '%s' within README/alert_actions.conf.spec
** These configurations have been auto-generated
[%s]
`, aa.Label, aa.StanzaName)

	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getAlertActionsSpec())
	}
	return buf.String()
}

// generateAlertActionsConf returns a string which can be used to define the alert action within the splunk configuration file default/alert_actions.conf
func (aa *AlertAction) generateAlertActionsConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `## Configurations for custom alert action '%s' within default/alert_actions.conf
## These configurations have been auto-generated
## See: https://docs.splunk.com/Documentation/Splunk/latest/Admin/Alertactionsconf
## See: https://dev.splunk.com/enterprise/docs/devtools/customalertactions/configappcaa

[%s]
label = %s
description = %s
icon_path = %s
is_custom = 1

# alert.execute.cmd=%s
# NOTE: if you specify this setting, then you also have to specify 
# 			alert.execute.cmd.arg.1 = --execute
# For custom alert actions, explicitly specifies the command to run when the alert action is triggered. This refers to a binary or script
# in the 'bin' folder of the app that the alert action is defined in, or to a path pointer file, also located in the 'bin' folder.
# If a path pointer file (*.path) is specified, the contents of the file is read and the result is used as the command to run.
# Environment variables in the path pointer file are substituted.
# If a python (*.py) script is specified, it is prefixed with the bundled python interpreter.
#alert.execute.cmd.arg.<n> = 
# Change the command line arguments passed to the script when it is invoked. 
# Provide additional arguments to the 'alert.execute.cmd'.
#  Environment variables are substituted.

# This alert action only supports JSON. Do not change this!
payload_format = json

##
## Discretionary settings
##   the developer of the custom aler action might want to set these as needed.
## 

ttl = 14400
# The minimum time to live, in seconds, of the search artifacts, if this action is triggered.
# If 'p' follows '<integer>', then '<integer>' is the number of scheduled periods.
# If no actions are triggered, the ttl for the artifacts are determined
# by the 'dispatch.ttl' setting in the savedsearches.conf file.
# Default: 10p
# Time To Live = 4 Std = 14400 seconds

forceCsvResults = true
# If set to "true", any saved search that includes this action
#  always stores results in CSV format, instead of the internal SRS format.
# If set to "false", results are always serialized using the internal SRS format.
# If set to "auto", results are serialized as CSV if the 'command' setting
# in this stanza starts with "sendalert" or contains the string "$results.file$".
# Default: auto

maxtime = 1h
# The maximum amount of time that the execution of an action is allowed to take before the action is aborted.
# Format: <integer>[m|s|h|d]

#maxresults = <integer>
# Set the global maximum number of search results sent through alerts.
# Default: 100

##
## Parameters specific for this alert
##   these can be autogenerated by starting the alert from the command line.
## The value of these settings is configured at run-time by the alert action configured by the user
`, aa.Label, aa.StanzaName, aa.Label, aa.Description, aa.IconPath, os.Args[0])

	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getAlertActionsConf())
	}

	return buf.String()
}

// generateRestMapConf returns a string which can be used to define the UI validations to be configured within file default/restmap.conf
func (aa *AlertAction) generateRestMapConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `## Configurations for custom alert action '%s' within default/restmap.conf
## These configurations have been auto-generated
## See: https://docs.splunk.com/Documentation/Splunk/8.1.1/Admin/restmapconf
[validation:savedsearch]
`, aa.StanzaName)
	for _, par := range aa.params {
		fmt.Fprintln(buf, par.getRestMapConf(aa.StanzaName))
	}
	return buf.String()
}

// generateAlertConfigJson returns a JSON formatted configuration for the input, based on interactively asked information
func (aa *AlertAction) generateAlertConfigJson() ([]byte, error) {
	ac, err := aa.getAlertConfigInteractive()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(ac, "", "  ")
}

// getAlertConfigInteractive uses the Params[] definition of an alert action to prepare a configuration based on:
// - command line parameters
// - interactively asking the user if no command-line parameter was found for an argument
func (aa *AlertAction) getAlertConfigInteractive() (*alertConfig, error) {
	// first, need to get splunk endpoint, username and password to be able to login into the service if necessary.
	ic := &alertConfig{}
	fmt.Println("Interactively provide information to access local splunkd service.")
	ic.ServerUri = utils.AskForInput("Splunkd URL", "https://localhost:8089", false)
	username := utils.AskForInput("Splunk username", "admin", false)
	password := utils.AskForInput("Splunk password", "", true)
	ss, err := client.New(ic.ServerUri, true, "")
	if err != nil {
		return nil, fmt.Errorf("connection failed to splunkd on '%s'. %w", ic.ServerUri, err)
	}
	if err = ss.Login(username, password, ""); err != nil {
		return nil, fmt.Errorf("login failed to splunkd on with username '%s': %w", username, err)
	}

	ic.SessionKey = ss.GetSessionKey()
	ic.App = utils.AskForInput("Splunk app context", "", false)
	ic.Owner = username
	//ic.ResultsFile =
	//ic.ResultsLink =
	//ic.ServerHost =
	ic.SearchUri = "interactive search"
	ic.Sid = "sid of interactive search"
	ic.SearchName = "interactive search"
	fmt.Println("Interactively provide values for alert action parameters.")
	ic.Configuration = make(map[string]string)
	for _, p := range aa.params {
		ic.Configuration[p.Name] = utils.AskForInput(p.Title, p.DefaultValue, false)
	}

	/*
		for seq, arg := range aa.Args {
			prompt = fmt.Sprintf("Provide parameter %s (%s, '%s')", arg.Title, arg.DataType, arg.Name)
			if arg.Description != "" {
				prompt = fmt.Sprintf("%s\n    %s\n", prompt, arg.Description)
			}
			val = AskForInput(prompt, arg.DefaultValue, false)
			stanza.Params[seq] = Param{Name: arg.Name, Value: val}
		}
	*/

	return ic, nil
}
