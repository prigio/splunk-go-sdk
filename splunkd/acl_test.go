package splunkd

import (
	"testing"

	"github.com/google/go-querystring/query"
)

func TestACLToUrl(t *testing.T) {
	a := AccessControlList{
		Owner: "test",
		App:   "search",
	}
	a.Perms.Read = []string{"user", "power", "admin"}
	a.Perms.Write = []string{"admin", "power"}

	vals, err := query.Values(&a)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf(vals.Encode())
}
