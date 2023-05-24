package pkg

import (
	"fmt"
	"os"

	"github.com/FerroO2000/canconv/pkg/symbols"
)

const dbcDefNode = "Vector__XXX"

const dbcHeaders = `
NS_ :
	NS_DESC_
	CM_
	BA_DEF_
	BA_
	VAL_
	CAT_DEF_
	CAT_
	FILTER
	BA_DEF_DEF_
	EV_DATA_
	ENVVAR_DATA_
	SGTYPE_
	SGTYPE_VAL_
	BA_DEF_SGTYPE_
	BA_SGTYPE_
	SIG_TYPE_REF_
	VAL_TABLE_
	SIG_GROUP_
	SIG_VALTYPE_
	SIGTYPE_VALTYPE_
	BO_TX_BU_
	BA_DEF_REL_
	BA_REL_
	BA_DEF_DEF_REL_
	BU_SG_REL_
	BU_EV_REL_
	BU_BO_REL_

BS_ :
`

type DBCWriter struct{}

func NewDBCWriter() *DBCWriter {
	return &DBCWriter{}
}

func (w *DBCWriter) Write(file *os.File, canModel *CanModel) error {
	f := newFile(file)

	f.print("VERSION", formatString(canModel.Version))
	f.print(dbcHeaders)
	w.writeNodes(f, canModel.Nodes)

	for msgName, msg := range canModel.Messages {
		w.writeMessage(f, msgName, msg)
	}

	w.writeMuxGroup(f, canModel.Messages)
	w.writeBitmaps(f, canModel)
	w.writeComments(f, canModel)

	return nil
}

func (w *DBCWriter) writeNodes(f *file, nodes map[string]*Node) {
	nodeNames := []string{}
	for nodeName := range nodes {
		nodeNames = append(nodeNames, nodeName)
	}

	str := []string{symbols.DBCNode, ":"}
	str = append(str, nodeNames...)
	f.print(str...)
	f.print()
}

func (w *DBCWriter) writeMessage(f *file, msgName string, msg *Message) {
	id := fmt.Sprintf("%d", msg.ID)
	length := fmt.Sprintf("%d", msg.Length)
	sender := msg.Sender
	if sender == "" {
		sender = dbcDefNode
	}
	f.print(symbols.DBCMessage, id, msgName+":", length, sender)

	for sigName, sig := range msg.Signals {
		w.writeSignal(f, sigName, sig, false)
	}

	f.print()
}

func (w *DBCWriter) writeSignal(f *file, sigName string, sig *Signal, multiplexed bool) {
	byteOrder := 0
	if sig.BigEndian {
		byteOrder = 1
	}
	valueType := "+"
	if sig.Signed {
		valueType = "-"
	}
	byteDef := fmt.Sprintf("%d|%d@%d%s", sig.StartBit, sig.Size, byteOrder, valueType)
	multiplier := fmt.Sprintf("(%s,%s)", formatFloat(sig.Scale), formatFloat(sig.Offset))
	valueRange := fmt.Sprintf("[%s|%s]", formatFloat(sig.Min), formatFloat(sig.Max))
	unit := fmt.Sprintf(`"%s"`, sig.Unit)

	receivers := ""
	if len(sig.Receivers) == 0 {
		receivers = dbcDefNode
	} else {
		for i, r := range sig.Receivers {
			if i == 0 {
				receivers += r
				continue
			}
			receivers += "," + r
		}
	}

	muxStr := ""
	if multiplexed {
		muxStr = "m" + formatUint(sig.MuxSwitch)
	}
	if sig.IsMultiplexor() {
		muxStr += "M"

		for muxSigName, muxSig := range sig.MuxGroup {
			w.writeSignal(f, muxSigName, muxSig, true)
		}
	}

	f.print("\t", symbols.DBCSignal, sigName, muxStr, ":", byteDef, multiplier, valueRange, unit, receivers)
}

func (w *DBCWriter) writeMuxGroup(f *file, messages map[string]*Message) {
	for _, msg := range messages {
		for sigName, sig := range msg.Signals {
			if sig.IsMultiplexor() {
				for muxSigName, muxSig := range sig.MuxGroup {
					w.writeMuxSignal(f, msg.FormatID(), sigName, muxSigName, muxSig)
				}
			}
		}
	}
}

func (w *DBCWriter) writeMuxSignal(f *file, msgID, muxSigName, sigName string, sig *Signal) {
	if sig.IsMultiplexor() {
		for innSigName, innSig := range sig.MuxGroup {
			w.writeMuxSignal(f, msgID, sigName, innSigName, innSig)
		}
	}

	f.print(symbols.DBCMuxValue, msgID, sigName, muxSigName, fmt.Sprintf("%d-%d", sig.MuxSwitch, sig.MuxSwitch), ";")
}

func (w *DBCWriter) writeBitmaps(f *file, m *CanModel) {
	for _, msg := range m.Messages {
		for sigName, sig := range msg.Signals {
			if sig.IsBitmap() {
				bitmap := ""
				first := true
				for name, val := range sig.Bitmap {
					if first {
						bitmap += formatUint(val) + " " + formatString(name)
						first = false
						continue
					}
					bitmap += " " + formatUint(val) + " " + formatString(name)
				}
				f.print(symbols.DBCValue, msg.FormatID(), sigName, bitmap, ";")
			}
		}
	}
}

func (w *DBCWriter) writeComments(f *file, m *CanModel) {
	for nodeName, node := range m.Nodes {
		w.writeNodeComment(f, nodeName, node)
	}

	for _, msg := range m.Messages {
		w.writeMessageComment(f, msg)
	}
}

func (w *DBCWriter) writeNodeComment(f *file, nodeName string, node *Node) {
	if node.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCNode, nodeName, formatString(node.Description), ";")
	}
}

func (w *DBCWriter) writeMessageComment(f *file, msg *Message) {
	msgID := msg.FormatID()
	if msg.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCMessage, msgID, formatString(msg.Description), ";")
	}

	for sigName, sig := range msg.Signals {
		w.writeSignalComment(f, msgID, sigName, sig)
	}
}

func (w *DBCWriter) writeSignalComment(f *file, msgID, sigName string, sig *Signal) {
	if sig.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCSignal, msgID, sigName, formatString(sig.Description), ";")
	}

	if sig.IsMultiplexor() {
		for muxSigName, muxSig := range sig.MuxGroup {
			w.writeSignalComment(f, msgID, muxSigName, muxSig)
		}
	}
}
