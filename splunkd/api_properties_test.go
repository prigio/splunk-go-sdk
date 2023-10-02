package splunkd

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPropertyGetStanza(t *testing.T) {

	var (
		props map[string]string
		err   error
	)

	ss := mustLoginToSplunk(t)

	propsCol := NewPropertiesCollection(ss, "props")

	t.Logf("INFO Reading infos for sourcetype='catalina'")
	props, err = propsCol.GetStanza("catalina")
	if err != nil {
		t.Errorf("Requesting stanza 'catalina' failed: %s.", err.Error())
		t.FailNow()
	}
	if !strings.Contains(props["description"], "Catalina") {
		t.Errorf("Wrong description recovered. %v", props)
		t.FailNow()
	}
}

func TestPropertyGetProperty(t *testing.T) {
	var (
		val string
		err error
	)
	ss := mustLoginToSplunk(t)
	propsCol := NewPropertiesCollection(ss, "props")

	t.Logf("INFO Reading 'description' for sourcetype='catalina'")
	val, err = propsCol.GetProperty("catalina", "description")
	if err != nil {
		t.Errorf("Requesting property 'catalina/description' failed: %s", err.Error())
	}
	if !strings.Contains(val, "Catalina") {
		t.Errorf("Wrong description recovered. '%s'", val)
		t.FailNow()
	}
}

func TestPropertyCreateStanza(t *testing.T) {
	var (
		err           error
		props         map[string]string
		newSourcetype string
	)
	newSourcetype = uuid.New().String()[0:8] + "-sourcetype"
	ss := mustLoginToSplunk(t)
	propsCol := NewPropertiesCollection(ss, "props")

	t.Logf("INFO Creating stanza for sourcetype '%s'", newSourcetype)

	err = propsCol.CreateStanza(newSourcetype, nil)
	if err != nil {
		t.Errorf("Creating stanza '%s' failed: %s", newSourcetype, err.Error())
	}
	params := url.Values{}
	params.Set("description", "My test")
	params.Set("LINE_BREAKER", "TL;DR")

	t.Logf("INFO Adding properties to stanza for sourcetype '%s'. %v", newSourcetype, params)
	err = propsCol.SetProperties(newSourcetype, &params)
	if err != nil {
		t.Errorf("Adding properties for '%s' failed: %s", newSourcetype, err.Error())
		t.FailNow()
	}

	props, err = propsCol.GetStanza(newSourcetype)
	if err != nil {
		t.Errorf("Getting properties for '%s' failed: %s", newSourcetype, err.Error())
		t.FailNow()
	}
	if props["description"] != "My test" {
		t.Errorf("Retrieved properties for '%s/%s' do not match the expected values: expected 'My test' found: '%s'", newSourcetype, "description", props["description"])
		t.FailNow()
	}
}

func TestPropertyDelete(t *testing.T) {
	var (
		err           error
		newSourcetype string
	)
	newSourcetype = uuid.New().String()[0:8] + "-sourcetype"
	ss := mustLoginToSplunk(t)
	propsCol := NewPropertiesCollection(ss, "props")

	t.Logf("INFO Creating stanza for sourcetype '%s'", newSourcetype)

	err = propsCol.CreateStanza(newSourcetype, nil)
	if err != nil {
		t.Errorf("Creating stanza '%s' failed: %s", newSourcetype, err.Error())
	}
	params := url.Values{}
	params.Set("description", "My test")
	params.Set("LINE_BREAKER", "TL;DR")

	t.Logf("INFO Adding properties to stanza for sourcetype '%s'. %v", newSourcetype, params)
	err = propsCol.SetProperties(newSourcetype, &params)
	if err != nil {
		t.Errorf("Adding properties for '%s' failed: %s", newSourcetype, err.Error())
		t.FailNow()
	}

	err = propsCol.Delete(newSourcetype)
	if err != nil {
		t.Errorf("Deleting stanza '%s' failed: %s", newSourcetype, err.Error())
		t.FailNow()
	}

}

func TestPropertyList(t *testing.T) {

	ss := mustLoginToSplunk(t)
	propsCol := NewPropertiesCollection(ss, "props")

	entries, err := propsCol.List()
	if err != nil {
		t.Errorf("Listing stanzas failed: %s", err.Error())
		t.FailNow()
	}
	t.Logf("Found %d", len(entries))
	var i int
	for i < 5 {
		t.Logf("Entry %d: '%s',", i, entries[i].Name)
		i += 1
	}

}
