package cangoru

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/squadracorsepolito/jsondbc/pkg/cangoru/dbc"
)

func NewCANFromDBC(dbcFilename string) (*CAN, error) {
	ext := path.Ext(dbcFilename)
	if ext != ".dbc" {
		return nil, fmt.Errorf("file %s: extension must be .dbc; got %s", dbcFilename, ext)
	}

	file, err := os.ReadFile(dbcFilename)
	if err != nil {
		return nil, err
	}

	fileMime := http.DetectContentType(file)
	if fileMime != "text/plain; charset=utf-8" {
		return nil, fmt.Errorf("file %s: content type must be text/plain; got %s", dbcFilename, fileMime)
	}

	parser := dbc.NewParser(file)
	dbcAST, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	can := NewCAN()

	can.SetVersionString(dbcAST.Version)

	can.SetBaudrate(uint(dbcAST.BitTiming.Baudrate))

	for _, nodeName := range dbcAST.Nodes.Names {
		if err := can.AddNode(NewNode(nodeName)); err != nil {
			return nil, err
		}
	}

	for _, dbcMsg := range dbcAST.Messages {
		if err := can.handleDBCMessage(dbcMsg); err != nil {
			return nil, err
		}
	}

	for _, valEnc := range dbcAST.ValueEncodings {
		if err := can.handleDBCValueEncoding(valEnc); err != nil {
			return nil, err
		}
	}

	for _, dbcCom := range dbcAST.Comments {
		if err := can.handleDBCComment(dbcCom); err != nil {
			return nil, err
		}
	}

	for _, dbcAtt := range dbcAST.Attributes {
		if err := can.handleDBCAttribute(dbcAtt); err != nil {
			return nil, err
		}
	}

	for _, dbcAttDef := range dbcAST.AttributeDefaults {
		if err := can.handleDBCAttributeDefault(dbcAttDef); err != nil {
			return nil, err
		}
	}

	for _, dbcAttVal := range dbcAST.AttributeValues {
		if err := can.handleDBCAttributeValue(dbcAttVal); err != nil {
			return nil, err
		}
	}

	for _, dbcExtMux := range dbcAST.ExtendedMuxes {
		if err := can.handleDBCExtendedMux(dbcExtMux); err != nil {
			return nil, err
		}
	}

	return can, nil
}

func (c *CAN) handleDBCMessage(dbcMsg *dbc.Message) error {
	msg := NewMessage(NewMessageID(dbcMsg.ID), dbcMsg.Name, uint(dbcMsg.Size))

	if err := c.AddMessage(msg); err != nil {
		return err
	}

	if dbcMsg.Transmitter != dbc.DummyNode {
		node, err := c.GetNode(dbcMsg.Transmitter)
		if err != nil {
			return err
		}
		node.AddTxMessage(msg)
	}

	for _, dbcSig := range dbcMsg.Signals {
		if err := c.handleDBCSignal(msg, dbcSig); err != nil {
			return err
		}
	}

	if msg.HasMuxSignals() && !msg.HasExtendedMuxSignals() {
		for _, muxor := range msg.multiplexorSignals {
			for _, dbcSig := range dbcMsg.Signals {
				if dbcSig.IsMultiplexed {
					sig, err := msg.GetSignal(dbcSig.Name)
					if err != nil {
						return err
					}
					if err := muxor.AddMuxSignal(sig, uint(dbcSig.MuxSwitchValue)); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *CAN) handleDBCSignal(msg *Message, dbcSig *dbc.Signal) error {
	byteOrd := LittleEndian
	if dbcSig.ByteOrder == dbc.SignalBigEndian {
		byteOrd = BigEndian
	}

	valTyp := Unsigned
	if dbcSig.ValueType == dbc.SignalSigned {
		valTyp = Signed
	}

	sig, err := NewSignal(dbcSig.Name, uint(dbcSig.Size), uint(dbcSig.StartBit), byteOrd, valTyp,
		dbcSig.Factor, dbcSig.Offset, dbcSig.Min, dbcSig.Max, dbcSig.Unit)

	if err != nil {
		return err
	}

	if dbcSig.IsMultiplexor {
		sig.SetIsMultiplexor()
	}
	if dbcSig.IsMultiplexed {
		sig.SetIsMultiplexed()
	}

	if err := msg.AddSignal(sig); err != nil {
		return err
	}

	for _, rxName := range dbcSig.Receivers {
		if rxName != dbc.DummyNode {
			node, err := c.GetNode(rxName)
			if err != nil {
				return err
			}
			node.AddRxSignal(sig)
		}
	}

	return nil
}

func (c *CAN) handleDBCValueEncoding(dbcValEnc *dbc.ValueEncoding) error {
	if dbcValEnc.Kind != dbc.ValueEncodingSignal {
		return nil
	}

	msg, err := c.GetMessage(NewMessageID(dbcValEnc.MessageID))
	if err != nil {
		return err
	}

	sig, err := msg.GetSignal(dbcValEnc.SignalName)
	if err != nil {
		return err
	}

	log.Print(dbcValEnc)

	for _, val := range dbcValEnc.Values {
		if err := sig.AddMapValue(uint(val.ID), val.Name); err != nil {
			return err
		}
	}

	return nil
}

func (c *CAN) handleDBCComment(dbcCom *dbc.Comment) error {
	switch dbcCom.Kind {
	case dbc.CommentGeneral:
		c.SetDescription(dbcCom.Text)

	case dbc.CommentNode:
		node, err := c.GetNode(dbcCom.NodeName)
		if err != nil {
			return err
		}
		node.SetDescription(dbcCom.Text)

	case dbc.CommentMessage:
		msg, err := c.GetMessage(NewMessageID(dbcCom.MessageID))
		if err != nil {
			return err
		}
		msg.SetDescription(dbcCom.Text)

	case dbc.CommentSignal:
		msg, err := c.GetMessage(NewMessageID(dbcCom.MessageID))
		if err != nil {
			return err
		}
		sig, err := msg.GetSignal(dbcCom.SignalName)
		if err != nil {
			return err
		}
		sig.SetDescription(dbcCom.Text)
	}

	return nil
}

func (c *CAN) handleDBCAttribute(dbcAtt *dbc.Attribute) error {
	kind := AttributeKindGeneral
	switch dbcAtt.Kind {
	case dbc.AttributeNode:
		kind = AttributeKindNode
	case dbc.AttributeMessage:
		kind = AttributeKindMessage
	case dbc.AttributeSignal:
		kind = AttributeKindSignal
	}

	switch dbcAtt.Type {
	case dbc.AttributeInt:
		if err := c.AddAttribute(NewIntAttribute(kind, dbcAtt.Name, dbcAtt.MinInt, dbcAtt.MaxInt)); err != nil {
			return err
		}

	case dbc.AttributeFloat:
		if err := c.AddAttribute(NewFloatAttribute(kind, dbcAtt.Name, dbcAtt.MinFloat, dbcAtt.MaxFloat)); err != nil {
			return err
		}

	case dbc.AttributeString:
		if err := c.AddAttribute(NewStringAttribute(kind, dbcAtt.Name)); err != nil {
			return err
		}

	case dbc.AttributeHex:
		if err := c.AddAttribute(NewHexAttribute(kind, dbcAtt.Name, dbcAtt.MinHex, dbcAtt.MaxHex)); err != nil {
			return err
		}

	case dbc.AttributeEnum:
		if err := c.AddAttribute(NewEnumAttribute(kind, dbcAtt.Name, dbcAtt.EnumValues)); err != nil {
			return err
		}
	}

	return nil
}

func (c *CAN) handleDBCAttributeDefault(dbcAttDef *dbc.AttributeDefault) error {
	att, err := c.GetAttribute(dbcAttDef.AttributeName)
	if err != nil {
		return err
	}

	switch att.Type {
	case AttributeTypeInt:
		if err := att.SetIntDefault(dbcAttDef.ValueInt); err != nil {
			return err
		}

	case AttributeTypeEnum:
		if err := att.SetEnumDefault(dbcAttDef.ValueInt); err != nil {
			return err
		}

	case AttributeTypeHex:
		if err := att.SetHexDefault(dbcAttDef.ValueHex); err != nil {
			return err
		}

	case AttributeTypeFloat:
		if err := att.SetFloatDefault(dbcAttDef.ValueFloat); err != nil {
			return err
		}

	case AttributeTypeString:
		if err := att.SetStringDefault(dbcAttDef.ValueString); err != nil {
			return err
		}
	}

	return nil
}

func (c *CAN) handleDBCAttributeValue(dbcAttVal *dbc.AttributeValue) error {
	att, err := c.GetAttribute(dbcAttVal.AttributeName)
	if err != nil {
		return err
	}

	var attVal *AttributeValue
	switch att.Type {
	case AttributeTypeInt:
		attVal, err = NewIntAttributeValue(att, dbcAttVal.ValueInt)
		if err != nil {
			return err
		}

	case AttributeTypeString:
		log.Print("+ ", dbcAttVal.ValueString)
		attVal, err = NewStringAttributeValue(att, dbcAttVal.ValueString)
		if err != nil {
			return err
		}

	case AttributeTypeFloat:
		attVal, err = NewFloatAttributeValue(att, dbcAttVal.ValueFloat)
		if err != nil {
			return err
		}

	case AttributeTypeHex:
		attVal, err = NewHexAttributeValue(att, dbcAttVal.ValueHex)
		if err != nil {
			return err
		}

	case AttributeTypeEnum:
		attVal, err = NewEnumAttributeValue(att, dbcAttVal.ValueInt)
		if err != nil {
			return err
		}
	}

	switch dbcAttVal.AttributeKind {
	case dbc.AttributeGeneral:
		c.AssignAttribute(attVal)

	case dbc.AttributeNode:
		node, err := c.GetNode(dbcAttVal.NodeName)
		if err != nil {
			return err
		}
		node.AssignAttribute(attVal)

	case dbc.AttributeMessage:
		msg, err := c.GetMessage(NewMessageID(dbcAttVal.MessageID))
		if err != nil {
			return err
		}

		switch dbcAttVal.AttributeName {
		case string(dbc.MsgPeriodMS):
			msg.SetPeriod(uint(attVal.IntValue))
		}

		msg.AssignAttribute(attVal)

	case dbc.AttributeSignal:
		msg, err := c.GetMessage(NewMessageID(dbcAttVal.MessageID))
		if err != nil {
			return err
		}
		sig, err := msg.GetSignal(dbcAttVal.SignalName)
		if err != nil {
			return err
		}
		sig.AssignAttribute(attVal)

	}

	return nil
}

func (c *CAN) handleDBCExtendedMux(extMux *dbc.ExtendedMux) error {
	msg, err := c.GetMessage(NewMessageID(extMux.MessageID))
	if err != nil {
		return err
	}

	muxor, ok := msg.multiplexorSignals[extMux.MultiplexorName]
	if !ok {
		return fmt.Errorf("message %s: multiplexor %s not found", msg.Name, extMux.MultiplexorName)
	}

	mux, ok := msg.multiplexedSignals[extMux.MultiplexedName]
	if !ok {
		return fmt.Errorf("message %s: multiplexed %s not found", msg.Name, extMux.MultiplexedName)
	}

	indexes := []uint{}
	for _, r := range extMux.Ranges {
		for i := r.From; i <= r.To; i++ {
			indexes = append(indexes, uint(i))
		}
	}

	if err := muxor.AddMuxSignal(mux, indexes...); err != nil {
		return err
	}

	return nil
}

func (c *CAN) ToDBC(dbcFilename string) error {
	writer := dbc.NewWriter()

	wFile, err := os.Create(dbcFilename)
	if err != nil {
		return err
	}
	defer wFile.Close()

	_, err = wFile.WriteString(writer.Write(c.tmp))
	return err
}
