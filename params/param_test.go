package params

import (
	"testing"
)

func TestParamValues(t *testing.T) {
	p := Param{
		title:        "title",
		configFile:   "something",
		name:         "name",
		description:  "descr",
		defaultValue: "a default",
		placeholder:  "a placeholder",
		required:     false,
	}

	if v, _ := p.GetValue(nil); v != p.defaultValue {
		t.Errorf("Parameter did not return correct default value")
	}

	p.ForceValue("a value")
	if v, _ := p.GetValue(nil); v != "a value" {
		t.Errorf("Parameter did not return correct actual value")
	}

	p.ForceValue("another value")
	if v, _ := p.GetValue(nil); v != "another value" {
		t.Errorf("Parameter did not return correct actual value")
	}
}

func TestParamAcceptableValues(t *testing.T) {
	p := Param{
		title:        "title",
		name:         "name",
		description:  "descr",
		defaultValue: "a default",
		placeholder:  "a placeholder",
		required:     false,
	}

	if p.GetChoices() == nil {
		t.Error("GetChoiceValues returned a nil value instead of an empty slice when there were no acceptable choices")
	}

	if len(p.GetChoices()) > 0 {
		t.Error("GetChoiceValues returned a non-empty slice when there were no acceptable choices")
	}

	p.AddChoice("c1", "Choice 1")

	if len(p.availableOptions) != 1 {
		t.Errorf("availableChoices has wrong length after adding first element. Expected=%v, Actual=%v. Contents: %v", 1, len(p.availableOptions), p.availableOptions)
	}

	p.AddChoice("c2", "Choice 2")

	if v, _ := p.GetValue(nil); v != p.defaultValue {
		t.Error("Parameter did not return correct default value")
	}
	if err := p.ForceValue("not accepted value"); err == nil {
		t.Error("SetValue did not return an error when provided with a non-acceptable value")
	}

	if err := p.ForceValue("c2"); err != nil {
		t.Error("SetValue returned an error when provided with an acceptable value")
	}

	if v, _ := p.GetValue(nil); v != "c2" {
		t.Error("Parameter did not return correct actual value")
	}

	if err := p.AddChoice("c3", "Choice 3"); err != nil {
		t.Errorf("AddChoice returned an error when adding a valild choice %s", err.Error())
	}

	if err := p.AddChoice("", "this should raise an error"); err == nil {
		t.Error("AddChoice did not return an when adding an invalid choice")
	}
	if err := p.ForceValue("c3"); err != nil {
		t.Error("SetValue returned an error when provided with an acceptable value")
	}

	if err := p.AddChoice("c1", "this should raise an error"); err == nil {
		t.Error("AddChoice did not return an when adding a duplicated choice")
	}
	if len(p.GetChoices()) != 3 {
		t.Errorf("GetChoiceValues returned the wrong number of choices. Expected=%v, Actual=%v. Contents: %v", 3, len(p.GetChoices()), p.GetChoices())
	}

}