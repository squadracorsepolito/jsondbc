package dbc

type keywordKind uint

const (
	keywordVersion keywordKind = iota

	keywordNewSimbol

	keywordBitTiming

	keywordNode

	keywordMessage
	keywordMessageTx

	keywordSignal

	keywordValueTable
	keywordValueDesc

	keywordEnvVar
	keywordEnvVarData

	keywordSignalType
	keywordSignalGroup

	keywordComment

	keywordAttribute
	keywordAttributeValue
	keywordAttributeInt
	keywordAttributeHex
	keywordAttributeFloat
	keywordAttributeString
	keywordAttributeEnum

	keywordExtendedMuxSignal
)

var keywords = map[string]keywordKind{
	"VERSION": keywordVersion,

	"NS_": keywordNewSimbol,

	"BS_": keywordBitTiming,

	"BU_": keywordNode,

	"BO_":       keywordMessage,
	"BO_TX_BU_": keywordMessageTx,

	"SG_": keywordSignal,

	"VAL_TABLE_": keywordValueTable,
	"VAL_":       keywordValueDesc,

	"EV_":          keywordEnvVar,
	"ENVVAR_DATA_": keywordEnvVarData,

	"SGTYPE_":    keywordSignalType,
	"SIG_GROUP_": keywordSignalGroup,

	"CM_": keywordComment,

	"BA_DEF_": keywordAttribute,
	"BA_":     keywordAttributeValue,
	"INT":     keywordAttributeInt,
	"HEX":     keywordAttributeHex,
	"FLOAT":   keywordAttributeFloat,
	"STRING":  keywordAttributeString,
	"ENUM":    keywordAttributeEnum,

	"SG_MUL_VAL_": keywordExtendedMuxSignal,
}
