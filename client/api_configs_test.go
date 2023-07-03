package client

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func TestConfigsSourcetypes(t *testing.T) {
	newSourcetype := uuid.New().String()[0:8] + "-sourcetype"

	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	propsCol := NewConfigsCollection(ss, "props")

	props := url.Values{}
	props.Set("TRUNCATE", "5431")
	props.Set("disabled", "1")

	t.Logf("INFO Creating new sourcetype='%s'", newSourcetype)
	entry, err := propsCol.CreateConfig(newSourcetype, &props)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("INFO Created: %s - %s", entry.Name, entry.Id)

	t.Logf("INFO Reading infos for sourcetype='%s'", newSourcetype)
	checkEntry, err := propsCol.Get(newSourcetype)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if checkEntry.Content["TRUNCATE"] != "5431" {
		t.Errorf("Created sourcetype '%s' has incorrect TRUNCATE setting. Expected=%s found=%s", newSourcetype, "5431", checkEntry.Content["TRUNCATE"])
	}

	t.Logf("INFO Updating sourcetype='%s'", newSourcetype)
	props = url.Values{}
	props.Set("TRUNCATE", "1234")
	err = propsCol.Update(entry.Name, &props)
	if err != nil {
		t.Error(err)
	}

	checkEntry, err = propsCol.Get(newSourcetype)
	if err != nil {
		t.Error(err)
	}

	if checkEntry.Content["TRUNCATE"] != "1234" {
		t.Errorf("Created sourcetype '%s' has incorrect TRUNCATE setting. Expected=%s found=%s", newSourcetype, "1234", checkEntry.Content["TRUNCATE"])
	}

	t.Logf("INFO Deleting sourcetype='%s'", newSourcetype)
	err = propsCol.Delete(entry.Name)
	if err != nil {
		t.Error(err)
	}

}

/*
func TestConfigCustom(t *testing.T) {
	newConfig := uuid.New().String()[0:8] + "-paolo"

	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	customConfCol := NewConfigsCollection(ss, "paolo")

	props := url.Values{}
	props.Set("paolo1", "5431")
	props.Set("paolo2", "1")

	t.Logf("INFO Creating new config='%s'", newConfig)
	entry, err := customConfCol.CreateConfig(newConfig, &props)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("INFO Created: %s - %s", entry.Name, entry.Id)

	t.Logf("INFO Reading infos for config='%s'", newConfig)
	checkEntry, err := customConfCol.Get(newConfig)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if checkEntry.Content["paolo1"] != "5431" {
		t.Errorf("Created config '%s' has incorrect paolo1 setting. Expected=%s found=%s", newConfig, "5431", checkEntry.Content["paolo1"])
	}

	t.Logf("INFO Updating config='%s'", newConfig)
	props = url.Values{}
	props.Set("paolo1", "1234")
	err = customConfCol.Update(entry.Name, &props)
	if err != nil {
		t.Error(err)
	}

	checkEntry, err = customConfCol.Get(newConfig)
	if err != nil {
		t.Error(err)
	}
	if checkEntry.Content["paolo1"] != "1234" {
		t.Errorf("Created confg '%s' has incorrect paolo1 setting. Expected=%s found=%s", newConfig, "1234", checkEntry.Content["paolo1"])
	}

	t.Logf("INFO Deleting sourcetype='%s'", newConfig)
	err = customConfCol.Delete(entry.Name)
	if err != nil {
		t.Error(err)
	}
}
*/
