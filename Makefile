#Environment settings for cross compilation
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

GOCONTAINERIMAGE=golang:1.20
ENV_OSX=--build-arg GOOS=darwin --build-arg GOARCH=amd64
ENV_LIN=--build-arg GOOS=linux --build-arg GOARCH=amd64

default: check

# the following builds and tests the libraries
check: build-modinputs test-modinputs build-client test-client

splunkd: splunk-test-start
	@echo ""
	@echo "Starting 'go build + test' for the splunkd package"
	@echo "This will throw errors if go cannot build the library."
	@echo "Note: output of a built library gets discarded by go, as there is no executable in it"
	@echo ""
	docker run --rm -v $(PWD):/src/ --workdir /src/splunkd ${GOCONTAINERIMAGE} go test
	docker run --rm -v $(PWD):/src/ --workdir /src/splunkd ${GOCONTAINERIMAGE} go build

build-modinputs: splunk-test-start
	@echo ""
	@echo "Starting 'go build' for the modinputs package"
	@echo "This will throw errors if go cannot build the library."
	@echo "Note: output of a built library gets discarded by go, as there is no executable in it"
	@echo ""
	docker run --rm -v $(PWD):/src/ --workdir /src/modinputs ${GOCONTAINERIMAGE} go build
	docker run --rm -v $(PWD):/src/ --workdir /src/modinputs ${GOCONTAINERIMAGE} go test

test-modinputs:
	@echo ""
	@echo "Starting 'go test' for the modinputs package"
	@echo ""
	docker run --rm -v $(PWD):/src/ --workdir /src/modinputs ${GOCONTAINERIMAGE} go test

# The following builds and executes the examples contained within the examples folder
build-examples: build-hello
run-examples: run-hello

build-hello:
	@echo "Building the example modular input in examples/hello using Dockerfile 'Dockerfile_examples_hello'"
	docker build -f Dockerfile_examples_hello -t modinputs-hello --platform linux ${ENV_LIN} .

run-hello:
	@echo "Starting container with built "hello" modular input. Paste a XML configuration and hit Ctrl-D to start the input processing."
	docker run --rm -ti modinputs-hello

splunk-test-start:
	docker compose up -d

splunk-test-stop:
	docker compose down