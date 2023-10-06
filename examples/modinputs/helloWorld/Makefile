default: all

all: print_config validate execute

print_config:
	@echo "\n\n## Running with parameter --example-config\n\n"
	go run main.go --example-config

validate:
	@echo "\n\n## Starting validation mode\n\n"
	go run main.go --validate-arguments < samplevalid.xml

execute: 
	@echo "\n\n## Starting execution mode\n\n"
	go run main.go < sampleconfig.xml
