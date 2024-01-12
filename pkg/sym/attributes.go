package sym

const (
	MsgPeriodAttribute string = "MsgPeriodMS"
	BaudrateAttribute  string = "Baudrate"
)

const MsgCycleTime string = "GenMsgCycleTime"

const MsgSendType string = "GenMsgSendType"

var MsgSendTypeValues = []string{"NoMsgSendType", "Cyclic", "IfActive", "cyclicIfActive", "NotUsed"}

const SigSendType string = "GenSigSendType"

var SigSendTypeValues = []string{"NoSigSendType", "Cyclic", "OnWrite", "OnWriteWithRepetition", "OnChange", "OnChangeWithRepetition", "IfActive", "IfActiveWithRepetition", "NotUsed"}
