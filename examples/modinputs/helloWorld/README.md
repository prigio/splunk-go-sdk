# Test modular input: Hello world!
This is a sample test to allow you to actually compile the library. 

## Features
This uses `single-instance` mode, in that Splunk invokes the modular input only once and sends ALL related configuration stanzas to it. 
The modular input is responsible for the scheduling etc, if necessary.

## Execution

```go
    go run main.go -h
    go run main.go --example-config
    go run main.go --scheme
    go run main.go < sampleconfig.xml
```
