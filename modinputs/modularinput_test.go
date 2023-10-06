package modinputs

import (
	"fmt"
	"testing"
)

func TestAddArgument(t *testing.T) {
	var err error

	mi := &ModularInput{}

	cases := []struct {
		name             string
		title            string
		description      string
		defaultValue     string
		dataType         string
		validation       string
		requiredOnCreate bool
		requireOnEdit    bool
	}{
		{"param1", "Param1", "A string parameter", "", "string", "", true, true},
		{"param2", "Param2", "A numeric parameter", "", "number", "", true, true},
	}

	for i, v := range cases {
		_, err = mi.RegisterNewParam(v.name, v.title, v.description, v.defaultValue, v.dataType, v.requiredOnCreate, false)
		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}
		if len(mi.params) != i+1 {
			t.Errorf("Argument %s not added", v.name)
		}
		if mi.params[i].GetName() != v.name {
			t.Errorf(`Wrong name added: expected="%s", got="%s"`, v.name, mi.params[i].GetName())
		}
		if mi.params[i].GetTitle() != v.title {
			t.Errorf(`Wrong title added: expected="%s", got="%s"`, v.title, mi.params[i].GetTitle())
		}
		if mi.params[i].GetDescription() != v.description {
			t.Errorf(`Wrong description added: expected="%s", got="%s"`, v.description, mi.params[i].GetDescription())
		}
	}
}
func TestSchemeXML(t *testing.T) {
	mi := &ModularInput{
		StanzaName:            "teststanzaname",
		Title:                 "Test Scheme",
		Description:           "This is the description of the test scheme",
		useExternalValidation: false,
		useSingleInstance:     false,
	}
	mi.RegisterNewParam("one", "Param one", "Test parameter one, of string type, without validation", "", "string", true, false)

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
</scheme>`, mi.Title, mi.Description, mi.params[0].GetName(), mi.params[0].GetTitle(), mi.params[0].GetDescription(), mi.params[0].GetDataType())

	generatedScheme, _ := mi.generateXMLScheme()

	if string(generatedScheme) != expectedScheme {
		t.Errorf("PrintXMLScheme() did not return the expacted value.\n## Expected=\n'%s'\n ## Generated:\n'%s'\n", expectedScheme, string(generatedScheme))
	}
}

func TestEvent(t *testing.T) {
	mi := ModularInput{
		StanzaName:            "teststanzaname",
		Title:                 "Test Scheme",
		Description:           "This is the description of the test scheme",
		useExternalValidation: false,
		useSingleInstance:     false,
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
	ev := NewEvent(&st)
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
