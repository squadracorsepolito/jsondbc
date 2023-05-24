package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/FerroO2000/canconv/pkg/symbols"
)

var dbcVersionRegex = regexp.MustCompile(fmt.Sprintf(`^(?:%s) *\"(?P<version>.+)\"$`, symbols.DBCVersion))

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
	fmt.Sprintf(`^(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<mux_name>\w+) *(?:\d+\-?){2} *;$`, symbols.DBCMuxValue),
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

var dbcMuxSignalRegex = regexp.MustCompile(
	fmt.Sprintf(`^(?:%s) *(?P<msg_id>\d+) *(?P<sig_name>\w+) *(?P<mux_name>\w+) *(?:\d+\-?){2} *;$`, symbols.DBCMuxValue),
)

type DBCReader struct {
	currLine int
}

func NewDBCReader() *DBCReader {
	return &DBCReader{
		currLine: 0,
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
		Messages: make(map[string]*Message),
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
						return nil, r.lineErr(fmt.Sprintf("signaal %s in message %d doesn't exist", sigComment.signalName, sigComment.messageID))
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

	return can, nil
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

	msgID, err := parseUint(match[dbcMuxSignalRegex.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[dbcMuxSignalRegex.SubexpIndex("sig_name")]
	muxName := match[dbcMuxSignalRegex.SubexpIndex("mux_name")]

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
