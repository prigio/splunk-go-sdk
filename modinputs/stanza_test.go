package modinputs

import (
	"testing"
)

func TestParameterRetrieval(t *testing.T) {
	s := &Stanza{
		Name: "teststz://t1",
		App:  "testapp",
		Params: []Param{
			Param{Name: "p1", Value: "v1"},
			Param{Name: "source", Value: "src"},
			Param{Name: "index", Value: "main"},
			Param{Name: "sourcetype", Value: "st"},
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
			Param{Name: "p1", Value: "v1"},
			Param{Name: "source", Value: "src"},
			Param{Name: "index", Value: "main"},
			Param{Name: "sourcetype", Value: "st"},
		},
	}
	if s.Scheme() != "teststz" {
		t.Errorf(`stanza.Scheme: Incorrect value returned: expected="%s" got="%s"`, "teststz", s.Scheme())
	}

	if s.InputName() != "t1" {
		t.Errorf(`stanza.InputName: Incorrect value returned: expected="%s" got="%s"`, "t1", s.InputName())
	}

}
