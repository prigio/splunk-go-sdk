package splunkd

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func TestConfigsSourcetypes(t *testing.T) {
	newSourcetype := uuid.New().String()[0:8] + "-sourcetype"

	ss := mustLoginToSplunk(t)

	propsCol := NewConfigsCollection(ss, "props")

	props := url.Values{}
	props.Set("TRUNCATE", "5431")
	props.Set("disabled", "1")

	t.Logf("INFO Creating new sourcetype='%s'", newSourcetype)
	entry, err := propsCol.CreateStanza(newSourcetype, &props)
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

func TestConfigResourceGetString(t *testing.T) {
	cr := make(ConfigResource, 0)

	cr["keyStr"] = "somestring"
	vStr, err := cr.GetString("keyStr")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vStr != "somestring" {
		t.Errorf("ConfigResource map: wrong string value retrieved. Expected '%s', actual: '%s'", "somestring", vStr)
	}

	cr["keyInt"] = 1
	vStr, err = cr.GetString("keyInt")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vStr != "1" {
		t.Errorf("ConfigResource map: wrong string value retrieved. Expected '%d', actual: '%s'", 1, vStr)
	}

	cr["keyIntAsString"] = "1"
	vStr, err = cr.GetString("keyIntAsString")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vStr != "1" {
		t.Errorf("ConfigResource map: wrong string value retrieved")
	}

	cr["keyFloat32"] = float32(3.14)
	vStr, err = cr.GetString("keyFloat32")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vStr != "3.14" {
		t.Errorf("ConfigResource map: wrong string value retrieved from '%s'. Expected '%v', actual '%v'", "keyFloat32", 3, vStr)
	}

}

func TestConfigResourceGetInt(t *testing.T) {
	cr := make(ConfigResource, 0)

	cr["keyStr"] = "somestring"
	_, err = cr.GetInt("keyStr")
	if err == nil {
		t.Errorf("ConfigResource map: incorrect nil error return value when trying to read string '%v' as int", cr["keyStr"])
	}

	cr["keyInt"] = 1
	vInt, err := cr.GetInt("keyInt")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	}
	if vInt != 1 {
		t.Errorf("ConfigResource map: wrong string value retrieved. Expected '%d', actual: '%d'", 1, vInt)
	}

	cr["keyIntAsString"] = "1"
	vInt, err = cr.GetInt("keyIntAsString")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vInt != 1 {
		t.Errorf("ConfigResource map: wrong int value retrieved")
	}

	cr["keyFloat32"] = float32(3.14)
	vInt, err = cr.GetInt("keyFloat32")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vInt != 3 {
		t.Errorf("ConfigResource map: wrong int value retrieved from '%s'. Expected '%v', actual '%v'", "keyFloat32", 3, vInt)
	}

	cr["keyFloatAsString"] = "3.14"
	_, err = cr.GetInt("keyFloatAsString")
	if err == nil {
		t.Errorf("ConfigResource map: expecting error when trying to parse string-based float valut to int. found nil")
	}
}

func TestConfigResourceGetFloat(t *testing.T) {
	cr := make(ConfigResource, 0)

	cr["keyStr"] = "somestring"
	_, err = cr.GetFloat("keyStr")
	if err == nil {
		t.Errorf("ConfigResource map: incorrect nil error return value when trying to read string '%v' as float", cr["keyStr"])
	}

	cr["keyInt"] = 1
	vFloat, err := cr.GetFloat("keyInt")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	}
	if vFloat != 1 {
		t.Errorf("ConfigResource map: wrong string value retrieved. Expected '%d', actual: '%v'", 1, vFloat)
	}

	cr["keyIntAsString"] = "1"
	vFloat, err = cr.GetFloat("keyIntAsString")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vFloat != 1 {
		t.Errorf("ConfigResource map: wrong float value retrieved")
	}

	cr["keyFloat32"] = float32(3.14)
	vFloat, err = cr.GetFloat("keyFloat32")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vFloat != 3.14 {
		t.Errorf("ConfigResource map: wrong float value retrieved from '%s'. Expected '%v', actual '%v'", "keyFloat32", 3.14, vFloat)
	}

	cr["keyFloatAsString"] = "3.14"
	vFloat, err = cr.GetFloat("keyFloatAsString")
	if err != nil {
		t.Errorf("ConfigResource map: %s", err.Error())
	} else if vFloat != 3.14 {
		t.Errorf("ConfigResource map: wrong float value retrieved from '%s'. Expected '%v', actual '%v'", "keyFloatAsString", 3.14, vFloat)
	}
}

func TestConfigsNS(t *testing.T) {
	ss := mustLoginToSplunk(t)
	sourceType := "sourcetype-" + uuid.New().String()[0:5]
	accountUser := "user"
	accountTest := "test"

	// create testing users if necessary
	for _, u := range []string{accountUser, accountTest} {
		if !ss.GetUsers().Exists(u) {
			t.Logf("INFO Creating necessary user='%s'", u)
			_, err := ss.GetUsers().CreateUser(u, UserResource{Password: uuid.New().String()[0:16], Roles: []string{"power"}})
			if err != nil {
				t.Errorf(err.Error())
				t.FailNow()
			}
		}
	}

	propsColNS := NewConfigsCollectionNS(ss, "props", accountUser, "search")

	nsTestLauncher, _ := NewNamespace(accountTest, "launcher", SplunkSharingApp)
	t.Logf("TestConfigNS: Working with sourcetype %s", sourceType)

	params := url.Values{}
	params.Set("REPORT-fields", "someConfigFor-"+accountUser)
	ce, err := propsColNS.CreateStanza(sourceType, &params)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	ce.ACL.Sharing = string(SplunkSharingApp)
	if err := propsColNS.UpdateACL(ce.Name, ce.ACL); err != nil {
		t.Error(err)
		t.FailNow()
	}
	checkCE, _ := propsColNS.Get(sourceType)
	if checkCE.ACL.Sharing != ce.ACL.Sharing {
		t.Errorf("TestConfigsNS: sharing for resource '%s/%s' is '%s', expected '%s'", "props", sourceType, checkCE.ACL.Sharing, ce.ACL.Sharing)
	}

	params.Set("REPORT-fields", "testSearchLauncher")
	_, err = propsColNS.CreateNS(nsTestLauncher, sourceType, &params)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

}

/*
func TestConfigCustom(t *testing.T) {
	newConfig := uuid.New().String()[0:8] + "-paolo"

	ss := mustLoginToSplunk(t)

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
