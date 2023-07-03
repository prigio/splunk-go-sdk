package client

import (
	"testing"
)

var ss *SplunkService
var err error

const (
	testing_endpoint = "https://splunk:3089"
	testing_user     = "admin"
	testing_password = "splunked"
	testing_jwtToken = "eyJraWQiOiJzcGx1bmsuc2VjcmV0IiwiYWxnIjoiSFM1MTIiLCJ2ZXIiOiJ2MiIsInR0eXAiOiJzdGF0aWMifQ.eyJpc3MiOiJhZG1pbiBmcm9tIHNwbHVuay12ZXJ0ZWMiLCJzdWIiOiJhZG1pbiIsImF1ZCI6InRlc3RpbmciLCJpZHAiOiJTcGx1bmsiLCJqdGkiOiI2NmFiNDA1ZDNhYWFlNzQ2ZjgzMmRjZGM5YjYyMDAyYzcyMjcxMTQ1YzgzNmEyMmFiNzVhZGVhMGViMzMxZWZhIiwiaWF0IjoxNjg4MTMwMDE3LCJleHAiOjE2OTA3MjIwMTcsIm5iciI6MTY4ODEzMDAxN30.weVJJmkZGioDUZM0CicAbJ1mO1OZXUPIE8PBwZHmLh5lW0eFUN8mCpkg4BzVvHTJXLBXjRhSyMq_m5obwuKgvg"
	//jwtToken           = "eyJraWQiOiJzcGx1bmsuc2VjcmV0IiwiYWxnIjoiSFM1MTIiLCJ2ZXIiOiJ2MiIsInR0eXAiOiJzdGF0aWMifQ.eyJpc3MiOiJhZG1pbiBmcm9tIHNwbHVuazktZG9ja2VyIiwic3ViIjoiYWRtaW4iLCJhdWQiOiJ0ZXN0IiwiaWRwIjoiU3BsdW5rIiwianRpIjoiYjI2NDQ3NWI4OGEzMDE5YjYwMjA1ODM2MWIyNGQxZGU1YmQ2ZDExM2Q1ZWIwZGIyNzE3MzU4NmJkZjZmNDQyOCIsImlhdCI6MTY4Njg3MjA1MCwiZXhwIjoxNjg5NDY0MDUwLCJuYnIiOjE2ODY4NzIwNTB9.cjUT0lwfOqLlCHlNYUGydoqckSxFZFYZKYZq6jfRx5Rr8Kvzd5PNd7XyY3ocGx2f1YHJOVSmjgMRnn56gESzPA"
	testing_mfaCode            = ""
	testing_insecureSkipVerify = true
	testing_proxy              = "" //"http://localhost:8080"
)

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

	if err = ss.LoginWithToken(testing_jwtToken); err != nil {
		t.Error("Login using token did not work")
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
