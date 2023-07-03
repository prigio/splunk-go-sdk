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
	// See https://docs.splunk.com/Documentation/Splunk/8.1.1/AdvancedDev/ModInputsScripts#Built-in_arguments_and_actions
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

// Parameters used by the ModularInput.
type InputArg struct {
	XMLName      xml.Name `xml:"arg"`
	Title        string   `xml:"title"`
	Description  string   `xml:"description,omitempty"`
	Name         string   `xml:"name,attr"`
	DataType     string   `xml:"data_type,omitempty"`
	DefaultValue string   `xml:"-"` // this is omitted in the XML format, since this is not foreseen by Splunk.
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
		fmt.Fprintf(buf, "* Custom validation: %s", mia.Validation)
	}
	return buf.String()
}

// getInputsConf returns a string which can be used to describe the parameter within splunk's defaault/inputs.conf file
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
