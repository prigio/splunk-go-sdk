// Package modinputs provides utilities to write a Splunk modular input.
//
// The following is a representation of the data structures and main public
// methods made available by this library
//
// ModularInput                          SplunkEvent
// │    │                                   │
// │    ├─Log(level, message)               └─NewSplunkEventFromDefaults()
// │    │
// │    ├─WriteToSplunk(event)
// │    │
// │    ├─NewDefaultEvent(stanza)
// │    │
// │    └─Run()
// │
// ├──ModInputConfig
// │  │     │
// │  │     └─LoadConfigFromStdin(xml)
// │  │
// │  └──[]Stanza
// │     │  │
// │     │  └─GetParam(name)
// │     │
// │     └──[]Param
// │
// │
// └──ModInputScheme
//    │     │
//    │     ├─PrintXMLScheme()
//    │     │
//    │     └─ExampleConf()
//    │
//    └──[]ModInputArg
package modinputs
