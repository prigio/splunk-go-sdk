# Splunk-Go - A simple framework to interact with Splunk using Golang

Reason behind this: sometime you want to deploy python scripts, but there is no python available on the targets (say, a thousand windows servers using the Splunk UF). In those cases, you need to deploy something else. 

Golang is probably the best choice for these use cases as it allows to packate scripts without external library dependencies.

However, how do you make these scripts talk with Splunk?! Enters this repository. 

This repository provides:

- **a framework & library to build modular inputs** - this is already production grade
- **a (simple) splunkd client** - so that your scripts can perform basic communication with splunkd on port `8089`. **In development**.
- a custom alerts framework - if needs arises. 


## Documentation
For detailed documentation, refer to: 

- Modular inputs: [modular inputs libray](modinputs/README.md).

## Usage

### Download library - DEPRECATED
**This instructions only apply to private repositories.**

Note, as this is a PRIVATE repository, you need to:

1. Get a build token from the team managing this repository. 
    The token is configured within github.com. 
2. Configure your build pipeline to use that token when downloading the library.
    Create a file `~/.netrc` with this content:

    ```
        machine github.com login go-build password <build-token>
    ```
	This file will be read from the _go_ utilities to get access to the repository.
3. Configure your build pipeline with the following environment variable:
    `    GOPRIVATE=github.com/*`
4. Test your build pipeline: manually execute the following command and check its results
    `    go get github.com/prigio/splunk-go-sdk/modinputs`

### Import in code
Within your _go_ source files, import the libraries:

```
// somesource.go
package main

import (
	// ... some other imports

	"github.com/prigio/splunk-go-sdk/modinputs"
    // if needed
    // "github.com/prigio/splunk-go-sdk/client"
)
// ... your code

```

## Various resources
These are links for future reference, but not directly relevant to this library.

https://eli.thegreenplace.net/2019/simple-go-project-layout-with-modules/

https://www.docker.com/blog/containerize-your-go-developer-environment-part-1/

