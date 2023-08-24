package modinputs

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// ArgValidation defines an enumeration of the available splunk-provided splunk argument evaluations
type ArgValidation string

const (
	// Ad-hoc validation codes.
	// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments
	ArgValidationIsAvailTCPPort ArgValidation = "is_avail_tcp_port"
	ArgValidationIsAvailUDPPort ArgValidation = "is_avail_udp_port"
	ArgValidationIsNonNegInt    ArgValidation = "is_nonneg_int"
	ArgValidationIsBool         ArgValidation = "is_bool"
	ArgValidationIsPort         ArgValidation = "is_port"
	ArgValidationIsPosInt       ArgValidation = "is_pos_int"
	ArgDataTypeStr                            = "string"
	ArgDataTypeBool                           = "boolean"
	ArgDataTypeNumber                         = "number"
)

// GenerateArgValidation returns a string which can be used within a modular input "scheme" definition to tell how Splunk's UI should validate the given parameter
// Available validations are listed as modinputs.ArgValidation*
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func GenerateArgValidation(paramName string, validation ArgValidation) string {
	return fmt.Sprintf("%s(%s)", string(validation), paramName)
}

// GenerateArgValidationComplex returns a string which can be used within a modular input "scheme" definition to tell how Splunk's UI should validate the given parameter
// The user provides the exact validatioin clause to be used and the error message to be displayed to the user in case of failed validation
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func GenerateArgValidationComplex(paramName, checkClause, errorMessage string) string {
	return fmt.Sprintf("validate(%s, \"%s\")", checkClause, strings.ReplaceAll(errorMessage, "\"", "'"))
}

// GenerateArgValidationRegex returns a string which can be used within a modular input "scheme" definition to tell how Splunk's UI should validate the given parameter
// The user provides the exact parameter name, a textual PCRE regular expression and the error message to be displayed to the user in case of failed validation
// See https://docs.splunk.com/Documentation/SplunkCloud/latest/AdvancedDev/ModInputsScripts#Validation_of_arguments for more info
func GenerateArgValidationRegex(paramName, regex, errorMessage string) string {
	return fmt.Sprintf("validate(match('%s',\"%s\"), \"%s\")", paramName, strings.ReplaceAll(regex, "\"", "\\\""), strings.ReplaceAll(errorMessage, "\"", "'"))
}

// Parameters used by the ModularInput.
type InputArg struct {
	XMLName xml.Name `xml:"arg"`
	// Title is the visible name of the parameter, used within the UI
	Title string `xml:"title"`
	// Name is the internal name of the parameter, the one actually provided within splunk configurations
	Name        string `xml:"name,attr"`
	Description string `xml:"description,omitempty"`
	DataType    string `xml:"data_type,omitempty"`
	// DefaultValue is used in case an actual value has not been set by the run-time configurations
	DefaultValue string `xml:"-"` // this is omitted in the XML format, since this is not foreseen by Splunk.
	// validation should be at best be configured through methods
	Validation       string `xml:"validation,omitempty"`
	RequiredOnCreate bool   `xml:"required_on_create"`
	RequiredOnEdit   bool   `xml:"required_on_edit"`
}

func (mia *InputArg) SetValidation(validationRule ArgValidation) {
	if validationRule == ArgValidationIsBool || validationRule == ArgValidationIsPort || validationRule == ArgValidationIsPosInt || validationRule == ArgValidationIsNonNegInt || validationRule == ArgValidationIsAvailUDPPort || validationRule == ArgValidationIsAvailTCPPort {
		mia.Validation = fmt.Sprintf("%s('%s')", string(validationRule), mia.Name)
	}
}

func (mia *InputArg) SetCustomValidation(condition string, errorMessage string) {
	if condition != "" {
		// write validate(<CustomValidation>, <CustomValidationErrMessage>) to buffer
		mia.Validation = fmt.Sprintf("validate(%s,\"%s\")", condition, strings.ReplaceAll(errorMessage, `"`, "'"))
	}
}

// getInputsSpec returns a string which can be used to describe the parameter within splunk's README/inputs.conf.spec file
func (mia *InputArg) getInputsSpec() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `%s = <%s>
*  %s: %s
*  Default value: "%s"
`, mia.Name, mia.DataType, mia.Title, strings.ReplaceAll(mia.Description, "\n", " "), strings.ReplaceAll(mia.DefaultValue, "\n", " "))

	if len(mia.Validation) > 0 {
		fmt.Fprintf(buf, "* Custom validation: %s\n", mia.Validation)
	}
	return buf.String()
}

// getInputsConf returns a string which can be used to describe the parameter within splunk's default/inputs.conf file
func (mia *InputArg) getInputsConf() string {
	buf := new(strings.Builder)
	// pre-growing the buffer to 512 bytes: this avoids doing this continuously when executing buf.WriteString()
	buf.Grow(512)

	fmt.Fprintf(buf, `#  %s: %s
#  Data type: %s
#  Default value: "%s"
`, mia.Title, strings.ReplaceAll(mia.Description, "\n", " "), mia.DataType, mia.DefaultValue)

	if len(mia.Validation) > 0 {
		fmt.Fprintf(buf, "# Custom validation: %s\n", mia.Validation)
	}

	fmt.Fprintf(buf, "%s = %s\n", mia.Name, strings.ReplaceAll(mia.DefaultValue, "\n", "\\\n"))

	return buf.String()
}

// GenerateDocumentation returns a markdown-formatted list-item which describes the parameter
func (mia *InputArg) GenerateDocumentation() string {
	buf := new(strings.Builder)

	fmt.Fprintf(buf, "- `%s` : %s - ", mia.Name, mia.Title)

	if mia.RequiredOnCreate || mia.RequiredOnEdit {
		fmt.Fprintf(buf, "(%s, required) ", mia.DataType)
	} else {
		fmt.Fprintf(buf, "(%s) ", mia.DataType)
	}

	fmt.Fprint(buf, mia.Description)

	if mia.DefaultValue != "" {
		fmt.Fprintf(buf, "    Default value: `%s`", mia.DefaultValue)
	}

	if mia.Validation != "" {
		fmt.Fprintf(buf, "    Validation: `%s`", mia.Validation)
	}

	return buf.String()
}
