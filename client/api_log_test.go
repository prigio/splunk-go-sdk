package client

import (
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	logger := ss.NewLogger("testLogger", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix, "main", "", "go-test", "go-test")
	t.Logf("Writing test messages into index=%s, sourcetype=%s\n", "main", "go-test")

	logger.Println("test log message sent using log.Println")
}

/*
func TestLoggerInternal(t *testing.T) {
	if ss, err = New(testing_endpoint, testing_insecureSkipVerify, testing_proxy); err != nil {
		t.Error(err)
		t.FailNow()
	}
	ss.Login(testing_user, testing_password, testing_mfaCode)

	logger := ss.NewLogger("testLogger", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix, "main", "", "go-test", "go-test")
	t.Logf("Writing test messages into index=%s, sourcetype=%s\n", "main", "go-test")


		msg := "test log message sent to splunkdLogger.Write"
		b := []byte(msg)
		n, err := splunkdLogger.Write(b)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if n != len(b) {
			t.Errorf("Write() method did not return corrent number of written bytes")
		}

	logger.Println("test log message sent using log.Println")
}
*/
