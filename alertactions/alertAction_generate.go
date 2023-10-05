package alertactions

/*

This file contain assistance methods to the AlertAction struct used to
generate splunk configuration files based on the registered parameter and global parameters.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/prigio/splunk-go-sdk/v2/splunkd"
	"github.com/prigio/splunk-go-sdk/v2/utils"
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
		fmt.Fprint(buf, par.GenerateSpec(fmt.Sprintf("action.%s.param", aa.StanzaName)))
	}
	return buf.String()
}

// generateUIXML returns a string which can be used to configure the UI for the alert action within the splunk configuration file default/data/ui/alerts/<stanzaname>.html
func (aa *AlertAction) generateUIXML() string {
	var (
		err  error
		buf  *strings.Builder
		html string
	)
	buf = new(strings.Builder)
	buf.Grow(512)
	fmt.Fprintf(buf, `<!-- 
Template for UI configuration. This has been automatically generated
Store this XML content at: default/data/ui/alerts/%s.html

Documentation for this file is at: 
	- https://dev.splunk.com/enterprise/docs/devtools/customalertactions/createuicaa/
	- https://docs.splunk.com/Documentation/SplunkCloud/9.0.2305/AdvancedDev/CustomVizFormatterApiRef
-->
<form>
<splunk-control-group label="Instructions">
	<span class="help-block">
		Values can contain tokens such as <code>$name$</code> and <code>$result.fieldname$</code>.<br/>
		Read more <a target="_blank" href="https://docs.splunk.com/Documentation/Splunk/9.1.0/Alert/EmailNotificationTokens">here</a>.
	</span>
</splunk-control-group>
`, aa.StanzaName)

	for _, par := range aa.params {
		// retrieve the UI type of the parameter from the custom properties of the parameter itself
		html, err = par.GenerateUIXML(aa.StanzaName, par.GetCustomProperty("uiType"))
		if err != nil {
			aa.Log("ERROR", "Cannot generate UI HTML for parameter '%s'. %s", par.GetName(), err)
		} else {
			fmt.Fprintln(buf, html)
		}
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
		fmt.Fprintln(buf, par.GenerateSpec("param."))
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

maxtime = 5m
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
		fmt.Fprintln(buf, par.GenerateConf("param."))
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
		fmt.Fprintln(buf, par.GenerateRestMapConf(aa.StanzaName))
	}
	return buf.String()
}

// generateDocumentation returns a markdown-formatted string describing the alert and its parameters
func (aa *AlertAction) generateDocumentation() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `# Alert action "%s"

%s

`, aa.Label, aa.Description)
	if aa.Documentation != "" {
		fmt.Fprintln(buf, aa.Documentation)
	}

	fmt.Fprint(buf, `
## User-facing parameters

The following describes the parameters whicn an end user can setup using the alert action user interface.

`)
	for _, par := range aa.params {
		fmt.Fprintln(buf, par.GenerateDocumentation())
	}

	fmt.Fprintf(buf, `

## Global parameters

Global parameters are set by administrators and are valid for all executions of the alert. 
They are set in a custom configuration file and stanza, as described in the following.

`)
	for _, par := range aa.globalParams {
		fmt.Fprintln(buf, par.GenerateDocumentation())
	}

	fmt.Fprintf(buf, `

## Troubleshooting

The following indicates ways to troubleshoot functionality of the alert.

Each execution of the alert generates a random "runId" which is then written alongside all the emitted logs.

### Logs
The alert performs logging on STDERR until a runtime configuration is loaded. 
Afterwards, it uses a dedicated sourcetype for its own logging.
This means, that before the load of runtime configuration, logs are sent to:

    index=_internal sourcetype=splunkd component=sendmodalert action="%s"

You can use the Splunk UI to enhance the verbosity of the logging of the "sendmodalert" component 
to more details of Splunk's own execution of the alert. 

After load of runtime configuration, the alert then writes its own data within:

    index=_internal sourcetype="%s"

You can therefore use the following splunk search to look for all these logs:

    index=_internal
       (sourcetype=splunkd component=sendmodalert action="%s")
       OR sourcetype="%s"

### Interactive execution

You can run the alert from the command-line by invoking it with the appropriate command-line switch. Execute './<filename> -h' for more information

### Common issues

In case you are seing ONLY logs like this in "index=_internal sourcetype=splunkd component=sendmodalert"

INFO  sendmodalert [21376 AlertNotifierWorker-0] - Invoking modular alert action=alert-jira-transition for search="something" sid="scheduler__user__search__RMD529dab0d5260fcdd7_at_1689737700_50" in app="search" owner="user" type="saved"
ERROR sendmodalert [21376 AlertNotifierWorker-0] - action=alert-jira-transition - Execution of alert action script failed

Chances are splunk was not able to start the alert script at all: is the executable really executable? 
Check within "$SPLUNK_HOME/etc/apps/<appname>/[linux/windows/darwin]_.../bin/" that the alert files are executable for the splunk OS user.

`, aa.StanzaName, aa.getLoggingSourcetype(), aa.StanzaName, aa.getLoggingSourcetype())

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
	fmt.Println("> Interactively provide information to access local splunkd service.")
	ss, err := splunkd.NewInteractive()
	if err != nil {
		return nil, fmt.Errorf("getAlertConfigInteractive: %w", err)
	}
	username, _ := ss.Username()
	ic.App = utils.AskForInput("Splunk app context", "", false)
	ic.Owner = utils.AskForInput("Owner (can be different to your username only if you have admin rights)", username, false)
	ic.ServerUri = ss.GetSplunkdURI()
	ic.SessionKey = ss.GetSessionKey()
	if splunkdCtx, err := ss.AuthContext(); err != nil {
		ic.Owner = splunkdCtx.Username
	}
	if splunkInfo, err := ss.Info(); err != nil {
		ic.ServerHost = splunkInfo.ServerName
	}
	//ic.ResultsFile =
	//ic.ResultsLink =
	ic.SearchUri = "interactive search"
	ic.Sid = "sid of interactive search"
	ic.SearchName = "interactive search"

	// in case the alert uses global parameters, ask for them
	if len(aa.globalParams) > 0 {
		resp := utils.AskForInput("Do you want to specify global parameters manually (y), or get their value from splunk (n)", "n", false)
		if strings.ToLower(resp) == "y" {
			for _, p := range aa.globalParams {
				pVal := utils.AskForInput(p.GetTitle(), p.GetDefaultValue(), p.IsSensitive())
				p.ForceValue(pVal)
			}
		}
	}

	fmt.Println("> Interactively provide values for alert action parameters.")
	ic.Configuration = make(map[string]string)
	for _, p := range aa.params {
		ic.Configuration[p.GetName()] = utils.AskForInput(p.GetTitle(), p.GetDefaultValue(), p.IsSensitive())
	}

	return ic, nil
}

func (aa *AlertAction) generateAdHocConfigSpecs() string {
	var paramsByFileAndStanza map[string]map[string][]string = make(map[string]map[string][]string)

	for _, p := range aa.globalParams {
		confFile, stanza, _ := p.GetConfigDefinition()
		if !strings.HasSuffix(confFile, ".conf") {
			confFile = confFile + ".conf"
		}
		if _, found := paramsByFileAndStanza[confFile]; !found {
			paramsByFileAndStanza[confFile] = make(map[string][]string)
			paramsByFileAndStanza[confFile][stanza] = make([]string, 0)
		}
		if _, found := paramsByFileAndStanza[confFile][stanza]; !found {
			paramsByFileAndStanza[confFile][stanza] = make([]string, 0)
		}
		paramsByFileAndStanza[confFile][stanza] = append(paramsByFileAndStanza[confFile][stanza], p.GenerateSpec(""))
	}

	buf := new(strings.Builder)

	fmt.Fprintf(buf, `
** Specification for custom configuration files for alert action '%s' [%s]
** These specs have been auto-generated
`, aa.Label, aa.StanzaName)

	for file, stanzas := range paramsByFileAndStanza {
		fmt.Fprintf(buf, `
**
** File: 'README/%s.spec'
**
`, file)
		for stanza, paramSpecs := range stanzas {
			fmt.Fprintf(buf, "[%s]\n", stanza)
			for _, specString := range paramSpecs {
				fmt.Fprintln(buf, specString)
			}

		}
		fmt.Fprintln(buf, "")
	}
	return buf.String()
}

func (aa *AlertAction) generateAdHocConfigConfs() string {
	var paramsByFileAndStanza map[string]map[string][]string = make(map[string]map[string][]string)

	for _, p := range aa.globalParams {
		confFile := p.GetConfigFile()
		if !strings.HasSuffix(confFile, ".conf") {
			confFile = confFile + ".conf"
		}
		if _, found := paramsByFileAndStanza[confFile]; !found {
			paramsByFileAndStanza[confFile] = make(map[string][]string)
			paramsByFileAndStanza[confFile][p.GetStanza()] = make([]string, 0)
		}
		if _, found := paramsByFileAndStanza[confFile][p.GetStanza()]; !found {
			paramsByFileAndStanza[confFile][p.GetStanza()] = make([]string, 0)
		}
		paramsByFileAndStanza[confFile][p.GetStanza()] = append(paramsByFileAndStanza[confFile][p.GetStanza()], p.GenerateConf(""))
	}

	buf := new(strings.Builder)

	fmt.Fprintf(buf, `
## Configurations for custom configuration files for alert action '%s' [%s]
## These configurations have been auto-generated
`, aa.Label, aa.StanzaName)

	for file, stanzas := range paramsByFileAndStanza {
		fmt.Fprintf(buf, `
##
## File: 'default/%s'
##
`, file)
		for stanza, paramSpecs := range stanzas {
			fmt.Fprintf(buf, "[%s]\n", stanza)
			for _, specString := range paramSpecs {
				fmt.Fprintln(buf, specString)
			}

		}
		fmt.Fprintln(buf, "")
	}
	return buf.String()
}
