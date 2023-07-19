package splunkd

import (
	"testing"
)

func TestInfo(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	t.Log("INFO Retrieving data from INFO endpoint")
	ir, err := ss.Info()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if ir.Build == "" {
		t.Errorf("Invalid Info value provided. %+v", ir)
	}
}
