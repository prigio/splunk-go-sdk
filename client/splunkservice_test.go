package client

import (
	"testing"
)

const (
	host               = "host.docker.internal"
	user               = "admin"
	password           = "splunked"
	mfaCode            = ""
	port               = 38089
	insecureSkipVerify = true
)

func TestLogin(t *testing.T) {
	_, err := NewSplunkServiceWithUsernameAndPassword(host, port, user, password, mfaCode, insecureSkipVerify)
	if err != nil {
		t.Error(err)
	}
}

func TestInfo(t *testing.T) {
	if ss, err := NewSplunkServiceWithUsernameAndPassword(host, port, user, password, mfaCode, insecureSkipVerify); err != nil {
		t.Error(err)
	} else {
		ir, err := ss.Info()
		if err != nil {
			t.Error(err)
		}
		if ir.Build == "" {
			t.Errorf("Invalid Info value provided. %+v", ir)
		}
	}
}
