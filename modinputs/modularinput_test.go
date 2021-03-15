package modinputs

import (
	"fmt"
	"testing"
)

func TestAddArgument(t *testing.T) {
	mi := &ModularInput{}

	cases := []struct {
		name             string
		title            string
		description      string
		dataType         string
		validation       string
		requiredOnCreate bool
		requireOnEdit    bool
	}{
		{"param1", "Param1", "A string parameter", ArgDataTypeStr, "", true, false},
		{"param2", "Param2", "A numeric parameter", ArgDataTypeNumber, "", true, true},
	}

	for i, v := range cases {
		mi.AddArgument(v.name, v.title, v.description, v.dataType, v.validation, v.requiredOnCreate, v.requireOnEdit)
		if len(mi.Args) != i+1 {
			t.Errorf("Argument %s not added", v.name)
		}
		if mi.Args[i].Name != v.name {
			t.Errorf(`Wrong name added: expected="%s", got="%s"`, v.name, mi.Args[i].Name)
		}
		if mi.Args[i].Title != v.title {
			t.Errorf(`Wrong title added: expected="%s", got="%s"`, v.title, mi.Args[i].Title)
		}
		if mi.Args[i].Description != v.description {
			t.Errorf(`Wrong description added: expected="%s", got="%s"`, v.description, mi.Args[i].Description)
		}
	}
}
func TestSchemeXML(t *testing.T) {
	mi := &ModularInput{
		StanzaName:            "teststanzaname",
		Title:                 "Test Scheme",
		Description:           "This is the description of the test scheme",
		UseExternalValidation: false,
		UseSingleInstance:     false,
	}
	mi.AddArgument("one", "Param one", "Test parameter one, of string type, without validation", ArgDataTypeStr, "", true, true)

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
</scheme>`, mi.Title, mi.Description, mi.Args[0].Name, mi.Args[0].Title, mi.Args[0].Description, mi.Args[0].DataType)

	generatedScheme, _ := mi.getXMLScheme()

	if string(generatedScheme) != expectedScheme {
		t.Errorf("PrintXMLScheme() did not return the expacted value.\n## Expected=\n'%s'\n ## Generated:\n'%s'\n", expectedScheme, string(generatedScheme))
	}
}

func TestEvent(t *testing.T) {
	mi := ModularInput{
		StanzaName:            "teststanzaname",
		Title:                 "Test Scheme",
		Description:           "This is the description of the test scheme",
		UseExternalValidation: false,
		UseSingleInstance:     false,
	}
	st := Stanza{
		Name: "testscheme://testinputname",
		App:  "testapp",
		Params: []Param{
			{Name: "sourcetype", Value: "testsourcetype"},
			{Name: "index", Value: "testindex"},
			{Name: "host", Value: "testhost"},
			{Name: "source", Value: "testsource"},
		},
	}
	ev := mi.NewDefaultEvent(&st)
	ev.Data = "some log message"
	if ev.Stanza != st.Name {
		t.Errorf("Event's stanza does not match: expected=%s got=%s", st.Name, ev.Stanza)
	}
	if ev.Stanza != st.Name {
		t.Errorf("Event's index does not match: expected=%s got=%s", st.Param("index"), ev.Index)
	}
	if ev.Stanza != st.Name {
		t.Errorf("Event's sourcetype does not match: expected=%s got=%s", st.Param("sourcetype"), ev.SourceType)
	}
	if ev.Stanza != st.Name {
		t.Errorf("Event's source does not match: expected=%s got=%s", st.Param("source"), ev.Source)
	}
	if ev.Stanza != st.Name {
		t.Errorf("Event's host does not match: expected=%s got=%s", st.Param("host"), ev.Host)
	}
	t.Log("Writing the event to stdout")
	if err := mi.WriteToSplunk(ev); err != nil {
		t.Errorf("Writing event to Stdout raised error: %s", err.Error())
	}
	if mi.cntDataEventsGeneratedbyStanza != 1 {
		t.Errorf("ModularInput did not track generated event within cntDataEventsGeneratedbyStanza. expected=1 got=%d", mi.cntDataEventsGeneratedbyStanza)
	}
	if mi.cntDataEventsGeneratedTotal != 1 {
		t.Errorf("ModularInput did not track generated event within cntDataEventsGeneratedTotal. expected=1 got=%d", mi.cntDataEventsGeneratedTotal)
	}

}
