/*
Package alertactions defines utilities to develop a Splunk alert-action script.

The main component of this package is type [AlertAction], which provides the base for development of the alert action.

Additionally, type [Param] is also necessary to define the parameters which end-users see in the splunk UI.
This type can also represent global paramenters, whose value is defined in some standard or custom splunk configuration file.

 1. Splunk invokes the compiled alert action (by default) using the "--execute" command-line flag.

 2. Splunk provides the JSON-formatted run-time configuration via STDIN.

 3. This executes the main() code provided by the developer of the actual alert action,

 4. The main() instantiates a new AlertAction, configures its parameters and actual alerting function,

 5. The main() method calls the "Run()" method

 6. The Run() method parses run-time configurations, reads-in the values of the global parameters and invokes the provided alerting function.

The following is an example of the run-time configuration sent by Splunk:

	{
	"app": "search",
	"owner": "admin",
	"results_file": "/Applications/Splunk/var/run/splunk/dispatch/scheduler__admin__search__<redacted>_at_1689609720_2/results.csv.gz",
	"results_link": "http://mac.local:8000/app/search/@go?sid=scheduler__admin__search__<redacted>_at_1689609720_2",
	"search_uri": "/servicesNS/admin/search/saved/searches/jira+test",
	"server_host": "mac.local",
	"server_uri": "https://127.0.0.1:8089",
	"session_key": "onrmI0....",
	"sid": "scheduler__admin__search__<redacted>_at_1689609720_2",
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
