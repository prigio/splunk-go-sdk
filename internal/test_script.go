package main

import (
	"fmt"
	"time"

	"./modinputs"
)

func main() {

	se := &modinputs.SplunkEvent{}

	se.Data = `action="value"`
	se.Time = float64(time.Now().UnixNano()) / 1000000000.0
	se.SourceType = "ST"
	se.Index = "IDX"
	se.Host = "H"
	se.Source = "S"
	se.Stanza = ""
	se.Done = false
	se.Unbroken = true

	xml, err := se.XML()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	fmt.Println(string(xml))

	s := `<input>
	<server_host>myHost</server_host>
	<server_uri>https://127.0.0.1:8089</server_uri>
	<session_key>123102983109283019283</session_key>
	<checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs</checkpoint_dir>
	<configuration>
	  <stanza name="myScheme://aaa">
		  <param name="param1">value1</param>
		  <param name="param2">value2</param>
		  <param name="disabled">0</param>
		  <param name="index">default</param>
	  </stanza>
	  <stanza name="myScheme://bbb">
		  <param name="param1">value11</param>
		  <param name="param2">value22</param>
		  <param name="disabled">0</param>
		  <param name="index">default</param>
	  </stanza>
	</configuration>
  </input>`

	script := &modinputs.ModularInput{}
	script.Title = "Test script"
	script.Description = "Whatever"
	script.StreamingMode = modinputs.SchemeStreamingModeXML
	script.UseExternalValidation = true
	script.UseSingleInstance = false
	script.Args = []modinputs.ModInputArg{
		{"Arg1", "Name of Arg1", "Desc of arg1", modinputs.ArgDataTypeStr, true, false, "", "", ""},
		{"Arg2", "Name of Arg2", "Desc of arg2", modinputs.ArgDataTypeStr, true, false, "", "", ""},
	}
	if scheme, err := script.XMLScheme(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(string(scheme))
	}

	if xerr := script.ParseConfig([]byte(s)); xerr != nil {
		fmt.Println(xerr.Error())
	} else {
		fmt.Println(script.SessionKey)
		for _, stanza := range script.Configuration.Stanzas {
			fmt.Printf("Name:%s\n", stanza.Name)
			for _, param := range stanza.Params {
				fmt.Printf("Name:%s, val:%s\n", param.Name, param.Value)
			}
		}
	}

}
