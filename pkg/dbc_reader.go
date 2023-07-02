package pkg

import (
	"bufio"
	"os"
	"regexp"

	"github.com/squadracorsepolito/jsondbc/pkg/reg"
	"github.com/squadracorsepolito/jsondbc/pkg/sym"
)

type DBCReader struct {
	currLine  int
	fileLines []string
}

func NewDBCReader() *DBCReader {
	return &DBCReader{
		currLine:  0,
		fileLines: []string{},
	}
}

func (r *DBCReader) Read(file *os.File) (*CanModel, error) {
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	lines := []string{}

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	cfg := &textReaderCfg{
		varsionIdent:   sym.DBCVersion,
		busSpeedIdent:  sym.DBCBusSpeed,
		nodeIdent:      sym.DBCNode,
		msgIdent:       sym.DBCMessage,
		sigIdent:       sym.DBCSignal,
		extMuxSigIdent: sym.DBCExtMuxValue,
		bitmapDefIdent: sym.DBCValue,
		commentIdent:   sym.DBCComment,
		attIdent:       sym.DBCAttDef,
		attDefIdent:    sym.DBCAttDefaultVal,
		attAssIdent:    sym.DBCAttAssignment,
	}
	reader := &textReader{
		cfg: cfg,

		lines: lines,

		versionReg: reg.DBCVersion,

		busSpeedReg: reg.DBCBusSpeed,

		nodeReg: reg.DBCNode,

		msgStartReg: reg.DBCMessage,
		msgEndReg:   regexp.MustCompile(`^$`),

		sigReg:       reg.DBCSignal,
		extMuxSigReg: reg.DBCExtMuxSignal,

		bitmapDefReg: reg.DBCBitmapDef,

		nodeCommentReg: reg.DBCNodeComment,
		msgCommentReg:  reg.DBCMessageComment,
		sigCommentReg:  reg.DBCSignalComment,

		attReg: reg.DBCAttribute,

		attDefReg: reg.DBCAttributeDefault,

		nodeAttAssReg: reg.DBCNodeAttributeAssignment,
		msgAttAssReg:  reg.DBCMessageAttributeAssignment,
		sigAttAssReg:  reg.DBCSignalAttributeAssignment,
	}

	return reader.read()
}
