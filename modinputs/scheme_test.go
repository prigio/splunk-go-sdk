package modinputs

import (
	"fmt"
	"testing"
)

func TestSchemeXML(t *testing.T) {
	scheme := &ModInputScheme{
		StanzaName:            "teststanzaname",
		Title:                 "Test Scheme",
		Description:           "This is the description of the test scheme",
		UseExternalValidation: false,
		UseSingleInstance:     false,
		streamingMode:         "xml",
		Args:                  []ModInputArg{},
	}
	argOne := &ModInputArg{
		Name:             "one",
		Title:            "Param One",
		Description:      "Test parameter one, of string type, without validation",
		DataType:         ArgDataTypeStr,
		RequiredOnCreate: true,
		RequiredOnEdit:   true,
		//		Validation: "",
		//		CustomValidation: "",
		//		CustomValidationErrMessage: "",
	}
	scheme.AddArgument(argOne)
	// when modifying this, you need to pay attention that the editor
	// may want to substitute spaces with tabs, thus causing tests to fail.
	expectedScheme := fmt.Sprintf(`<scheme>
  <title>%s</title>
  <description>%s</description>
  <use_external_validation>false</use_external_validation>
  <use_single_instance>false</use_single_instance>
  <streaming_mode>xml</streaming_mode>
  <endpoint>
    <args>
      <arg name="%s">
        <title>%s</title>
        <description>%s</description>
        <data_type>%s</data_type>
        <required_on_create>true</required_on_create>
        <required_on_edit>true</required_on_edit>
      </arg>
    </args>
  </endpoint>
</scheme>`, scheme.Title, scheme.Description, scheme.Args[0].Name, scheme.Args[0].Title, scheme.Args[0].Description, scheme.Args[0].DataType)

	generatedScheme, _ := scheme.PrintXMLScheme()

	if string(generatedScheme) != expectedScheme {
		t.Errorf("PrintXMLScheme() did not return the expacted value.\n## Expected=\n'%s'\n ## Generated:\n'%s'\n", expectedScheme, string(generatedScheme))
	}
}
