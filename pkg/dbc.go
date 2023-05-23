package pkg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

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

// DBCGenerator is a struct that wraps the methods to generate the DBC file.
type DBCGenerator struct{}

// NewDBCGenerator returns a new DBCGenerator.
func NewDBCGenerator() *DBCGenerator {
	return &DBCGenerator{}
}

// Generate generates the DBC file.
func (dbc *DBCGenerator) Generate(model *CanModel, file *os.File) {
	f := newFile(file)

	f.print("VERSION", formatString(model.Version))
	f.print(dbcHeaders)
	dbc.genNodes(f, model.Nodes)

	for msgName, msg := range model.Messages {
		dbc.genMessage(f, msgName, msg)
	}

	dbc.genMuxGroup(f, model.Messages)
	dbc.genBitmaps(f, model)
	dbc.genComments(f, model)
}

// genNodes generates the node definitions of the DBC file.
func (dbc *DBCGenerator) genNodes(f *file, nodes map[string]*Node) {
	nodeNames := []string{}
	for nodeName := range nodes {
		nodeNames = append(nodeNames, nodeName)
	}

	str := []string{symbols.DBCNode, ":"}
	str = append(str, nodeNames...)
	f.print(str...)
	f.print()
}

// genMessage generates the message definitions of the DBC file.
func (dbc *DBCGenerator) genMessage(f *file, msgName string, msg *Message) {
	id := fmt.Sprintf("%d", msg.ID)
	length := fmt.Sprintf("%d", msg.Length)
	sender := msg.Sender
	if sender == "" {
		sender = dbcDefNode
	}
	f.print(symbols.DBCMessage, id, msgName+":", length, sender)

	for sigName, sig := range msg.Signals {
		dbc.genSignal(f, sigName, sig, false)
	}

	f.print()
}

// genSignal generates the signal definitions of the DBC file.
func (dbc *DBCGenerator) genSignal(f *file, sigName string, sig *Signal, multiplexed bool) {
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
			dbc.genSignal(f, muxSigName, muxSig, true)
		}
	}

	f.print("\t", symbols.DBCSignal, sigName, muxStr, ":", byteDef, multiplier, valueRange, unit, receivers)
}

// genMuxGroup generates the multiplexed signals of the DBC file.
func (dbc *DBCGenerator) genMuxGroup(f *file, messages map[string]*Message) {
	for _, msg := range messages {
		for sigName, sig := range msg.Signals {
			if sig.IsMultiplexor() {
				for muxSigName, muxSig := range sig.MuxGroup {
					dbc.genMuxSignal(f, msg.FormatID(), sigName, muxSigName, muxSig)
				}
			}
		}
	}
}

// genMuxSignal generates a multiplexed signal value.
func (dbc *DBCGenerator) genMuxSignal(f *file, msgID, muxSigName, sigName string, sig *Signal) {
	if sig.IsMultiplexor() {
		for innSigName, innSig := range sig.MuxGroup {
			dbc.genMuxSignal(f, msgID, sigName, innSigName, innSig)
		}
	}

	f.print(symbols.DBCMuxValue, msgID, sigName, muxSigName, fmt.Sprintf("%d-%d", sig.MuxSwitch, sig.MuxSwitch), ";")
}

// genBitmaps generates the bitmats of the DBC file.
func (dbc *DBCGenerator) genBitmaps(f *file, m *CanModel) {
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

// genComments generates the comments of the DBC file.
func (dbc *DBCGenerator) genComments(f *file, m *CanModel) {
	for nodeName, node := range m.Nodes {
		dbc.genNodeComment(f, nodeName, node)
	}

	for _, msg := range m.Messages {
		dbc.genMessageComment(f, msg)
	}
}

// genNodeComment generates the comment of a node.
func (dbc *DBCGenerator) genNodeComment(f *file, nodeName string, node *Node) {
	if node.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCNode, nodeName, formatString(node.Description), ";")
	}
}

// genMessageComment generates the comment of a message.
func (dbc *DBCGenerator) genMessageComment(f *file, msg *Message) {
	msgID := msg.FormatID()
	if msg.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCMessage, msgID, formatString(msg.Description), ";")
	}

	for sigName, sig := range msg.Signals {
		dbc.genSignalComment(f, msgID, sigName, sig)
	}
}

// genSignalComment generates the comment of a signal.
func (dbc *DBCGenerator) genSignalComment(f *file, msgID, sigName string, sig *Signal) {
	if sig.HasDescription() {
		f.print(symbols.DBCComment, symbols.DBCSignal, msgID, sigName, formatString(sig.Description), ";")
	}

	if sig.IsMultiplexor() {
		for muxSigName, muxSig := range sig.MuxGroup {
			dbc.genSignalComment(f, msgID, muxSigName, muxSig)
		}
	}
}

func (dbc *DBCGenerator) Read(file *os.File) *CanModel {
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	lines := []string{}

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	m := &CanModel{}

	version, _ := dbc.readVersion(lines)
	m.Version = version

	return m
}

var dbcVersionRegex = regexp.MustCompile(`^(VERSION) *\"(?P<version>.+)\"`)

func (dbc *DBCGenerator) readVersion(lines []string) (string, error) {
	for _, line := range lines {
		match := dbcVersionRegex.FindAllStringSubmatch(line, -1)
		if len(match) == 0 {
			continue
		}

		versionIndex := dbcVersionRegex.SubexpIndex("version")

		log.Print(match[0][versionIndex])
	}

	return "", nil
}
