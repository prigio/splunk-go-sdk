package splunkd

import (
	"os"
	"testing"
)

const (
	testing_endpoint           = "https://localhost:8089"
	testing_user               = "admin"
	testing_password           = "splunked"
	testing_jwtToken           = ""
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
	ss, err := New(testing_endpoint, testing_insecureSkipVerify, testing_proxy)
	if err != nil {
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
