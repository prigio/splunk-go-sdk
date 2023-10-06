package splunkd

import (
	"testing"
)

func TestLogin(t *testing.T) {
	wrongUrl := "http://thisdoesnotexist.nonexistingdomain.com"
	_, err := New(wrongUrl, testing_insecureSkipVerify, testing_proxy)
	if err == nil {
		t.Errorf("Client pretended to establish connection to a non existing service: '%s'", wrongUrl)
	}

	ss, err := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
		t.Error("Client could not establish connection to an existing service")
		t.FailNow()
	}

	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error("Client not perform login with username and password")
	}
	if err = ss.Login(testing_user, "some wrong password", ""); err == nil {
		t.Error("Client performed login with wrong password")
	}
}

func TestLoginSessionKey(t *testing.T) {
	ss, err := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
		t.Error("Client could not establish connection to an existing service")
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	sessKey := ss.GetSessionKey()

	clientSess, _ := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err = clientSess.LoginWithSessionKey(sessKey); err != nil {
		t.Errorf("Login using session key did not work. Session key: '%s'", sessKey)
	}

}

func TestLoginToken(t *testing.T) {
	ss, err := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if testing_jwtToken != "" {
		if err = ss.LoginWithToken(testing_jwtToken); err != nil {
			t.Error("Login using token did not work")
		}
	}

}

/*
func TestCredential(t *testing.T) {
	if ss, err = New(endpoint, insecureSkipVerify, proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}

	cred, acl, err := ss.getCredential("automation", "uxmone.web.porsche.de")
	if err != nil {
		t.Error(err)
	}
	if acl.App != "ta-infra-modinputs" {
		t.Error("Wrong ACL read-in")
	}
	if cred.EncrPassword == "" {
		t.Error("Missing password")
	}

}
*/
/*
func TestSetCredential(t *testing.T) {
	if ss, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, password, mfaCode, insecureSkipVerify); err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err := ss.setCredential("someusername", "somerealm", "somepassword")
	if err != nil {
		t.Error(err)
	}

}
*/
