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

// ModInputScheme contains the basic configurations needed for Splunk to know how to use the ModularInput
type ModInputScheme struct {
	StanzaName            string // Name appearing within inputs.conf stanza as [stanzaname://...]. Must be lowercase
	Title                 string // Title displayed on the UI
	Description           string // Description within the UI
	UseExternalValidation bool
	UseSingleInstance     bool
	streamingMode         string
	Args                  []ModInputArg
}

func (mis *ModInputScheme) AddArgument(arg *ModInputArg) {
	mis.Args = append(mis.Args, *arg)
}

// PrintXMLScheme  is an exported function which outputs the scheme using the Splunk-defined XML structure.
func (mis *ModInputScheme) PrintXMLScheme() ([]byte, error) {
	// using the tecnique described at https://riptutorial.com/go/example/14194/marshaling-structs-with-private-fields//
	// in order to output streaming_mode, which otherwise would have to be publicly exported, which is unwanted.
	return xml.MarshalIndent(struct {
		XMLName               xml.Name `xml:"scheme"`
		Title                 string   `xml:"title"`
		Description           string   `xml:"description"`
		UseExternalValidation bool     `xml:"use_external_validation"`
		UseSingleInstance     bool     `xml:"use_single_instance"`
		//Adding a fixed StreamingMode, not present within the original structure
		StreamingMode string        `xml:"streaming_mode"`
		Args          []ModInputArg `xml:"endpoint>args"`
	}{
		Title:                 mis.Title,
		Description:           mis.Description,
		UseExternalValidation: mis.UseExternalValidation,
		UseSingleInstance:     mis.UseSingleInstance,
		//Adding a fixed StreamingMode
		StreamingMode: "xml",
		Args:          mis.Args,
	}, "", "  ")
	//mis.streamingMode = "xml"
}

// ExampleConf returns a string containing a sample configuration
// for the modular input based on its definition
// this can help an user test a configuration within splunk's inputs.conf
func (mis *ModInputScheme) ExampleConf() string {
	var sb strings.Builder
	fmt.Fprint(sb, "# Example configs for inputs.conf\n")
	fmt.Fprintf(sb, "# %s\n", mis.Description)
	fmt.Fprintf(sb, "[%s://name-of-input]\n", mis.StanzaName)
	for _, arg := range mis.Args {
		fmt.Fprintf(sb, "# %s - %s\n", arg.title, arg.Description)
		fmt.Fprintf(sb, "%s = <%s>\n", arg.Name, arg.DataType)
	}
	fmt.Fprint(sb, "# Standard input configurations\n")
	fmt.Fprint(sb, "index = <index>\n")
	fmt.Fprint(sb, "sourcetype = <sourcetype>\n")
	fmt.Fprint(sb, "interval = <cron schedule>\n")
	return sb.String()
}

// Parameters used by the ModularInput.
type ModInputArg struct {
	XMLName          xml.Name `xml:"arg"`
	Name             string   `xml:"name,attr"`
	Title            string   `xml:"title"`
	Description      string   `xml:"description,omitempty"`
	DataType         string   `xml:"data_type,omitempty"`
	RequiredOnCreate bool     `xml:"required_on_create"`
	RequiredOnEdit   bool     `xml:"required_on_edit"`
	// validation should be at best be configured through methods
	Validation string `xml:"validation,omitempty"`
}

func (mia *ModInputArg) SetValidation(validationRule ArgValidation) {
	if validationRule == ArgValidationIsBool || validationRule == ArgValidationIsPort || validationRule == ArgValidationIsPosInt || validationRule == ArgValidationIsNonNegInt || validationRule == ArgValidationIsAvailUDPPort || validationRule == ArgValidationIsAvailTCPPort {
		mia.Validation = fmt.Sprintf("%s('%s')", string(validationRule), mia.Name)
	}
}

func (mia *ModInputArg) SetCustomValidation(condition string, errorMessage string) {
	if condition != "" {
		// write validate(<CustomValidation>, <CustomValidationErrMessage>) to buffer
		mia.Validation = fmt.Sprintf("validate(%s,\"%s\")", condition, strings.ReplaceAll(errorMessage, `"`, "'"))
	}
}
