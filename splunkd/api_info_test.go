package splunkd

import (
	"testing"
)

func TestInfo(t *testing.T) {
	ss := mustLoginToSplunk(t)

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
