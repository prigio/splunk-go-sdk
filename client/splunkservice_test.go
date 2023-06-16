package client

import (
	"testing"
)

var ss *SplunkService
var err error

const (
	endpoint           = "https://localhost:2089"
	user               = "admin"
	password           = "splunked"
	jwtToken           = "eyJraWQiOiJzcGx1bmsuc2VjcmV0IiwiYWxnIjoiSFM1MTIiLCJ2ZXIiOiJ2MiIsInR0eXAiOiJzdGF0aWMifQ.eyJpc3MiOiJhZG1pbiBmcm9tIHNwbHVuazktZG9ja2VyIiwic3ViIjoiYWRtaW4iLCJhdWQiOiJ0ZXN0IiwiaWRwIjoiU3BsdW5rIiwianRpIjoiYjI2NDQ3NWI4OGEzMDE5YjYwMjA1ODM2MWIyNGQxZGU1YmQ2ZDExM2Q1ZWIwZGIyNzE3MzU4NmJkZjZmNDQyOCIsImlhdCI6MTY4Njg3MjA1MCwiZXhwIjoxNjg5NDY0MDUwLCJuYnIiOjE2ODY4NzIwNTB9.cjUT0lwfOqLlCHlNYUGydoqckSxFZFYZKYZq6jfRx5Rr8Kvzd5PNd7XyY3ocGx2f1YHJOVSmjgMRnn56gESzPA"
	mfaCode            = ""
	insecureSkipVerify = true
)

func TestLogin(t *testing.T) {
	_, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, password, mfaCode, insecureSkipVerify)
	if err != nil {
		t.Error(err)
	}

	_, err = NewSplunkServiceWithUsernameAndPassword("https://thisdoesnotexist:443", user, password, mfaCode, insecureSkipVerify)
	if err == nil {
		t.Error("Client pretended to establish connection to a non existing service")
	}

	_, err = NewSplunkServiceWithUsernameAndPassword(endpoint, "missing-user", password, mfaCode, insecureSkipVerify)
	if err == nil {
		t.Error("Client pretended to establish connection using a non existing user")
	}

	_, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, "this is the wrong password", mfaCode, insecureSkipVerify)
	if err == nil {
		t.Error("Client pretended to establish connection using a wrong password")
	}
}

func TestLoginSessionKey(t *testing.T) {

	ss, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, password, mfaCode, insecureSkipVerify)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = NewSplunkServiceWithSessionKey(endpoint, ss.sessionKey, insecureSkipVerify)
	if err != nil {
		t.Error(err)
	}

	_, err = NewSplunkServiceWithSessionKey(endpoint, "wrong session key", insecureSkipVerify)
	if err == nil {
		t.Error("Client pretended to establish connection using wrong session key")
	}

}

func TestLoginToken(t *testing.T) {

	_, err = NewSplunkServiceWithToken(endpoint, jwtToken, insecureSkipVerify)
	if err != nil {
		t.Error(err)
	}

	_, err = NewSplunkServiceWithToken(endpoint, "wrong token", insecureSkipVerify)
	if err == nil {
		t.Error("Client pretended to establish connection using wrong token")
	}

}

func TestInfo(t *testing.T) {
	if ss, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, password, mfaCode, insecureSkipVerify); err != nil {
		t.Error(err)
		t.FailNow()
	}

	ir, err := ss.Info()
	if err != nil {
		t.Error(err)
	}
	if ir.Build == "" {
		t.Errorf("Invalid Info value provided. %+v", ir)
	}

}

func TestCredential(t *testing.T) {
	if ss, err = NewSplunkServiceWithUsernameAndPassword(endpoint, user, password, mfaCode, insecureSkipVerify); err != nil {
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
