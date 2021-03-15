default: build

all: build test benchmark

help: 
	@echo "Available make commands:"
	@echo "  - all (build + test + benchmark)"
	@echo "  - build"
	@echo "  - test"
	@echo "  - benchmark"
	@echo "  - profile (profile execution and expose results on a http server)"

build: 
	@echo "Building library"
	@echo "================"
	@echo "Results will be discarded as this library does NOT contain a main() function\n"
	go build

test:
	@echo "\n\nRunning 'go test'"
	@echo "=================\n"
	go test

benchmark:
	@echo "\nStarting benchmarking"
	@echo "=====================\n"
	go test -bench .

profile: 
	@echo "\n\nStarting benchmarking and profiling"
	@echo "===================================\n"
	@echo "Generating profile file"
	@echo "-----------------------\n"
	
	go test -bench . -cpuprofile /tmp/cpu.pprof
	
	@echo "\nStarting profiler on port :8081"
	@echo "-------------------------------\n"

	go tool pprof -http="0.0.0.0:8081" /tmp/cpu.pprof
