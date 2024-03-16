package pkg

import (
	"fmt"
	"os"
	"sort"

	"github.com/squadracorsepolito/jsondbc/pkg/sym"
)

type sourceType int

const (
	sourceTypeJSON sourceType = iota
	sourceTypeDBC
)

type Reader interface {
	Read(file *os.File) (*CanModel, error)
}

type Writer interface {
	Write(file *os.File, canModel *CanModel) error
}

// CanModel represents the CAN model.
type CanModel struct {
	Version           string                       `json:"version"`
	Baudrate          uint32                       `json:"baudrate,omitempty"`
	Nodes             map[string]*Node             `json:"nodes"`
	GeneralAttributes map[string]*Attribute        `json:"general_attributes"`
	NodeAttributes    map[string]*NodeAttribute    `json:"node_attributes"`
	MessageAttributes map[string]*MessageAttribute `json:"message_attributes"`
	SignalAttributes  map[string]*SignalAttribute  `json:"signal_attributes"`
	Messages          map[string]*Message          `json:"messages"`
	SignalEnums       map[string]map[string]uint32 `json:"signal_enums"`

	source sourceType
}

func (c *CanModel) Init() {
	baudrateAtt, hasBaudrateAtt := c.GeneralAttributes[sym.BaudrateAttribute]
	if hasBaudrateAtt {
		c.Baudrate = uint32(baudrateAtt.Int.Default)
		delete(c.GeneralAttributes, sym.BaudrateAttribute)
	} else {
		if c.Baudrate > 0 {
			if c.GeneralAttributes == nil {
				c.GeneralAttributes = make(map[string]*Attribute)
			}
			c.GeneralAttributes[sym.BaudrateAttribute] = &Attribute{
				Int: &AttributeInt{
					Default: int(c.Baudrate),
					From:    0,
					To:      int(c.Baudrate),
				},
			}
		}
	}

	for attName, att := range c.GeneralAttributes {
		att.initAttribute(attName)
	}
	for attName, att := range c.NodeAttributes {
		att.initNodeAttribute(attName)
	}

	_, hasFreqAtt := c.MessageAttributes[sym.MsgPeriodAttribute]
	if hasFreqAtt {
		delete(c.MessageAttributes, sym.MsgPeriodAttribute)
	} else {
		if c.MessageAttributes == nil {
			c.MessageAttributes = make(map[string]*MessageAttribute)
		}
		c.MessageAttributes[sym.MsgPeriodAttribute] = &MessageAttribute{
			Attribute: &Attribute{
				Int: &AttributeInt{
					Default: 0,
					From:    0,
					To:      65535,
				},
			},
		}
	}

	if c.MessageAttributes == nil {
		c.MessageAttributes = make(map[string]*MessageAttribute)
	}

	if c.SignalAttributes == nil {
		c.SignalAttributes = make(map[string]*SignalAttribute)
	}

	c.handleCustomAttributes()

	for attName, att := range c.MessageAttributes {
		att.initMessageAttribute(attName)
	}

	for attName, att := range c.SignalAttributes {
		att.initSignalAttribute(attName)
	}

	for nodeName, node := range c.Nodes {
		node.initNode(nodeName)
	}

	for msgName, msg := range c.Messages {
		msg.initMessage(msgName, c.source)
	}

	for _, node := range c.Nodes {
		for attName := range node.Attributes {
			if nodeAtt, ok := c.NodeAttributes[attName]; ok {
				nodeAtt.assignNode(node)
			}
		}
	}
	for _, msg := range c.Messages {
		for attName := range msg.Attributes {
			if msgAtt, ok := c.MessageAttributes[attName]; ok {
				msgAtt.assignMessage(msg)
			}
		}

		for _, sig := range msg.childSignals {
			for sigName := range sig.Attributes {
				if sigAtt, ok := c.SignalAttributes[sigName]; ok {
					sigAtt.assignSignal(msg.ID, sig)
				}

				// check if signal has an enum_ref, if so, attach the global signal_enum
				if len(sig.Enum) == 0 && len(sig.EnumRef) > 0 {
					if enum, ok := c.SignalEnums[sig.EnumRef]; ok {
						sig.Enum = enum
					} else {
						fmt.Printf("WARNING: signal '%s' -> enum_ref '%s' is not defined in global signal enums -> SKIPPED\n", sig.signalName, sig.EnumRef)
					}
				}
			}
		}
	}
}

func (c *CanModel) handleCustomAttributes() {
	switch c.source {
	case sourceTypeJSON:
		// MsgCycleTime
		c.MessageAttributes[sym.MsgCycleTime] = &MessageAttribute{
			Attribute: &Attribute{
				Int: &AttributeInt{0, 0, 1000},
			},
		}

		// MsgSendType
		c.MessageAttributes[sym.MsgSendType] = &MessageAttribute{
			Attribute: &Attribute{
				Enum: &AttributeEnum{
					Default: sym.MsgSendTypeValues[0],
					Values:  sym.MsgSendTypeValues,
				},
			},
		}

		// SigSendType
		c.SignalAttributes[sym.SigSendType] = &SignalAttribute{
			Attribute: &Attribute{
				Enum: &AttributeEnum{
					Default: sym.SigSendTypeValues[0],
					Values:  sym.SigSendTypeValues,
				},
			},
		}

	case sourceTypeDBC:
		// MsgCycleTime
		if _, ok := c.MessageAttributes[sym.MsgCycleTime]; ok {
			delete(c.MessageAttributes, sym.MsgCycleTime)
		}

		// MsgSendType
		if _, ok := c.MessageAttributes[sym.MsgSendType]; ok {
			delete(c.MessageAttributes, sym.MsgSendType)
		}

		// SigSendType
		if _, ok := c.SignalAttributes[sym.SigSendType]; ok {
			delete(c.SignalAttributes, sym.SigSendType)
		}
	}
}

// Validate validates the CAN model.
func (c *CanModel) Validate() error {
	msgIDMap := make(map[uint32]string)
	for _, msg := range c.Messages {
		if msgName, ok := msgIDMap[msg.ID]; ok {
			return fmt.Errorf("[%s] message id [%d] is already taken by [%s]", msg.messageName, msg.ID, msgName)
		}
		msgIDMap[msg.ID] = msg.messageName
	}

	for _, msg := range c.Messages {
		if err := msg.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *CanModel) getAttributes() []*Attribute {
	attributes := []*Attribute{}

	for _, att := range c.GeneralAttributes {
		attributes = append(attributes, att)
	}
	for _, att := range c.NodeAttributes {
		attributes = append(attributes, att.asAttribute())
	}
	for _, att := range c.MessageAttributes {
		attributes = append(attributes, att.asAttribute())
	}
	for _, att := range c.SignalAttributes {
		attributes = append(attributes, att.asAttribute())
	}

	return attributes
}

func (c *CanModel) getNodeAttributes() []*NodeAttribute {
	attributes := []*NodeAttribute{}
	for _, att := range c.NodeAttributes {
		attributes = append(attributes, att)
	}
	return attributes
}

func (c *CanModel) getMessageAttributes() []*MessageAttribute {
	attributes := []*MessageAttribute{}
	for _, att := range c.MessageAttributes {
		attributes = append(attributes, att)
	}
	return attributes
}

func (c *CanModel) getSignalAttributes() []*SignalAttribute {
	attributes := []*SignalAttribute{}
	for _, att := range c.SignalAttributes {
		attributes = append(attributes, att)
	}
	return attributes
}

func (c *CanModel) getMessages() []*Message {
	messages := make([]*Message, len(c.Messages))

	idx := 0
	for _, msg := range c.Messages {
		messages[idx] = msg
		idx++
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	return messages
}
