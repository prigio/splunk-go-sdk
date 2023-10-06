/*
Package alertactions defines utilities to develop a Splunk alert-action script.

The main component of this package is type [AlertAction], which provides the base for development of the alert action.

Additionally, type [params.Param] is also necessary to define the parameters which end-users see in the splunk UI.
This type can also represent global paramenters, whose value is defined in some standard or custom splunk configuration file.

# To develop an alert action

 1. Create a `main()` function which instantiates an [AlertAction] struct using the provided utility.

 2. 'Register' the 'global' parameters, if any, which hold global configurations for the alert action. These are generally not configurable by the end-user and can be present in a custom .conf file.

 3. 'Register' the alert parameters, if any. These are the ones appearing for the end-user in the UI to configure the alert.

 4. Optionally define a parameter validation function and register it within the alert action. See type [AlertingFunc] for this and [AlertAction.RegisterValidationFunc].

 5. Define an alerting function, which holds the actual action performed by your alert action and register it.See type [AlertingFunc] for this and [AlertAction.RegisterAlertFunc].

    - use the [AlertAction.Log] to emit logs for administrators (index=_internal).

    - use the [AlertAction.RegisterEndUserLogger] function to configure logging for the end-user, that is writing events in a user-configured index.

    - use the [AlertAction.LogForEndUser] function to emit logs for end-users.

    - use the [AlertAction.GetSplunkService] function to retrieve a client to the splunkd API, logged in using the credential of the splunk user owning the executing alert.

 5. Invoke the [AlertAction.Run] function to start its processing.

 6. Compile your alert action for the target architectures.

The compiled alert action provides a number of command-line based utilities to generate necessary Splunk configuration files (`.conf`), specification files (`.conf.spec`), example configurations etc.
Invoke the compiled alert action with the `-h` parameter to see them.

# How does Splunk invoke the alert action?

 1. Splunk invokes the compiled alert action (by default) using the "--execute" command-line flag.

 2. Splunk provides the JSON-formatted run-time configuration via STDIN. See example next.

 3. The alert actions main() code is executed.

    - The main() instantiates a new AlertAction, configures its parameters and actual validation and alerting functions

    - The main() method calls the [AlertAction.Run] method

 6. The [AlertAction.Run] method

    - parses run-time configurations from STDIN,

    - instantiates a client to the splunkd API,

    - prepares the logging infrastructure,

    - assigns actual values to the parameters,

    - reads-in the values of the global parameters,

    - invokes the provided alerting function,

    - logs completion of the execution, or any errors if returned by the alerting function.

# Example of the run-time configuration sent by Splunk via STDIN

	{
	"app": "search",
	"owner": "admin",
	"results_file": "/Applications/Splunk/var/run/splunk/dispatch/scheduler__.../results.csv.gz",
	"results_link": "http://mac.local:8000/app/search/@go?sid=scheduler__...",
	"search_uri": "/servicesNS/admin/search/saved/searches/jira+test",
	"server_host": "mac.local",
	"server_uri": "https://127.0.0.1:8089",
	"session_key": "onrmI0....",
	"sid": "scheduler__...",
	"search_name": "<name of saved search>",
	"configuration": {
		"parameter1": "val1",
		"parameter2": "val2",
		"last-parameter": "valLast",
	},
	"result": {
		"_bkt": "_internal~9~0132F894-2D30-4834-A898-90FF4468F923",
		"_cd": "9:114711",
		"_eventtype_color": "",
		"_indextime": "1689609697",
		"_kv": "1",
		"_raw": "<content of first log result>",
		"_serial": "0",
		"_si": ["mac.local","_internal"],
		"_sourcetype": "splunkd",
		"_subsecond": ".996",
		"_time": "1689609696.996",
		"date_hour": "18",
		"date_mday": "17",
		"date_minute": "1",
		"date_month": "july",
		"date_second": "36",
		"date_wday": "monday",
		"date_year": "2023",
		"date_zone": "120",
		"host": "mac.local",
		"index": "_internal",
		"linecount": "1",
		"punct": "--_::._+____-_=,_=,_=,_=,_=,_=,_=,_=,_=,_=,_=.,_=.",
		"source": "/opt/splunk/var/log/splunk/metrics.log",
		"sourcetype": "splunkd",
		"splunk_server": "mac.local",
		"splunk_server_group": "",
		"<any other search-specific fields>": "...."
		}
	}
*/
package alertactions
