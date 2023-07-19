package splunkd

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func TestCredentialNoRealm(t *testing.T) {
	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	credentials := ss.GetCredentials()
	credNoRealm := uuid.New().String()[0:8] + "-no-realm"
	credPassword := "this is a password"

	t.Logf("INFO Creating credential realm='' user='%s'", credNoRealm)
	cr, err := credentials.CreateCred(credNoRealm, "", credPassword)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if cr.Content.Username != credNoRealm {
		t.Errorf("Invalid credential returned. %+v", cr)
	}

	// retrieve the new credential
	t.Logf("INFO Retrieving credential realm='' user='%s'", credNoRealm)
	cr2, err := credentials.GetCred(credNoRealm, "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if cr2.Content.Username != credNoRealm || cr2.Content.ClearPassword != credPassword {
		t.Errorf("Invalid credential returned. %+v", cr2)
	}
	// update password of new credential
	t.Logf("INFO Updating password for user='%s'", credNoRealm)
	updatedCredPassword := "this is another password"
	if err = credentials.UpdateCred(credNoRealm, "", updatedCredPassword); err != nil {
		t.Errorf("Invalid password returned. %s", err.Error())
	}

	// Retrieve the updated credential and check its password
	cr3, err := credentials.GetCred(credNoRealm, "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if cr3.Content.ClearPassword != updatedCredPassword {
		t.Errorf("Password not updated. %+v", cr3)
	}

	t.Logf("INFO Deleting credential realm='' user='%s'", credNoRealm)
	// Delete the credential created for the test
	err = credentials.Delete(urlEncodeCredential(credNoRealm, ""))
	if err != nil {
		t.Errorf("Deletion of credential failed: %s", err.Error())
	}
}

func TestCredentialWithRealm(t *testing.T) {
	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	credentials := ss.GetCredentials()
	credWithRealm := uuid.New().String()[0:8] + "-with-realm"
	credRealm := "testRealm"
	credPassword := "this is a password"

	t.Logf("INFO Creating credential realm='%s' user='%s'", credRealm, credWithRealm)
	cr, err := credentials.CreateCred(credWithRealm, credRealm, credPassword)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if cr.Content.Username != credWithRealm {
		t.Errorf("Invalid credential returned. %+v", cr)
	}

	// retrieve the new credential
	t.Logf("INFO Retrieving credential realm='%s' user='%s'", credRealm, credWithRealm)
	cr2, err := credentials.GetCred(credWithRealm, credRealm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if cr2.Content.Username != credWithRealm || cr2.Content.ClearPassword != credPassword {
		t.Errorf("Invalid credential returned. %+v", cr2)
	}
	// update password of new credential
	t.Logf("INFO Updating password for user='%s'", credWithRealm)
	updatedCredPassword := "this is another password"
	if err = credentials.UpdateCred(credWithRealm, credRealm, updatedCredPassword); err != nil {
		t.Errorf("Invalid password returned. %s", err.Error())
	}

	// Retrieve the updated credential and check its password
	cr3, err := credentials.GetCred(credWithRealm, credRealm)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if cr3.Content.ClearPassword != updatedCredPassword {
		t.Errorf("Password not updated. %+v", cr3)
	}

	t.Logf("INFO Deleting credential realm='%s' user='%s'", credRealm, credWithRealm)
	// Delete the credential created for the test
	err = credentials.Delete(urlEncodeCredential(credWithRealm, credRealm))
	if err != nil {
		t.Errorf("Deletion of credential failed: %s", err.Error())
	}
}

func TestCredentialACL(t *testing.T) {
	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	credentials := ss.GetCredentials()
	credWithRealm := uuid.New().String()[0:8] + "-with-realm"
	credRealm := "testRealm"
	credPassword := "this is a password"

	t.Logf("INFO Creating credential realm='%s' user='%s'", credRealm, credWithRealm)
	cr, err := credentials.CreateCred(credWithRealm, credRealm, credPassword)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if cr.Content.Username != credWithRealm {
		t.Errorf("Invalid credential returned. %+v", cr)
	}
	if cr.ACL.Owner != testing_user {
		t.Errorf("Invalid ACL returned. %+v", cr)
	}

	// retrieve the new credential
	t.Logf("INFO Changing owner for credential realm='%s' user='%s'", credRealm, credWithRealm)

	params := url.Values{}
	params.Set("owner", "test")
	params.Set("perms.read", "power")
	params.Add("perms.read", "user")

	params.Set("perms.write", "admin")
	params.Add("perms.write", "user")

	err = credentials.UpdateCredACL(credWithRealm, credRealm, &params)
	if err != nil {
		t.Error(err)
	}

	t.Logf("INFO Verifying owner for credential realm='%s' user='%s'", credRealm, credWithRealm)
	cr, err = credentials.GetCred(credWithRealm, credRealm)
	if err != nil {
		t.Error(err)
	}

	if cr.ACL.Owner != "test" {
		t.Errorf("Changing of owner did not work for credential realm='%s' user='%s'. Expected: '%s', found: '%s'", credRealm, credWithRealm, params.Get("owner"), cr.ACL.Owner)
		// do not delete the credential in case testing failed
		t.FailNow()
	}
	permReadFound := 0
	for _, pr := range cr.ACL.Perms.Read {
		if pr == "power" || pr == "user" {
			permReadFound += 1
		}
	}
	if permReadFound < 2 {
		t.Errorf("Changing perms.read did not work for credential realm='%s' user='%s'. Expected: '%s', found: '%v'", credRealm, credWithRealm, params.Get("perms.read"), cr.ACL.Perms.Read)
		t.FailNow()
	}

	t.Logf("INFO Deleting credential realm='%s' user='%s'", credRealm, credWithRealm)
	// Delete the credential created for the test
	err = credentials.Delete(urlEncodeCredential(credWithRealm, credRealm))
	if err != nil {
		t.Errorf("Deletion of credential failed: %s", err.Error())
	}
}
