package alertactions

// Parameters used by the ModularInput.
type GlobalParam struct {
	ConfigFile string
	Stanza     string
	Param
}

/*
func (gp *GlobalParam) GetValue(ss *client.SplunkService) (string, error) {
	if gp.actualValueIsSet {
		return gp.actualValue, nil
	}
	if ss == nil {
		return "", fmt.Errorf("globalParam GetValue: reference to splunk service cannot be nil")
	}
	col := ss.GetConfigs(gp.ConfigFile)
	stanza, err := col.GetStanza(gp.Stanza)
	if err != nil {
		return "", fmt.Errorf("globalParam GetValue: %w", err)
	}
	v, err := stanza.GetString(gp.Name)
	if err != nil {
		return "", fmt.Errorf("globalParam GetValue: %w", err)
	}
	gp.actualValue = v
	gp.actualValueIsSet = true
	return v, nil
}
*/
