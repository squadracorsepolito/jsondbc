package pkg

import (
	"fmt"
	"os"

	"github.com/FerroO2000/canconv/symbols"
)

const dbcDefNode = "Vector_XXX"

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

BU_:
`

// DBCGenerator is a struct that wraps the methods to generate the DBC file.
type DBCGenerator struct{}

// NewDBCGenerator returns a new DBCGenerator.
func NewDBCGenerator() *DBCGenerator {
	return &DBCGenerator{}
}

// Generate generates the DBC file.
func (g *DBCGenerator) Generate(model *CanModel, file *os.File) {
	f := newFile(file)

	f.print("VERSION", formatString(model.Version))

	f.print(dbcHeaders)

	g.genNodes(f, model.Nodes)
	f.print()

	for msgName, msg := range model.Messages {
		g.genMessage(f, msgName, &msg)
	}
	f.print()

	g.genComments(f, model)
	f.print()

	g.genBitmaps(f, model)
}

// genNodes generates the node definitions of the DBC file.
func (g *DBCGenerator) genNodes(f *file, nodes map[string]Node) {
	nodeNames := []string{}
	for nodeName := range nodes {
		nodeNames = append(nodeNames, nodeName)
	}

	str := []string{symbols.DBCNode, ":"}
	str = append(str, nodeNames...)
	f.print(str...)
}

// genMessage generates the message definitions of the DBC file.
func (g *DBCGenerator) genMessage(f *file, msgName string, msg *Message) {
	id := fmt.Sprintf("%d", msg.ID)
	length := fmt.Sprintf("%d", msg.Length)
	sender := msg.Sender
	if sender == "" {
		sender = dbcDefNode
	}
	f.print(symbols.DBCMessage, id, msgName+":", length, sender)

	for sigName, sig := range msg.Signals {
		sig.Validate()
		g.genSignal(f, sigName, &sig)
	}
}

// genSignal generates the signal definitions of the DBC file.
func (g *DBCGenerator) genSignal(f *file, sigName string, sig *Signal) {
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

	f.print("", symbols.DBCSignal, sigName, ":", byteDef, multiplier, valueRange, unit, receivers)
}

// genComments generates the comments of the DBC file.
func (g *DBCGenerator) genComments(f *file, m *CanModel) {
	for nodeName, node := range m.Nodes {
		if node.HasDescription() {
			f.print(symbols.DBCComment, symbols.DBCNode, nodeName, formatString(node.Description), ";")
		}
	}
	f.print()

	for _, msg := range m.Messages {
		if msg.HasDescription() {
			f.print(symbols.DBCComment, symbols.DBCMessage, msg.FormatID(), formatString(msg.Description), ";")
		}

		for sigName, sig := range msg.Signals {
			if sig.HasDescription() {
				f.print(symbols.DBCComment, symbols.DBCSignal, msg.FormatID(), sigName, formatString(sig.Description), ";")
			}
		}
	}
}

// genBitmaps generates the 'VAL_' of the DBC file.
func (g *DBCGenerator) genBitmaps(f *file, m *CanModel) {
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
