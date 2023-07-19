package splunkd

import (
	"testing"

	"github.com/google/uuid"
)

func TestKVStoreList(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	kvc := ss.GetKVStore()

	t.Log("INFO Retrieving list of KVStore collections")
	kvceList, err := kvc.List()
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	t.Logf("INFO %d collections found", len(kvceList))

}

func TestKVStoreCreate(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	kvc := ss.GetKVStore()
	collectionName := "test-collection-" + uuid.New().String()[0:8]
	t.Logf("INFO Creating a new KVStore collection '%s'", collectionName)

	fields := make(map[string]string, 0)
	fields["f1-str"] = KVStoreFieldTypeString
	fields["f2-num"] = KVStoreFieldTypeNumber
	ns, _ := NewNamespace("nobody", "search", SplunkSharingApp)
	kvce, err := kvc.CreateKVStoreColl(ns, collectionName, fields, nil, true, false)

	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	if kvce.Content.Fields["f1-str"] != KVStoreFieldTypeString {
		t.Errorf("Wrong field type for '%s'. Expected:'%s', actual: '%s'", "f1-str", kvce.Content.Fields["f1-str"], KVStoreFieldTypeString)
	}

	t.Logf("INFO Deleting KVStore collection '%s'", collectionName)
	if err = kvce.Delete(ss); err != nil {
		t.Error(err)
	}

}

func TestKVStoreDataManagement(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	kvc := ss.GetKVStore()
	collectionName := "test-collection-" + uuid.New().String()[0:8]
	t.Logf("INFO Creating a new KVStore collection '%s'", collectionName)

	fields := make(map[string]string, 0)
	fields["f1-str"] = KVStoreFieldTypeString
	fields["f2-num"] = KVStoreFieldTypeNumber
	ns, _ := NewNamespace("nobody", "search", SplunkSharingApp)
	kvce, err := kvc.CreateKVStoreColl(ns, collectionName, fields, nil, true, false)

	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	if kvce.Content.Fields["f1-str"] != KVStoreFieldTypeString {
		t.Errorf("Wrong field type for '%s'. Expected:'%s', actual: '%s'", "f1-str", kvce.Content.Fields["f1-str"], KVStoreFieldTypeString)
	}
	data := make([]map[string]interface{}, 0)
	kvce.Query(ss, "", "", "", 0, 0, false, &data)

	d := `{"f1-str":"some value", "f2-num": 12345}`
	if key, err := kvce.Insert(ss, d); err != nil {
		t.Error(err)
	} else {
		t.Logf("INFO Inserted key:'%s'", key)
	}

	kvce.Query(ss, "", "", "", 0, 0, false, &data)
	if len(data) != 1 {
		t.Errorf("Collection '%s' is expected to only have 1 entry. Found:%d", collectionName, len(data))
		t.FailNow()
	}

	t.Logf("INFO Deleting KVStore collection '%s'", collectionName)
	if err = kvce.Delete(ss); err != nil {
		t.Error(err)
	}

}
