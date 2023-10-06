package splunkd

import (
	"testing"
)

func TestContext(t *testing.T) {
	ss, err := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err = ss.Login(testing_user, testing_password, testing_mfaCode); err != nil {
		t.Error(err)
		t.FailNow()
	}

	cr, err := ss.AuthContext()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if cr.Username != testing_user {
		t.Errorf("Invalid Context value provided. %+v", cr)
	}
}
