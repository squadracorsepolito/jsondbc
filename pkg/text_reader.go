package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func applyReg(r *regexp.Regexp, str string) ([]string, bool) {
	matches := r.FindAllStringSubmatch(str, -1)
	if len(matches) == 0 {
		return nil, false
	}
	return matches[0], true
}

type extMuxSignal struct {
	msgID      uint32
	sigName    string
	muxSigName string
}

type bitmapDefinition struct {
	msgID   uint32
	sigName string
	bitmap  map[string]uint32
}

type nodeComment struct {
	nodeName string
	desc     string
}

type msgComment struct {
	msgID uint32
	desc  string
}

type sigComment struct {
	msgID   uint32
	sigName string
	desc    string
}

type attributeDefault struct {
	attName string
	attData string
}

type nodeAttAss struct {
	attName  string
	attData  string
	nodeName string
}

type msgAttAss struct {
	attName string
	attData string
	msgID   uint32
}

type sigAttAss struct {
	attName string
	attData string
	msgID   uint32
	sigName string
}

type textReaderCfg struct {
	varsionIdent   string
	busSpeedIdent  string
	nodeIdent      string
	msgIdent       string
	sigIdent       string
	extMuxSigIdent string
	bitmapDefIdent string
	commentIdent   string
	attIdent       string
	attDefIdent    string
	attAssIdent    string
}

type textReader struct {
	cfg *textReaderCfg

	lines    []string
	canModel *CanModel

	versionReg *regexp.Regexp

	busSpeedReg *regexp.Regexp

	nodeReg *regexp.Regexp

	msgStartReg *regexp.Regexp
	msgEndReg   *regexp.Regexp

	sigReg *regexp.Regexp

	extMuxSigReg  *regexp.Regexp
	extMuxSignals map[uint32][]*extMuxSignal

	bitmapDefReg *regexp.Regexp

	nodeCommentReg *regexp.Regexp
	msgCommentReg  *regexp.Regexp
	sigCommentReg  *regexp.Regexp

	attReg    *regexp.Regexp
	attDefReg *regexp.Regexp

	nodeAttAssReg *regexp.Regexp
	msgAttAssReg  *regexp.Regexp
	sigAttAssReg  *regexp.Regexp
}

func (r *textReader) getError(lineNum int, errStr string) error {
	return fmt.Errorf("line %d: %s", lineNum+1, errStr)
}

func (r *textReader) read() (*CanModel, error) {
	canModel := &CanModel{}
	r.canModel = canModel

	msgIdxs := []int{}
	extMuxSigIdxs := []int{}
	bitmapDefIdxs := []int{}
	commentIdxs := []int{}
	attIdxs := []int{}
	attDefIdxs := []int{}
	attAssIdxs := []int{}

	for lineIdx, line := range r.lines {
		if strings.HasPrefix(line, r.cfg.varsionIdent) {
			version, err := r.readVersion(lineIdx)
			if err != nil {
				return nil, err
			}
			canModel.Version = version
			continue
		}
		if strings.HasPrefix(line, r.cfg.busSpeedIdent) {
			busSpeed, err := r.readBusSpeed(lineIdx)
			if err != nil {
				return nil, err
			}
			canModel.Baudrate = busSpeed
			continue
		}
		if strings.HasPrefix(line, r.cfg.nodeIdent) {
			nodes, err := r.readNodes(lineIdx)
			if err != nil {
				return nil, err
			}
			canModel.Nodes = nodes
			continue
		}
		if strings.HasPrefix(line, r.cfg.msgIdent) {
			msgIdxs = append(msgIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.extMuxSigIdent) {
			extMuxSigIdxs = append(extMuxSigIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.bitmapDefIdent) {
			bitmapDefIdxs = append(bitmapDefIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.commentIdent) {
			commentIdxs = append(commentIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.attDefIdent) {
			attDefIdxs = append(attDefIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.attIdent) {
			attIdxs = append(attIdxs, lineIdx)
			continue
		}
		if strings.HasPrefix(line, r.cfg.attAssIdent) {
			attAssIdxs = append(attAssIdxs, lineIdx)
			continue
		}
	}

	if err := r.handleExtMuxSignals(extMuxSigIdxs); err != nil {
		return nil, err
	}

	messages, err := r.handleMessages(msgIdxs)
	if err != nil {
		return nil, err
	}
	canModel.Messages = messages

	if err := r.handleBitmapDefinitions(bitmapDefIdxs); err != nil {
		return nil, err
	}

	if err := r.handleComments(commentIdxs); err != nil {
		return nil, err
	}

	if err := r.handleAttributes(attIdxs); err != nil {
		return nil, err
	}

	if err := r.handleAttributeDefaults(attDefIdxs); err != nil {
		return nil, err
	}

	if err := r.handleAttributeAssignments(attAssIdxs); err != nil {
		return nil, err
	}

	canModel.source = sourceTypeDBC

	return canModel, nil
}

func (r *textReader) readVersion(lineIdx int) (string, error) {
	match, ok := applyReg(r.versionReg, r.lines[lineIdx])
	if !ok {
		return "", r.getError(lineIdx, "invalid version syntax")
	}

	return match[r.versionReg.SubexpIndex("version")], nil
}

func (r *textReader) readBusSpeed(lineIdx int) (uint32, error) {
	match, ok := applyReg(r.busSpeedReg, r.lines[lineIdx])
	if !ok {
		return 0, r.getError(lineIdx, "invalid bus speed syntax")
	}

	if match[r.busSpeedReg.SubexpIndex("speed")] == "" {
		return 0, nil
	}

	speed, err := parseUint(match[r.busSpeedReg.SubexpIndex("speed")])
	if err != nil {
		return 0, err
	}

	return speed, nil
}

func (r *textReader) readNodes(lineIdx int) (map[string]*Node, error) {
	match, ok := applyReg(r.nodeReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid node syntax")
	}

	strNodes := strings.TrimSpace(match[r.nodeReg.SubexpIndex("nodes")])
	nodeNames := strings.Split(strNodes, " ")
	nodes := make(map[string]*Node, len(nodeNames))
	for _, name := range nodeNames {
		nodes[name] = &Node{
			AttributeAssignments: &AttributeAssignments{
				Attributes: make(map[string]any),
			},
		}
	}

	return nodes, nil
}

func (r *textReader) handleExtMuxSignals(lineIdxs []int) error {
	extMuxSignals := make(map[uint32][]*extMuxSignal)
	for _, lineIdx := range lineIdxs {
		sig, err := r.readExtMuxSignal(lineIdx)
		if err != nil {
			return r.getError(lineIdx, err.Error())
		}
		extMuxSignals[sig.msgID] = append(extMuxSignals[sig.msgID], sig)
	}
	r.extMuxSignals = extMuxSignals
	return nil
}

func (r *textReader) readExtMuxSignal(lineIdx int) (*extMuxSignal, error) {
	match, ok := applyReg(r.extMuxSigReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid extended multiplexed signal syntax")
	}

	msgID, err := parseUint(match[r.extMuxSigReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[r.extMuxSigReg.SubexpIndex("sig_name")]
	muxSigName := match[r.extMuxSigReg.SubexpIndex("mux_sig_name")]

	return &extMuxSignal{
		msgID:      msgID,
		sigName:    sigName,
		muxSigName: muxSigName,
	}, nil
}

func (r *textReader) handleMessages(lineIdxs []int) (map[string]*Message, error) {
	messages := make(map[string]*Message)
	for _, lineIdx := range lineIdxs {
		msg, err := r.readMessage(lineIdx)
		if err != nil {
			return nil, r.getError(lineIdx, err.Error())
		}
		messages[msg.messageName] = msg
	}
	return messages, nil
}

func (r *textReader) readMessage(lineIdx int) (*Message, error) {
	startMatch, ok := applyReg(r.msgStartReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid message syntax")
	}

	msgID, err := parseUint(startMatch[r.msgStartReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	name := startMatch[r.msgStartReg.SubexpIndex("msg_name")]
	length, err := parseUint(startMatch[r.msgStartReg.SubexpIndex("length")])
	if err != nil {
		return nil, err
	}
	sender := startMatch[r.msgStartReg.SubexpIndex("sender")]

	msg := &Message{
		ID:      msgID,
		Length:  length,
		Sender:  sender,
		Signals: make(map[string]*Signal),
		AttributeAssignments: &AttributeAssignments{
			Attributes: make(map[string]any),
		},

		messageName: name,
		fromDBC:     true,
	}

	splMuxSigs := make(map[string]*Signal)
	isExtMux := false
	extMuxSigs := make(map[string]*Signal)

	for nextLineIdx := lineIdx + 1; nextLineIdx < len(r.lines); nextLineIdx++ {
		_, ok := applyReg(r.msgEndReg, r.lines[nextLineIdx])
		if ok {
			break
		}

		sig, err := r.readSignal(nextLineIdx)
		if err != nil {
			return nil, r.getError(nextLineIdx, err.Error())
		}

		if !sig.isMultiplexed {
			msg.Signals[sig.signalName] = sig
			continue
		}

		if sig.isMultiplexed && !sig.isMultiplexor && !isExtMux {
			splMuxSigs[sig.signalName] = sig
			continue
		}

		if !isExtMux {
			isExtMux = true
			for sigName, splSig := range splMuxSigs {
				extMuxSigs[sigName] = splSig
				delete(splMuxSigs, sigName)
			}
		}

		extMuxSigs[sig.signalName] = sig
	}

	if len(splMuxSigs) > 0 {
		for _, sig := range msg.Signals {
			if !sig.isMultiplexor {
				continue
			}

			for splSigName, splSig := range splMuxSigs {
				sig.MuxGroup[splSigName] = splSig
			}
		}

		return msg, nil
	}

	if isExtMux {
		extMuxs, ok := r.extMuxSignals[msgID]
		if !ok {
			return nil, r.getError(lineIdx, "missing extended multiplexed signal definition")
		}

		for _, extMuxSig := range extMuxs {
			multiplexorSig := &Signal{}

			if tmpSig, ok := extMuxSigs[extMuxSig.muxSigName]; ok {
				multiplexorSig = tmpSig
			} else if tmpSig, ok := msg.Signals[extMuxSig.muxSigName]; ok {
				multiplexorSig = tmpSig
			}

			multiplexorSig.MuxGroup[extMuxSig.sigName] = extMuxSigs[extMuxSig.sigName]
		}
	}

	return msg, nil
}

func (r *textReader) readSignal(lineIdx int) (*Signal, error) {
	match, ok := applyReg(r.sigReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid signal syntax")
	}

	sigName := match[r.sigReg.SubexpIndex("sig_name")]

	strMuxSwitch := match[r.sigReg.SubexpIndex("mux_switch")]
	muxSwitch := uint32(0)
	isMultiplexed := strMuxSwitch != ""
	if isMultiplexed {
		tmpMuxSwitch, err := parseUint(strMuxSwitch[1:])
		if err != nil {
			return nil, err
		}
		muxSwitch = tmpMuxSwitch
	}
	isMultiplexor := match[r.sigReg.SubexpIndex("mux")] == "M"

	startBit, err := parseUint(match[r.sigReg.SubexpIndex("start_bit")])
	if err != nil {
		return nil, err
	}
	size, err := parseUint(match[r.sigReg.SubexpIndex("size")])
	if err != nil {
		return nil, err
	}

	endianness := "little"
	bigEndian := false
	if match[r.sigReg.SubexpIndex("order")] == "0" {
		bigEndian = true
		endianness = "big"
	}
	signed := false
	if match[r.sigReg.SubexpIndex("signed")] == "-" {
		signed = true
	}

	scale, err := parseFloat(match[r.sigReg.SubexpIndex("scale")])
	if err != nil {
		return nil, err
	}
	offset, err := parseFloat(match[r.sigReg.SubexpIndex("offset")])
	if err != nil {
		return nil, err
	}

	min, err := parseFloat(match[r.sigReg.SubexpIndex("min")])
	if err != nil {
		return nil, err
	}
	strMax := match[r.sigReg.SubexpIndex("max")]
	max, err := parseFloat(strMax)
	if err != nil {
		return nil, err
	}

	unit := match[r.sigReg.SubexpIndex("unit")]

	tmpReceivers := strings.Split(match[r.sigReg.SubexpIndex("receivers")], ",")
	receivers := []string{}
	for _, tmp := range tmpReceivers {
		if tmp != dbcDefNode {
			receivers = append(receivers, tmp)
		}
	}

	return &Signal{
		StartBit:   uint32(startBit),
		Size:       uint32(size),
		Signed:     signed,
		Unit:       unit,
		Endianness: endianness,
		Receivers:  receivers,
		Scale:      scale,
		Offset:     offset,
		Min:        min,
		Max:        max,
		Enum:       make(map[string]uint32),
		MuxGroup:   make(map[string]*Signal),
		MuxSwitch:  muxSwitch,
		AttributeAssignments: &AttributeAssignments{
			Attributes: make(map[string]any),
		},

		isBigEndian:   bigEndian,
		signalName:    sigName,
		isMultiplexor: isMultiplexor,
		isMultiplexed: isMultiplexed,
	}, nil
}

func (r *textReader) handleBitmapDefinitions(lineIdxs []int) error {
	bitmapDefinitions := []*bitmapDefinition{}
	for _, lineIdx := range lineIdxs {
		bitmapDef, err := r.readBitmapDefinition(lineIdx)
		if err != nil {
			return r.getError(lineIdx, err.Error())
		}
		bitmapDefinitions = append(bitmapDefinitions, bitmapDef)
	}

	for _, bitmapDef := range bitmapDefinitions {
		for _, msg := range r.canModel.Messages {
			if msg.ID != bitmapDef.msgID {
				continue
			}

			for _, sig := range msg.Signals {
				if r.setBitmapDefinitionRec(sig, bitmapDef) {
					break
				}
			}

			break
		}
	}

	return nil
}

func (r *textReader) setBitmapDefinitionRec(sig *Signal, bitmapDef *bitmapDefinition) bool {
	if sig.signalName == bitmapDef.sigName {
		sig.Enum = bitmapDef.bitmap
		return true
	}

	if sig.isMultiplexor {
		for _, muxedSig := range sig.MuxGroup {
			if r.setBitmapDefinitionRec(muxedSig, bitmapDef) {
				return true
			}
		}
	}

	return false
}

func (r *textReader) readBitmapDefinition(lineIdx int) (*bitmapDefinition, error) {
	match, err := applyRegex(r.bitmapDefReg, r.lines[lineIdx])
	if err != nil {
		return nil, err
	}

	msgID, err := parseUint(match[r.bitmapDefReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[r.bitmapDefReg.SubexpIndex("sig_name")]

	strBitmap := match[r.bitmapDefReg.SubexpIndex("bitmap")]
	splBitmap := strings.Split(strings.TrimSpace(strBitmap), " ")
	bitmap := make(map[string]uint32, len(splBitmap)/2)
	for i := 0; i < len(splBitmap); i = i + 2 {
		val, _ := strconv.ParseUint(splBitmap[i], 10, 32)
		name := strings.Replace(splBitmap[i+1], `"`, "", 2)
		bitmap[name] = uint32(val)
	}

	return &bitmapDefinition{
		msgID:   msgID,
		sigName: sigName,
		bitmap:  bitmap,
	}, nil
}

func (r *textReader) handleComments(lineIdxs []int) error {
	nodeComments := []*nodeComment{}
	msgComments := []*msgComment{}
	sigComments := []*sigComment{}

	for _, lineIdx := range lineIdxs {
		tmpLine := strings.TrimSpace(r.lines[lineIdx][len(r.cfg.commentIdent):])

		if strings.HasPrefix(tmpLine, r.cfg.nodeIdent) {
			comment, err := r.readNodeComment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}
			nodeComments = append(nodeComments, comment)
			continue
		}

		if strings.HasPrefix(tmpLine, r.cfg.msgIdent) {
			comment, err := r.readMsgComment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}
			msgComments = append(msgComments, comment)
			continue
		}

		if strings.HasPrefix(tmpLine, r.cfg.sigIdent) {
			comment, err := r.readSigComment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}
			sigComments = append(sigComments, comment)
			continue
		}
	}

	for _, nodeComment := range nodeComments {
		node, ok := r.canModel.Nodes[nodeComment.nodeName]
		if !ok {
			continue
		}
		node.Description = nodeComment.desc
	}

	for _, msgComment := range msgComments {
		for _, msg := range r.canModel.Messages {
			if msg.ID != msgComment.msgID {
				continue
			}
			msg.Description = msgComment.desc
			break
		}
	}

	for _, sigComment := range sigComments {
		for _, msg := range r.canModel.Messages {
			if msg.ID != sigComment.msgID {
				continue
			}

			found := false
			for _, sig := range msg.Signals {
				if r.setSignalCommentRec(sig, sigComment) {
					found = true
					break
				}
			}

			if found {
				break
			}
		}
	}

	return nil
}

func (r *textReader) setSignalCommentRec(sig *Signal, comment *sigComment) bool {
	if sig.signalName == comment.sigName {
		sig.Description = comment.desc
		return true
	}

	if sig.isMultiplexor {
		for _, muxedSig := range sig.MuxGroup {
			if r.setSignalCommentRec(muxedSig, comment) {
				return true
			}
		}
	}

	return false
}

func (r *textReader) readNodeComment(lineIdx int) (*nodeComment, error) {
	match, ok := applyReg(r.nodeCommentReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid node comment syntax")
	}

	nodeName := match[r.nodeCommentReg.SubexpIndex("node_name")]
	desc := match[r.nodeCommentReg.SubexpIndex("desc")]

	return &nodeComment{
		nodeName: nodeName,
		desc:     desc,
	}, nil
}

func (r *textReader) readMsgComment(lineIdx int) (*msgComment, error) {
	match, ok := applyReg(r.msgCommentReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid message comment syntax")
	}

	msgID, err := parseUint(match[r.msgCommentReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	desc := match[r.msgCommentReg.SubexpIndex("desc")]

	return &msgComment{
		msgID: msgID,
		desc:  desc,
	}, nil
}

func (r *textReader) readSigComment(lineIdx int) (*sigComment, error) {
	match, ok := applyReg(r.sigCommentReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid signal comment syntax")
	}

	msgID, err := parseUint(match[r.sigCommentReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[r.sigCommentReg.SubexpIndex("sig_name")]
	desc := match[r.sigCommentReg.SubexpIndex("desc")]

	return &sigComment{
		msgID:   msgID,
		sigName: sigName,
		desc:    desc,
	}, nil
}

func (r *textReader) handleAttributes(lineIdxs []int) error {
	genAtt := make(map[string]*Attribute)
	nodeAtt := make(map[string]*NodeAttribute)
	msgAtt := make(map[string]*MessageAttribute)
	sigAtt := make(map[string]*SignalAttribute)
	for _, lineIdx := range lineIdxs {
		att, err := r.readAttribute(lineIdx)
		if err != nil {
			return r.getError(lineIdx, err.Error())
		}

		switch att.attributeKind {
		case attributeKindGeneral:
			genAtt[att.attributeName] = att
		case attributeKindNode:
			nodeAtt[att.attributeName] = &NodeAttribute{Attribute: att}
		case attributeKindMessage:
			msgAtt[att.attributeName] = &MessageAttribute{Attribute: att}
		case attributeKindSignal:
			sigAtt[att.attributeName] = &SignalAttribute{Attribute: att}
		}
	}

	r.canModel.GeneralAttributes = genAtt
	r.canModel.NodeAttributes = nodeAtt
	r.canModel.MessageAttributes = msgAtt
	r.canModel.SignalAttributes = sigAtt

	return nil
}

func (r *textReader) readAttribute(lineIdx int) (*Attribute, error) {
	match, ok := applyReg(r.attReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid attribute syntax")
	}

	att := &Attribute{}

	attKind := match[r.attReg.SubexpIndex("att_kind")]
	switch attKind {
	case r.cfg.nodeIdent:
		att.attributeKind = attributeKindNode
	case r.cfg.msgIdent:
		att.attributeKind = attributeKindMessage
	case r.cfg.sigIdent:
		att.attributeKind = attributeKindSignal
	default:
		att.attributeKind = attributeKindGeneral
	}

	att.attributeName = match[r.attReg.SubexpIndex("att_name")]

	attType := match[r.attReg.SubexpIndex("att_type")]
	attData := strings.TrimSpace(match[r.attReg.SubexpIndex("att_data")])

	switch attType {
	case "INT":
		att.attributeType = attributeTypeInt
		att.Int = &AttributeInt{}
		_, err := fmt.Sscanf(attData, "%d %d", &att.Int.From, &att.Int.To)
		if err != nil {
			return nil, err
		}

	case "STRING":
		att.attributeType = attributeTypeString
		att.String = &AttributeString{}

	case "FLOAT":
		att.attributeType = attributeTypeFloat
		att.Float = &AttributeFloat{}
		_, err := fmt.Sscanf(attData, "%f %f", &att.Float.From, &att.Float.To)
		if err != nil {
			return nil, err
		}

	case "ENUM":
		att.attributeType = attributeTypeEnum
		att.Enum = &AttributeEnum{}
		splData := strings.Split(attData, ",")
		for _, val := range splData {
			att.Enum.Values = append(att.Enum.Values, strings.Replace(val, `"`, "", 2))
		}
	}

	return att, nil
}

func (r *textReader) handleAttributeDefaults(lineIdxs []int) error {
	for _, lineIdx := range lineIdxs {
		attDef, err := r.readAttributeDefaults(lineIdx)
		if err != nil {
			return r.getError(lineIdx, err.Error())
		}

		if att, ok := r.canModel.GeneralAttributes[attDef.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				val, err := parseInt(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Int.Default = val

			case attributeTypeString:
				att.String.Default = strings.Replace(attDef.attData, `"`, "", 2)

			case attributeTypeFloat:
				val, err := parseFloat(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Float.Default = val

			case attributeTypeEnum:
				strVal := strings.Replace(attDef.attData, `"`, "", 2)
				idx := 0
				for i, val := range att.Enum.Values {
					if val == strVal {
						idx = i
						break
					}
				}
				att.Enum.Default = att.Enum.Values[idx]
				att.Enum.defaultIdx = idx
			}
			continue
		}
		if att, ok := r.canModel.NodeAttributes[attDef.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				val, err := parseInt(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Int.Default = val

			case attributeTypeString:
				att.String.Default = strings.Replace(attDef.attData, `"`, "", 2)

			case attributeTypeFloat:
				val, err := parseFloat(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Float.Default = val

			case attributeTypeEnum:
				strVal := strings.Replace(attDef.attData, `"`, "", 2)
				idx := 0
				for i, val := range att.Enum.Values {
					if val == strVal {
						idx = i
						break
					}
				}
				att.Enum.Default = att.Enum.Values[idx]
				att.Enum.defaultIdx = idx

			}
			continue
		}
		if att, ok := r.canModel.MessageAttributes[attDef.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				val, err := parseInt(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Int.Default = val

			case attributeTypeString:
				att.String.Default = strings.Replace(attDef.attData, `"`, "", 2)

			case attributeTypeFloat:
				val, err := parseFloat(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Float.Default = val

			case attributeTypeEnum:
				strVal := strings.Replace(attDef.attData, `"`, "", 2)
				idx := 0
				for i, val := range att.Enum.Values {
					if val == strVal {
						idx = i
						break
					}
				}
				att.Enum.Default = att.Enum.Values[idx]
				att.Enum.defaultIdx = idx

			}
			continue
		}
		if att, ok := r.canModel.SignalAttributes[attDef.attName]; ok {
			switch att.attributeType {
			case attributeTypeInt:
				val, err := parseInt(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Int.Default = val

			case attributeTypeString:
				att.String.Default = strings.Replace(attDef.attData, `"`, "", 2)

			case attributeTypeFloat:
				val, err := parseFloat(attDef.attData)
				if err != nil {
					return r.getError(lineIdx, err.Error())
				}
				att.Float.Default = val

			case attributeTypeEnum:
				strVal := strings.Replace(attDef.attData, `"`, "", 2)
				idx := 0
				for i, val := range att.Enum.Values {
					if val == strVal {
						idx = i
						break
					}
				}
				att.Enum.Default = att.Enum.Values[idx]
				att.Enum.defaultIdx = idx

			}
		}
	}

	return nil
}

func (r *textReader) readAttributeDefaults(lineIdx int) (*attributeDefault, error) {
	match, ok := applyReg(r.attDefReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid attribute default syntax")
	}

	attName := match[r.attDefReg.SubexpIndex("att_name")]
	attData := strings.TrimSpace(match[r.attDefReg.SubexpIndex("att_data")])

	return &attributeDefault{
		attName: attName,
		attData: attData,
	}, nil
}

func (r *textReader) getAttributeValue(attType attributeType, attData string, enumValues []string) (any, error) {
	data := strings.TrimSpace(attData)

	switch attType {
	case attributeTypeInt:
		val, err := parseInt(data)
		if err != nil {
			return nil, err
		}
		return val, nil

	case attributeTypeString:
		return strings.Replace(data, `"`, "", 2), nil

	case attributeTypeFloat:
		val, err := parseFloat(data)
		if err != nil {
			return nil, err
		}
		return val, nil

	case attributeTypeEnum:
		idx, err := parseInt(data)
		if err != nil {
			return nil, err
		}
		return enumValues[idx], nil
	}

	return nil, nil
}

func (r *textReader) handleAttributeAssignments(lineIdxs []int) error {
	for _, lineIdx := range lineIdxs {
		tmpLine := strings.Split(strings.TrimSpace(r.lines[lineIdx]), " ")[2]
		if strings.HasPrefix(tmpLine, r.cfg.nodeIdent) {
			nodeAss, err := r.readNodeAttributeAssignment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}

			if node, ok := r.canModel.Nodes[nodeAss.nodeName]; ok {
				if att, ok := r.canModel.NodeAttributes[nodeAss.attName]; ok {
					enumValues := []string{}
					if att.attributeType == attributeTypeEnum {
						enumValues = att.Enum.Values
					}
					val, err := r.getAttributeValue(att.attributeType, nodeAss.attData, enumValues)
					if err != nil {
						return r.getError(lineIdx, err.Error())
					}
					node.Attributes[nodeAss.attName] = val
				}
			}

			continue
		}

		if strings.HasPrefix(tmpLine, r.cfg.msgIdent) {
			msgAss, err := r.readMessageAttributeAssignment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}

			for _, msg := range r.canModel.Messages {
				if msg.ID != msgAss.msgID {
					continue
				}

				if att, ok := r.canModel.MessageAttributes[msgAss.attName]; ok {
					enumValues := []string{}
					if att.attributeType == attributeTypeEnum {
						enumValues = att.Enum.Values
					}
					val, err := r.getAttributeValue(att.attributeType, msgAss.attData, enumValues)
					if err != nil {
						return r.getError(lineIdx, err.Error())
					}
					msg.Attributes[msgAss.attName] = val
				}
			}

			continue
		}

		if strings.HasPrefix(tmpLine, r.cfg.sigIdent) {
			sigAss, err := r.readSignalAttributeAssignment(lineIdx)
			if err != nil {
				return r.getError(lineIdx, err.Error())
			}

			for _, msg := range r.canModel.Messages {
				if msg.ID != sigAss.msgID {
					continue
				}

				for _, sig := range msg.Signals {
					found, err := r.setSignalAttributeAssignmentRec(sig, sigAss)
					if err != nil {
						return r.getError(lineIdx, err.Error())
					}
					if found {
						break
					}
				}
			}
		}
	}

	return nil
}

func (r *textReader) setSignalAttributeAssignmentRec(sig *Signal, sigAss *sigAttAss) (bool, error) {
	if sig.signalName == sigAss.sigName {
		if att, ok := r.canModel.SignalAttributes[sigAss.attName]; ok {
			enumValues := []string{}
			if att.attributeType == attributeTypeEnum {
				enumValues = att.Enum.Values
			}
			val, err := r.getAttributeValue(att.attributeType, sigAss.attData, enumValues)
			if err != nil {
				return false, err
			}
			sig.Attributes[sigAss.attName] = val
		}

		return true, nil
	}

	if sig.isMultiplexor {
		for _, muxedSig := range sig.MuxGroup {
			found, err := r.setSignalAttributeAssignmentRec(muxedSig, sigAss)
			if err != nil {
				return false, err
			}
			if found {
				return true, nil
			}
		}
	}

	return false, nil
}

func (r *textReader) readNodeAttributeAssignment(lineIdx int) (*nodeAttAss, error) {
	match, ok := applyReg(r.nodeAttAssReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid node attribute assignment syntax")
	}

	attName := match[r.nodeAttAssReg.SubexpIndex("att_name")]
	nodeName := match[r.nodeAttAssReg.SubexpIndex("node_name")]
	attData := strings.TrimSpace(match[r.nodeAttAssReg.SubexpIndex("att_data")])

	return &nodeAttAss{
		nodeName: nodeName,
		attName:  attName,
		attData:  attData,
	}, nil
}

func (r *textReader) readMessageAttributeAssignment(lineIdx int) (*msgAttAss, error) {
	match, ok := applyReg(r.msgAttAssReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid message attribute assignment syntax")
	}

	attName := match[r.msgAttAssReg.SubexpIndex("att_name")]
	attData := match[r.msgAttAssReg.SubexpIndex("att_data")]
	msgID, err := parseUint(match[r.msgAttAssReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}

	return &msgAttAss{
		attName: attName,
		attData: attData,
		msgID:   msgID,
	}, nil
}

func (r *textReader) readSignalAttributeAssignment(lineIdx int) (*sigAttAss, error) {
	match, ok := applyReg(r.sigAttAssReg, r.lines[lineIdx])
	if !ok {
		return nil, r.getError(lineIdx, "invalid signal attribute assignment syntax")
	}

	attName := match[r.sigAttAssReg.SubexpIndex("att_name")]
	attData := match[r.sigAttAssReg.SubexpIndex("att_data")]
	msgID, err := parseUint(match[r.sigAttAssReg.SubexpIndex("msg_id")])
	if err != nil {
		return nil, err
	}
	sigName := match[r.sigAttAssReg.SubexpIndex("sig_name")]

	return &sigAttAss{
		attName: attName,
		attData: attData,
		msgID:   msgID,
		sigName: sigName,
	}, nil
}
