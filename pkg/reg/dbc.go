package reg

import "regexp"

var (
	DBCVersion = regexp.MustCompile(`^(?:VERSION) *"(?P<version>.*)"$`)

	DBCBusSpeed = regexp.MustCompile(`^(?:BS_ *\:) *(?P<speed>\d+)?$`)

	DBCNode = regexp.MustCompile(`^(?:BU_ *\:)(?: *)(?P<nodes>.*)`)

	DBCMessage = regexp.MustCompile(`^(?:BO_) *(?P<msg_id>\d+) *(?P<msg_name>\w+) *: *(?P<length>\d+) *(?P<sender>\w+)$`)

	DBCSignal = regexp.MustCompile(
		`^(?:(?:\t| *)SG_) *(?P<sig_name>\w+) *(?P<mux_switch>m\d+)?(?P<mux>M)?(?: *): (?P<start_bit>\d+)\|(?P<size>\d+)@(?P<order>0|1)(?P<signed>\+|\-) *\((?P<scale>.*),(?P<offset>.*)\) *\[(?P<min>.*)\|(?P<max>.*)\] *"(?P<unit>.*)" *(?P<receivers>.*)`,
	)
	DBCExtMuxSignal = regexp.MustCompile(`^(?:SG_MUL_VAL_) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<mux_sig_name>\w+) *(?:\d+\-?){2} *;$`)

	DBCBitmapDef = regexp.MustCompile(`^(?:VAL_) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<bitmap>.*);$`)

	DBCNodeComment    = regexp.MustCompile(`^(?:CM_) *(?:BU_) *(?P<node_name>\w+) *"(?P<desc>.*)" *;$`)
	DBCMessageComment = regexp.MustCompile(`^(?:CM_) *(?:BO_) *(?P<msg_id>\d+) *"(?P<desc>.*)" *;$`)
	DBCSignalComment  = regexp.MustCompile(`^(?:CM_) *(?:SG_) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *"(?P<desc>.*)" *;$`)

	DBCAttribute = regexp.MustCompile(`^(?:BA_DEF_) *(?P<att_kind>SG_|BU_|BO_)? *"(?P<att_name>\w+)" *(?P<att_type>INT|FLOAT|STRING|ENUM) *(?P<att_data>.*[^;]) *;`)

	DBCAttributeDefault = regexp.MustCompile(`^(?:BA_DEF_DEF_) *"(?P<att_name>\w+)" *(?P<att_data>.*[^;]) *;`)

	DBCNodeAttributeAssignment    = regexp.MustCompile(`^(?:BA_) *"(?P<att_name>\w+)" *(?:BU_) *(?P<node_name>\w+) *(?P<att_data>.*[^;]) *;`)
	DBCMessageAttributeAssignment = regexp.MustCompile(`^(?:BA_) *"(?P<att_name>\w+)" *(?:BO_) *(?P<msg_id>\d+) *(?P<att_data>.*[^;]) *;`)
	DBCSignalAttributeAssignment  = regexp.MustCompile(`^(?:BA_) *"(?P<att_name>\w+)" *(?:SG_) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<att_data>.*[^;]) *;`)
)
