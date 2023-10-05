package splunkd

import (
	"testing"

	"github.com/prigio/splunk-go-sdk/v2/utils"
)

func TestCollectionList(t *testing.T) {
	ss := mustLoginToSplunk(t)

	propsCol := NewConfigsCollection(ss, "props")

	allProps, err := propsCol.List()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	propsNames := utils.ListOfVals(allProps, func(v *entry[ConfigResource]) string { return v.Name })
	if len(allProps) < 200 {
		t.Errorf("collection.List did not return all configurations for props: %v", propsNames)
	}
}

func TestCollectionSearch(t *testing.T) {
	ss := mustLoginToSplunk(t)
	filter := "name=\"splunk_*\""
	propsCol := NewConfigsCollection(ss, "props")

	allProps, err := propsCol.Search(filter)
	t.Logf("INFO Search of config-props for '%s' returned %d values", filter, len(allProps))

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(allProps) == 0 || len(allProps) > 50 {
		propsNames := utils.ListOfVals(allProps, func(v *entry[ConfigResource]) string { return v.Name })
		t.Errorf("collection.List did not return all configurations for props: %v", propsNames)
	}
}
