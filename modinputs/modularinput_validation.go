package modinputs

import (
	"fmt"
	"strings"

	"github.com/prigio/splunk-go-sdk/v2/errors"
)

// ValidationRule defines an enumeration of the available splunk-provided splunk argument validations
//
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments
type ValidationRule string

// ValidationFun is the signature of the function used to validate the parameters received from Splunk
// (only used if the mod input is configured to use external validation)
type ValidationFunc func(*ModularInput, Stanza) error

const (
	// Ad-hoc validation codes.
	// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments
	ValidationIsAvailTCPPort ValidationRule = "is_avail_tcp_port"
	ValidationIsAvailUDPPort ValidationRule = "is_avail_udp_port"
	ValidationIsNonNegInt    ValidationRule = "is_nonneg_int"
	ValidationIsBool         ValidationRule = "is_bool"
	ValidationIsPort         ValidationRule = "is_port"
	ValidationIsPosInt       ValidationRule = "is_pos_int"
)

// RegisterValidationFunc registers a function of type [ValidationFunc] in charge of analysing the parameters provided within a stanza as a whole and decide whether they are valid.
func (mi *ModularInput) RegisterValidationFunc(f ValidationFunc) {
	mi.useExternalValidation = true
	mi.validate = f
}

// SetParamBasicValidation configures a predefined validation for the selected parameter.
// Available validations are listed as modinputs.Validation*.
//
// This is used within the input "scheme" definition to tell how Splunk's UI should validate the given parameter.
//
// The developer provides the parameter name, and the necessary validation rule.
//
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func (mi *ModularInput) SetParamBasicValidation(paramName string, validationRule ValidationRule) error {
	if !(validationRule == ValidationIsBool || validationRule == ValidationIsPort || validationRule == ValidationIsPosInt || validationRule == ValidationIsNonNegInt || validationRule == ValidationIsAvailUDPPort || validationRule == ValidationIsAvailTCPPort) {
		return errors.NewErrInvalidParam("setParamBasicValidation["+paramName+"]", nil, "'validationRule' must be one of '', provided='%s'", validationRule)
	}

	p, err := mi.GetParam(paramName)
	if err != nil {
		return errors.NewErrInvalidParam("setParamBasicValidation["+paramName+"]", err, "")
	}

	p.SetCustomProperty("validation", fmt.Sprintf("%s('%s')", string(validationRule), paramName))
	return nil
}

// SetParamCustomValidation configures a custom validation for the selected parameter.
// This is used within the input "scheme" definition to tell how Splunk's UI should validate the given parameter.
//
// The developer provides the parameter name, the exact validation clause and the error message to be displayed to the end-user in case of failed validation.
//
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func (mi *ModularInput) SetParamCustomValidation(paramName, condition, errorMessage string) error {
	if condition == "" || errorMessage == "" {
		return errors.NewErrInvalidParam("setParamCustomValidation["+paramName+"]", nil, "'condition' and 'errorMessage' cannot be empty")
	}

	p, err := mi.GetParam(paramName)
	if err != nil {
		return errors.NewErrInvalidParam("setParamCustomValidation["+paramName+"]", err, "")
	}
	errorMessage = strings.ReplaceAll(errorMessage, `"`, "'")
	p.SetCustomProperty("validation", fmt.Sprintf(`validate(%s,"%s")`, condition, errorMessage))
	return nil
}

// SetParamRegexValidation configures a regex-based validation for the selected parameter.
// This is used within the input "scheme" definition to tell how Splunk's UI should validate the given parameter.
//
// The developer provides the exact parameter name, a textual PCRE regular expression and the error message to be displayed to the end-user in case of failed validation.
//
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func (mi *ModularInput) SetParamRegexValidation(paramName, regex, errorMessage string) error {
	if regex == "" || errorMessage == "" {
		return errors.NewErrInvalidParam("setParamRegexValidation["+paramName+"]", nil, "'regex' and 'errorMessage' cannot be empty")
	}

	p, err := mi.GetParam(paramName)
	if err != nil {
		return errors.NewErrInvalidParam("setParamRegexValidation["+paramName+"]", err, "")
	}
	// replace double quotes within the error message
	errorMessage = strings.ReplaceAll(errorMessage, `"`, `'`)
	// escape double quotes within the regex
	regex = strings.ReplaceAll(regex, `"`, `\"`)

	p.SetCustomProperty("validation", fmt.Sprintf(`validate(match('%s',"%s"), "%s")`, paramName, regex, errorMessage))
	return nil
}
