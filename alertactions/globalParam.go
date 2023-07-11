package alertactions

/*
import (
	"fmt"

	"github.com/prigio/splunk-go-sdk/client"
)

// Parameters used by the ModularInput.
type GlobalParam struct {
	ConfigFile string
	Stanza     string
	Param
}

func (gp *GlobalParam) GetValue(ss *client.SplunkService) (string, error) {
	if ss == nil {
		return "", fmt.Errorf("globalParam GetValue: reference to splunk service cannot be nil")
	}
	col := ss.GetConfigs(gp.ConfigFile)
	if stanza, err := col.GetStanza(gp.Stanza); err != nil {
		return "", fmt.Errorf("globalParam GetValue: %w", err)
	} else {
		return stanza.GetString(gp.Name)
	}
}
*/
