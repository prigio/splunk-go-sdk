module helloWorld

go 1.20

replace github.com/prigio/splunk-go-sdk/v2 => ../../../

require github.com/prigio/splunk-go-sdk/v2 v2.0.0-alpha-1

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/term v0.9.0 // indirect
)
