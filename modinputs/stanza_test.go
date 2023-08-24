package modinputs

import (
	"testing"
)

func TestParameterRetrieval(t *testing.T) {
	s := &Stanza{
		Name: "teststz://t1",
		App:  "testapp",
		Params: []Param{
			{Name: "p1", Value: "v1"},
			{Name: "source", Value: "src"},
			{Name: "index", Value: "main"},
			{Name: "sourcetype", Value: "st"},
		},
	}
	if s.Index() != "main" {
		t.Errorf(`stanza.Index: Incorrect value returned: expected="%s" got="%s"`, "main", s.Index())
	}
	if s.Source() != "src" {
		t.Errorf(`stanza.Source: Incorrect value returned: expected="%s" got="%s"`, "src", s.Source())
	}
	if s.Param("p1") != "v1" {
		t.Errorf(`stanza.Param: Incorrect value returned for parameter: expected="%s" got="%s"`, "v1", s.Param("p1"))
	}
	if s.Sourcetype() != "st" {
		t.Errorf(`stanza.Sourcetype: Incorrect value returned: expected="%s" got="%s"`, "st", s.Sourcetype())
	}

}

func TestSchemeAndInputName(t *testing.T) {
	s := &Stanza{
		Name: "teststz://t1",
		App:  "testapp",
		Params: []Param{
			{Name: "p1", Value: "v1"},
			{Name: "source", Value: "src"},
			{Name: "index", Value: "main"},
			{Name: "sourcetype", Value: "st"},
		},
	}
	if s.Scheme() != "teststz" {
		t.Errorf(`stanza.Scheme: Incorrect value returned: expected="%s" got="%s"`, "teststz", s.Scheme())
	}

	if s.InputName() != "t1" {
		t.Errorf(`stanza.InputName: Incorrect value returned: expected="%s" got="%s"`, "t1", s.InputName())
	}

}

func TestParamCSV(t *testing.T) {
	s := &Stanza{
		Name: "teststz://t1",
		App:  "testapp",
		Params: []Param{
			{Name: "l1", Value: "   v1, v2  , v3,v4  "},
			{Name: "emptyList", Value: ""},
		},
	}
	if s.ParamAsCSVList("missing") != nil {
		t.Errorf(`stanza.ParamAsCSVList: Incorrect value returned for non-existing parameter: expected=nil got="%v"`, s.ParamAsCSVList("missing"))
	}

	if s.ParamAsCSVList("emptyList") != nil {
		t.Errorf(`stanza.ParamAsCSVList: Incorrect value returned for emtpy parametter: expected=nil got="%v"`, s.ParamAsCSVList("emptyList"))
	}

	r := s.ParamAsCSVList("l1")
	if len(r) != 4 {
		t.Errorf(`stanza.ParamAsCSVList: Incorrect number of elements returned: expected=%d got="%d"`, 4, len(r))
	}
	if !(r[0] == "v1" && r[1] == "v2" && r[2] == "v3" && r[3] == "v4") {
		t.Errorf(`stanza.ParamAsCSVList: Incorrect values of elements returned: expected="'v1', 'v2', 'v3', 'v4'"  got="%v"`, r)
	}

}
