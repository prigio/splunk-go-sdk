package modinputs

/*
This file contains ModularInput functions to register and deal with parameters.
*/

import (
	"fmt"

	"github.com/prigio/splunk-go-sdk/v2/errors"
	"github.com/prigio/splunk-go-sdk/v2/params"
)

// GetParam searches for the param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (mi *ModularInput) GetParam(name string) (*params.Param, error) {
	for _, p := range mi.params {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("getParam[%s]: not found", name)
}

// GetGlobalParam searches for the global param having the provided name.
// Returns a pointer to the found parameter, or an error if the parameter was not found
func (aa *ModularInput) GetGlobalParam(name string) (*params.Param, error) {
	for _, p := range aa.globalParams {
		if p.GetName() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf(`getGlobalParam[%s]: parameter not found`, name)
}

// RegisterNewParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
//
// The following are the only adminissible values for the dataType. Anything else will generate an error.
// - "string"
// - "boolean"
// - "number"
func (mi *ModularInput) RegisterNewParam(name, title, description, defaultValue, dataType string, required, sensitive bool) (*params.Param, error) {
	var (
		p   *params.Param
		err error
	)
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = mi.GetParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam("inputs.conf", mi.StanzaName, name, title, description, defaultValue, required, sensitive)
	if err != nil {
		return nil, fmt.Errorf("registerNewParam[%s]: %w", name, err)
	}
	err = p.SetDataType(dataType)
	if err != nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", err, `'dataType' provided="%s"`, dataType)
	}

	if mi.params == nil {
		mi.params = make([]*params.Param, 0, 1)
	}
	mi.params = append(mi.params, p)
	return p, nil
}

// RegisterNewGlobalParam adds a new parameter to the alert action.
// The argument is additionally returned for further processing, if needed.
func (mi *ModularInput) RegisterNewGlobalParam(configFile, stanza, name, title, description, defaultValue, dataType string, required, sensitive bool) (*params.Param, error) {
	var p *params.Param
	var err error
	// check if the parameter is already present
	// return error in case it is already there
	if _, err = mi.GetGlobalParam(name); err == nil {
		return nil, errors.NewErrInvalidParam("registerNewGlobalParam["+name+"]", nil, "'%s' already exists", name)
	}
	p, err = params.NewParam(configFile, stanza, name, title, description, defaultValue, required, sensitive)
	if err != nil {
		return nil, fmt.Errorf("registerNewGlobalParam[%s]: %w", name, err)
	}
	err = p.SetDataType(dataType)
	if err != nil {
		return nil, errors.NewErrInvalidParam("registerNewParam["+name+"]", err, `'dataType' provided="%s"`, dataType)
	}

	if mi.globalParams == nil {
		mi.globalParams = make([]*params.Param, 0, 1)
	}
	mi.globalParams = append(mi.globalParams, p)
	return p, nil
}
