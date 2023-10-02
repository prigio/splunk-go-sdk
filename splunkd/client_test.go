package splunkd

import (
	"os"
	"testing"
)

var ss *Client
var err error

const (
	testing_endpoint = "https://splunk:2089"
	testing_user     = "admin"
	testing_password = "splunked"
	testing_jwtToken = ""
	//jwtToken           = "eyJraWQiOiJzcGx1bmsuc2VjcmV0IiwiYWxnIjoiSFM1MTIiLCJ2ZXIiOiJ2MiIsInR0eXAiOiJzdGF0aWMifQ.eyJpc3MiOiJhZG1pbiBmcm9tIHNwbHVuazktZG9ja2VyIiwic3ViIjoiYWRtaW4iLCJhdWQiOiJ0ZXN0IiwiaWRwIjoiU3BsdW5rIiwianRpIjoiYjI2NDQ3NWI4OGEzMDE5YjYwMjA1ODM2MWIyNGQxZGU1YmQ2ZDExM2Q1ZWIwZGIyNzE3MzU4NmJkZjZmNDQyOCIsImlhdCI6MTY4Njg3MjA1MCwiZXhwIjoxNjg5NDY0MDUwLCJuYnIiOjE2ODY4NzIwNTB9.cjUT0lwfOqLlCHlNYUGydoqckSxFZFYZKYZq6jfRx5Rr8Kvzd5PNd7XyY3ocGx2f1YHJOVSmjgMRnn56gESzPA"
	testing_mfaCode            = ""
	testing_insecureSkipVerify = true
	testing_proxy              = "" //"http://localhost:8080"
)

// TestMain can be used for global initialization and tear-down of the testing environment
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	/*ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
		fmt.Printf("TestMain: %s", err)
		os.Exit(1)
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		fmt.Printf("TestMain: %s", err)
		os.Exit(1)
	}
	*/
	os.Exit(m.Run())
}

// mustLoginToSplunk performs a username+password login on splunk and returns a *Client instance to it.
// If login is not successful, the test t is interrrupted.
func mustLoginToSplunk(t *testing.T) *Client {
	t.Log("INFO Connecting to Splunk")
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error("Client can not perform login with username and password")
		t.FailNow()
		return nil
	}
	return ss
}

func TestLogin(t *testing.T) {
	wrongUrl := "http://thisdoesnotexist.nonexistingdomain.com"
	_, err = New(wrongUrl, testing_insecureSkipVerify, testing_proxy)
	if err == nil {
		t.Errorf("Client pretended to establish connection to a non existing service: '%s'", wrongUrl)
	}

	ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
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
	ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
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
	ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
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
