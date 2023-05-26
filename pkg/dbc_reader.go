package pkg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/FerroO2000/canconv/pkg/symbols"
	"golang.org/x/sync/errgroup"
)

var dbcVersionRegex = regexp.MustCompile(fmt.Sprintf(`^(?:%s) *\"(?P<version>.+)\"$`, symbols.DBCVersion))

var dbcBusSpeedRegex = regexp.MustCompile(fmt.Sprintf(`^(?:%s *\:) *(?P<speed>\d+)$`, symbols.DBCBusSpeed))

var dbcNodesRegex = regexp.MustCompile(fmt.Sprintf(`^(?:%s *\:)(?: *)(?P<nodes>.*)`, symbols.DBCNode))

var dbcMessageRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?P<id>\d+) *(?P<name>\w+) *: *(?P<length>\d+) *(?P<sender>\w+)$`, symbols.DBCMessage),
)

var dbcSignalRegex = regexp.MustCompile(
	fmt.Sprintf(
		`^(?:(?:\t| *)%s) *(?P<name>\w+) *(?P<mux_switch>m\d+)?(?P<mux>M)?(?: *): (?P<start_bit>\d+)\|(?P<size>\d+)@(?P<order>0|1)(?P<signed>\+|\-) *\((?P<scale>-?\d+\.?\d+|-?\d+),(?P<offset>-?\d+\.?\d+|-?\d+)\) *\[(?P<min>-?\d+\.?\d+|-?\d+)\|(?P<max>-?\d+\.?\d+|-?\d+)\] *"(?P<unit>.*)" *(?P<receivers>.*)$`,
		symbols.DBCSignal,
	),
)

var dbcMuxValueRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<mux_name>\w+) *(?:\d+\-?){2} *;$`, symbols.DBCExtMuxValue),
)

var dbcSignalBitmapRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<bitmap>.*);$`, symbols.DBCValue),
)

var dbcNodeCommentRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?:%s) *(?P<name>\w+) *"(?P<desc>.*)" *;$`, symbols.DBCComment, symbols.DBCNode),
)

var dbcMessageCommentRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?:%s) *(?P<id>\d+) *"(?P<desc>.*)" *;$`, symbols.DBCComment, symbols.DBCMessage),
)

var dbcSignalCommentRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *"(?P<desc>.*)" *;$`, symbols.DBCComment, symbols.DBCSignal),
)

var dbcExtMuxValueRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<mux_name>\w+) *(?:\d+\-?){2} *;$`, symbols.DBCExtMuxValue),
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

	can := &CanModel{
		Nodes:             make(map[string]*Node),
		Messages:          make(map[string]*Message),
		NodeAttributes:    make(map[string]*Attribute),
		MessageAttributes: make(map[string]*Attribute),
		SignalAttributes:  make(map[string]*Attribute),
	}

	splMuxSignals := make(map[string]*Signal)
	extMuxSignals := make(map[string]map[string]*Signal)
	parentMsgName := ""
	for lineNum, line := range lines {
		r.currLine = lineNum

		if can.Version == "" {
			version, err := r.readVersion(line)
			if !errors.Is(err, errRegexNoMatch) {
				if err != nil {
					return nil, err
				}

				can.Version = version
				continue
			}
		}

		busSpeed, err := r.readBusSpeed(line)
		if !errors.Is(err, errRegexNoMatch) {
			if err != nil {
				return nil, err
			}

			can.BusSpeed = busSpeed
			continue
		}

		if len(can.Nodes) == 0 {
			nodes, err := r.readNodes(line)
			if !errors.Is(err, errRegexNoMatch) {
				if err != nil {
					return nil, err
				}

				can.Nodes = nodes
				continue
			}
		}

		if parentMsgName == "" {
			msg, err := r.readMessage(line)
			if !errors.Is(err, errRegexNoMatch) {
				if err != nil {
					return nil, err
				}

				can.Messages[msg.name] = msg
				parentMsgName = msg.name

				continue
			}
		}

		if parentMsgName != "" {
			sig, err := r.readSignal(line)
			if !errors.Is(err, errRegexNoMatch) {
				if err != nil {
					return nil, err
				}

				if !sig.isMultiplexed {
					can.Messages[parentMsgName].Signals[sig.name] = sig
				} else if sig.isMultiplexed && !sig.isMultiplexor {
					if len(extMuxSignals[parentMsgName]) == 0 {
						splMuxSignals[sig.name] = sig
					} else {
						extMuxSignals[parentMsgName][sig.name] = sig
					}

				} else if sig.isMultiplexed && sig.isMultiplexor {
					if len(extMuxSignals[parentMsgName]) == 0 {
						extMuxSignals[parentMsgName] = make(map[string]*Signal)
					}

					extMuxSignals[parentMsgName][sig.name] = sig
					for sigName, muxSig := range splMuxSignals {
						extMuxSignals[parentMsgName][sigName] = muxSig
						delete(splMuxSignals, sigName)
					}
				}

				continue
			}

			for sigName, muxSig := range splMuxSignals {
				for _, muxorSig := range can.Messages[parentMsgName].Signals {
					if muxorSig.isMultiplexor {
						muxorSig.MuxGroup[sigName] = muxSig

						break
					}
				}

				delete(splMuxSignals, sigName)
			}

			parentMsgName = ""
		}

		bitmap, err := r.readBitmap(line)
		if !errors.Is(err, errRegexNoMatch) {
			if err != nil {
				return nil, err
			}

			for msgName, msg := range can.Messages {
				if msg.ID == bitmap.messageID {
					if sig, ok := msg.Signals[bitmap.signalName]; ok {
						sig.Bitmap = bitmap.bitmap
					} else if sig, ok := extMuxSignals[msgName][bitmap.signalName]; ok {
						sig.Bitmap = bitmap.bitmap
					}

					break
				}
			}

			continue
		}

		nodeComment, err := r.readNodeComment(line)
		if !errors.Is(err, errRegexNoMatch) {
			if err != nil {
				return nil, err
			}

			if node, ok := can.Nodes[nodeComment.nodeName]; ok {
				node.Description = nodeComment.description
			} else {
				return nil, r.lineErr(fmt.Sprintf("node %s doesn't exist", nodeComment.nodeName))
			}

			continue
		}

		msgComment, err := r.readMessageComment(line)
		if !errors.Is(err, errRegexNoMatch) {
			if err != nil {
				return nil, err
			}

			found := false
			for _, msg := range can.Messages {
				if msg.ID == msgComment.messageID {
					msg.Description = msgComment.description
					found = true
					break
				}
			}
			if !found {
				return nil, r.lineErr(fmt.Sprintf("message %d doesn't exist", msgComment.messageID))
			}

			continue
		}

		sigComment, err := r.readSignalComment(line)
		if !errors.Is(err, errRegexNoMatch) {
			if err != nil {
				return nil, err
			}

			found := false
			for msgName, msg := range can.Messages {
				if msg.ID == sigComment.messageID {
					if sig, ok := msg.Signals[sigComment.signalName]; ok {
						sig.Description = sigComment.description
					} else if sig, ok := extMuxSignals[msgName][sigComment.signalName]; ok {
						sig.Description = sigComment.description
					} else {
						for _, muxorSig := range msg.Signals {
							if muxorSig.isMultiplexor {
								if muxedSig, ok := muxorSig.MuxGroup[sigComment.signalName]; ok {
									muxedSig.Description = sigComment.description
									break
								}

								return nil, r.lineErr(fmt.Sprintf("signal %s in message %d doesn't exist", sigComment.signalName, sigComment.messageID))
							}
						}

					}

					found = true

					break
				}
			}
			if !found {
				return nil, r.lineErr(fmt.Sprintf("message %d doesn't exist", sigComment.messageID))
			}

			continue
		}

		if len(extMuxSignals) > 0 {
			extVal, err := r.readExtMuxValue(line)
			if !errors.Is(err, errRegexNoMatch) {
				if err != nil {
					return nil, err
				}

				found := false
				for msgName, msg := range can.Messages {
					if msg.ID == extVal.messageID {
						if muxorSig, ok := msg.Signals[extVal.multiplexorSignalName]; ok {
							if muxSig, muxOk := extMuxSignals[msgName][extVal.signalName]; muxOk {
								muxorSig.MuxGroup[extVal.signalName] = muxSig
							} else {
								return nil, r.lineErr(fmt.Sprintf("multiplexed signal %s in message %d doesn't exist", extVal.signalName, extVal.messageID))
							}

						} else if muxorSig, ok := extMuxSignals[msgName][extVal.multiplexorSignalName]; ok {
							if muxSig, muxOk := extMuxSignals[msgName][extVal.signalName]; muxOk {
								muxorSig.MuxGroup[extVal.signalName] = muxSig
							} else {
								return nil, r.lineErr(fmt.Sprintf("multiplexed signal %s in message %d doesn't exist", extVal.signalName, extVal.messageID))
							}

						} else {
							return nil, r.lineErr(fmt.Sprintf("multiplexor signal %s in message %d doesn't exist", extVal.multiplexorSignalName, extVal.messageID))
						}

						found = true

						break
					}
				}
				if !found {
					return nil, r.lineErr(fmt.Sprintf("message %d doesn't exist", extVal.messageID))
				}

				continue
			}
		}

	}

	r.fileLines = lines

	attLineNumbers := []int{}
	attDefValLineNumbers := []int{}
	for n, line := range lines {
		if strings.HasPrefix(line, symbols.DBCAttDef+" ") {
			attLineNumbers = append(attLineNumbers, n)
			continue
		}
		if strings.HasPrefix(line, symbols.DBCAttDefaultVal+" ") {
			attDefValLineNumbers = append(attDefValLineNumbers, n)
			continue
		}
	}

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		attributes, err := r.handleAttributes(attLineNumbers)
		if err != nil {
			return err
		}

		for _, att := range attributes {
			switch att.attributeKind {
			case attributeKindNode:
				can.NodeAttributes[att.name] = att
			case attributeKindMessage:
				can.MessageAttributes[att.name] = att
			case attributeKindSignal:
				can.SignalAttributes[att.name] = att
			}
		}

		return nil
	})

	defAttValues := []*dbcDefAttVal{}
	g.Go(func() error {
		tmpDefAttValues, err := r.handleDefaultAttributeValues(attDefValLineNumbers)
		if err != nil {
			return err
		}
		defAttValues = tmpDefAttValues

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, defAttVal := range defAttValues {
		if att, ok := can.NodeAttributes[defAttVal.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				att.Int.Default = defAttVal.intVal
			case attributeTypeString:
				att.String.Default = defAttVal.stringVal
			case attributeTypeEnum:
				enumIdx := defAttVal.intVal
				if enumIdx >= 0 && enumIdx < len(att.Enum.Values) {
					att.Enum.Default = att.Enum.Values[enumIdx]
					att.Enum.defaultIdx = enumIdx
				}
			}
		}

		if att, ok := can.MessageAttributes[defAttVal.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				att.Int.Default = defAttVal.intVal
			case attributeTypeString:
				att.String.Default = defAttVal.stringVal
			case attributeTypeEnum:
				enumIdx := defAttVal.intVal
				if enumIdx >= 0 && enumIdx < len(att.Enum.Values) {
					att.Enum.Default = att.Enum.Values[enumIdx]
					att.Enum.defaultIdx = enumIdx
				}
			}
		}

		if att, ok := can.SignalAttributes[defAttVal.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				att.Int.Default = defAttVal.intVal
			case attributeTypeString:
				att.String.Default = defAttVal.stringVal
			case attributeTypeEnum:
				enumIdx := defAttVal.intVal
				if enumIdx >= 0 && enumIdx < len(att.Enum.Values) {
					att.Enum.Default = att.Enum.Values[enumIdx]
					att.Enum.defaultIdx = enumIdx
				}
			}
		}
	}

	return can, nil
}

func (r *DBCReader) handleAttributes(lineNumbers []int) ([]*Attribute, error) {
	attributes := []*Attribute{}

	for _, idx := range lineNumbers {
		line := r.fileLines[idx]
		att, err := r.readAttribute(line)
		if err != nil {
			if errors.Is(err, errRegexNoMatch) {
				continue
			}
			return nil, fmt.Errorf("line %d: attribute: %w", idx, err)
		}
		attributes = append(attributes, att)
	}

	return attributes, nil
}

var dbcAttRegex = regexp.MustCompile(
	fmt.Sprintf(
		`^(?:%s) *(?P<att_kind>%s|%s|%s) *"(?P<att_name>\w+)" *(?P<att_type>%s|%s|%s) *`,
		symbols.DBCAttDef,
		symbols.DBCNode,
		symbols.DBCMessage,
		symbols.DBCSignal,
		symbols.DBCAttIntType,
		symbols.DBCAttStringType,
		symbols.DBCAttEnumType,
	),
)

func (r *DBCReader) readAttribute(line string) (*Attribute, error) {
	match, err := applyRegex(dbcAttRegex, line)
	if err != nil {
		return nil, err
	}

	att := &Attribute{
		name: match[dbcAttRegex.SubexpIndex("att_name")],
	}

	switch match[dbcAttRegex.SubexpIndex("att_kind")] {
	case symbols.DBCNode:
		att.attributeKind = attributeKindNode
	case symbols.DBCMessage:
		att.attributeKind = attributeKindMessage
	case symbols.DBCSignal:
		att.attributeKind = attributeKindSignal
	}

	switch match[dbcAttRegex.SubexpIndex("att_type")] {
	case symbols.DBCAttIntType:
		att.attributeType = attributeTypeInt
		intAtt, err := r.readIntAttribute(line)
		if err != nil {
			return nil, err
		}
		att.Int = intAtt

	case symbols.DBCAttStringType:
		att.attributeType = attributeTypeString
		att.String = &AttributeString{}

	case symbols.DBCAttEnumType:
		att.attributeType = attributeTypeEnum
		enumAtt, err := r.readEnumAttribute(line)
		if err != nil {
			return nil, err
		}
		att.Enum = enumAtt
	}

	return att, nil
}

var dbcIntAttRegex = regexp.MustCompile(
	fmt.Sprintf(
		`^(?:%s) *(?:%s|%s|%s) *"(?:\w+)" *(?:%s) *(?P<from>-?\d+) *(?P<to>-?\d+) *;$`,
		symbols.DBCAttDef,
		symbols.DBCNode,
		symbols.DBCMessage,
		symbols.DBCSignal,
		symbols.DBCAttIntType,
	),
)

func (r *DBCReader) readIntAttribute(line string) (*AttributeInt, error) {
	match, err := applyRegex(dbcIntAttRegex, line)
	if err != nil {
		return nil, err
	}

	from, err := parseInt(match[dbcIntAttRegex.SubexpIndex("from")])
	if err != nil {
		return nil, err
	}
	to, err := parseInt(match[dbcIntAttRegex.SubexpIndex("to")])
	if err != nil {
		return nil, err
	}

	return &AttributeInt{
		From: from,
		To:   to,
	}, nil
}

var dbcEnumAttRegex = regexp.MustCompile(
	fmt.Sprintf(
		`^(?:%s) *(?:%s|%s|%s) *"(?:\w+)" *(?:%s) *(?P<enum>.*);$`,
		symbols.DBCAttDef,
		symbols.DBCNode,
		symbols.DBCMessage,
		symbols.DBCSignal,
		symbols.DBCAttEnumType,
	),
)

func (r *DBCReader) readEnumAttribute(line string) (*AttributeEnum, error) {
	match, err := applyRegex(dbcEnumAttRegex, line)
	if err != nil {
		return nil, err
	}

	enums := strings.Split(strings.TrimSpace(match[dbcEnumAttRegex.SubexpIndex("enum")]), ",")
	values := []string{}
	for _, e := range enums {
		values = append(values, strings.Replace(e, "\"", "", -1))
	}

	return &AttributeEnum{
		Values: values,
	}, nil
}

var dbcDefAttValRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *"(?P<att_name>\w+)" *(?P<def_val>""|"\w+"|-?\d+) *;$`, symbols.DBCAttDefaultVal),
)

type dbcDefAttVal struct {
	attName   string
	intVal    int
	stringVal string
}

func (r *DBCReader) handleDefaultAttributeValues(lineNumbers []int) ([]*dbcDefAttVal, error) {
	attValues := []*dbcDefAttVal{}

	for _, idx := range lineNumbers {
		line := r.fileLines[idx]
		attVal, err := r.readDefaultAttributeValue(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: default attribute value: %w", idx, err)
		}
		attValues = append(attValues, attVal)
	}

	return attValues, nil
}

func (r *DBCReader) readDefaultAttributeValue(line string) (*dbcDefAttVal, error) {
	match, err := applyRegex(dbcDefAttValRegex, line)
	if err != nil {
		return nil, err
	}

	attVal := &dbcDefAttVal{
		attName: match[dbcDefAttValRegex.SubexpIndex("att_name")],
	}

	strVal := strings.TrimSpace(match[dbcDefAttValRegex.SubexpIndex("def_val")])
	if len(strVal) == 2 {
		return attVal, nil
	}

	re := regexp.MustCompile(`^"(?P<val>\w+)"$`)
	submatch, err := applyRegex(re, strVal)
	if err != nil {
		intVal, err := parseInt(strVal)
		if err != nil {
			return nil, err
		}
		attVal.intVal = intVal
		return attVal, nil
	}

	attVal.stringVal = submatch[re.SubexpIndex("val")]

	return attVal, nil
}

func (r *DBCReader) lineErr(errStr string) error {
	return fmt.Errorf("line %d: %s", r.currLine, errStr)
}

func (r *DBCReader) readVersion(line string) (string, error) {
	match, err := applyRegex(dbcVersionRegex, line)
	if err != nil {
		return "", err
	}

	return match[dbcVersionRegex.SubexpIndex("version")], nil
}

func (r *DBCReader) readBusSpeed(line string) (uint32, error) {
	match, err := applyRegex(dbcBusSpeedRegex, line)
	if err != nil {
		return 0, err
	}

	speed, err := parseUint(match[dbcBusSpeedRegex.SubexpIndex("speed")])
	if err != nil {
		return 0, err
	}

	return speed, nil
}

func (r *DBCReader) readNodes(line string) (map[string]*Node, error) {
	match, err := applyRegex(dbcNodesRegex, line)
	if err != nil {
		return nil, err
	}

	strNodes := strings.TrimSpace(match[dbcNodesRegex.SubexpIndex("nodes")])
	nodeNames := strings.Split(strNodes, " ")
	nodes := make(map[string]*Node, len(nodeNames))
	for _, name := range nodeNames {
		nodes[name] = &Node{}
	}

	return nodes, nil
}

func (r *DBCReader) readSignal(line string) (*Signal, error) {
	match, err := applyRegex(dbcSignalRegex, line)
	if err != nil {
		return nil, err
	}

	name := match[dbcSignalRegex.SubexpIndex("name")]

	strMuxSwitch := match[dbcSignalRegex.SubexpIndex("mux_switch")]
	muxSwitch := uint32(0)
	isMultiplexed := strMuxSwitch != ""
	if isMultiplexed {
		tmpMuxSwitch, err := parseUint(strMuxSwitch[1:])
		if err != nil {
			return nil, err
		}
		muxSwitch = tmpMuxSwitch
	}
	isMultiplexor := match[dbcSignalRegex.SubexpIndex("mux")] == "M"

	startBit, err := parseUint(match[dbcSignalRegex.SubexpIndex("start_bit")])
	if err != nil {
		return nil, err
	}
	size, err := parseUint(match[dbcSignalRegex.SubexpIndex("size")])
	if err != nil {
		return nil, err
	}

	bigEndian := false
	if match[dbcSignalRegex.SubexpIndex("order")] == "1" {
		bigEndian = true
	}
	signed := false
	if match[dbcSignalRegex.SubexpIndex("signed")] == "-" {
		signed = true
	}

	scale, err := parseFloat(match[dbcSignalRegex.SubexpIndex("scale")])
	if err != nil {
		return nil, err
	}
	offset, err := parseFloat(match[dbcSignalRegex.SubexpIndex("offset")])
	if err != nil {
		return nil, err
	}

	min, err := parseFloat(match[dbcSignalRegex.SubexpIndex("min")])
	if err != nil {
		return nil, err
	}
	strMax := match[dbcSignalRegex.SubexpIndex("max")]
	max, err := parseFloat(strMax)
	if err != nil {
		return nil, err
	}

	unit := match[dbcSignalRegex.SubexpIndex("unit")]

	tmpReceivers := strings.Split(match[dbcSignalRegex.SubexpIndex("receivers")], ",")
	receivers := []string{}
	for _, tmp := range tmpReceivers {
		if tmp != dbcDefNode {
			receivers = append(receivers, tmp)
		}
	}

	return &Signal{
		StartBit:  uint32(startBit),
		Size:      uint32(size),
		BigEndian: bigEndian,
		Signed:    signed,
		Unit:      unit,
		Receivers: receivers,
		Scale:     scale,
		Offset:    offset,
		Min:       min,
		Max:       max,
		Bitmap:    make(map[string]uint32),
		MuxGroup:  make(map[string]*Signal),
		MuxSwitch: muxSwitch,

		name:          name,
		isMultiplexor: isMultiplexor,
		isMultiplexed: isMultiplexed,
	}, nil
}

func (r *DBCReader) readMessage(line string) (*Message, error) {
	match, err := applyRegex(dbcMessageRegex, line)
	if err != nil {
		return nil, err
	}

	id, err := parseUint(match[dbcMessageRegex.SubexpIndex("id")])
	if err != nil {
		return nil, err
	}
	name := match[dbcMessageRegex.SubexpIndex("name")]
	length, err := parseUint(match[dbcMessageRegex.SubexpIndex("length")])
	if err != nil {
		return nil, err
	}
	sender := match[dbcMessageRegex.SubexpIndex("sender")]

	return &Message{
		ID:      id,
		Length:  length,
		Sender:  sender,
		Signals: make(map[string]*Signal),

		name: name,
	}, nil
}

type dbcExtMuxValue struct {
	messageID             uint32
	multiplexorSignalName string
	signalName            string
}

func (r *DBCReader) readExtMuxValue(line string) (*dbcExtMuxValue, error) {
	match, err := applyRegex(dbcMuxValueRegex, line)
	if err != nil {
		return nil, err
	}

	msgID, err := parseUint(match[dbcExtMuxValueRegex.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[dbcExtMuxValueRegex.SubexpIndex("sig_name")]
	muxName := match[dbcExtMuxValueRegex.SubexpIndex("mux_name")]

	return &dbcExtMuxValue{
		messageID:             msgID,
		multiplexorSignalName: muxName,
		signalName:            sigName,
	}, nil
}

type dbcBitmap struct {
	messageID  uint32
	signalName string
	bitmap     map[string]uint32
}

func (r *DBCReader) readBitmap(line string) (*dbcBitmap, error) {
	match, err := applyRegex(dbcSignalBitmapRegex, line)
	if err != nil {
		return nil, err
	}

	msgID, err := parseUint(match[dbcSignalBitmapRegex.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[dbcSignalBitmapRegex.SubexpIndex("sig_name")]

	strBitmap := match[dbcSignalBitmapRegex.SubexpIndex("bitmap")]
	splBitmap := strings.Split(strings.TrimSpace(strBitmap), " ")
	bitmap := make(map[string]uint32, len(splBitmap)/2)
	for i := 0; i < len(splBitmap); i = i + 2 {
		val, _ := strconv.ParseUint(splBitmap[i], 10, 32)
		name := strings.Replace(splBitmap[i+1], `"`, "", 2)
		bitmap[name] = uint32(val)
	}

	return &dbcBitmap{
		messageID:  msgID,
		signalName: sigName,
		bitmap:     bitmap,
	}, nil
}

type dbcNodeComment struct {
	nodeName    string
	description string
}

func (r *DBCReader) readNodeComment(line string) (*dbcNodeComment, error) {
	match, err := applyRegex(dbcNodeCommentRegex, line)
	if err != nil {
		return nil, err
	}

	name := match[dbcNodeCommentRegex.SubexpIndex("name")]
	desc := match[dbcNodeCommentRegex.SubexpIndex("desc")]

	return &dbcNodeComment{
		nodeName:    name,
		description: desc,
	}, nil
}

type dbcMessageComment struct {
	messageID   uint32
	description string
}

func (r *DBCReader) readMessageComment(line string) (*dbcMessageComment, error) {
	match, err := applyRegex(dbcMessageCommentRegex, line)
	if err != nil {
		return nil, err
	}

	msgID, err := parseUint(match[dbcMessageCommentRegex.SubexpIndex("id")])
	if err != nil {
		return nil, err
	}
	desc := match[dbcMessageCommentRegex.SubexpIndex("desc")]

	return &dbcMessageComment{
		messageID:   msgID,
		description: desc,
	}, nil
}

type dbcSignalComment struct {
	messageID   uint32
	signalName  string
	description string
}

func (r *DBCReader) readSignalComment(line string) (*dbcSignalComment, error) {
	match, err := applyRegex(dbcSignalCommentRegex, line)
	if err != nil {
		return nil, err
	}

	msgID, err := parseUint(match[dbcSignalCommentRegex.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[dbcSignalCommentRegex.SubexpIndex("sig_name")]
	desc := match[dbcSignalCommentRegex.SubexpIndex("desc")]

	return &dbcSignalComment{
		messageID:   msgID,
		signalName:  sigName,
		description: desc,
	}, nil
}
