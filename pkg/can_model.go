package pkg

import (
	"fmt"
	"os"

	"github.com/FerroO2000/jsondbc/pkg/sym"
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
	Baudrate          uint32                       `json:"baudrate"`
	Nodes             map[string]*Node             `json:"nodes"`
	GeneralAttributes map[string]*Attribute        `json:"general_attributes"`
	NodeAttributes    map[string]*NodeAttribute    `json:"node_attributes"`
	MessageAttributes map[string]*MessageAttribute `json:"message_attributes"`
	SignalAttributes  map[string]*SignalAttribute  `json:"signal_attributes"`
	Messages          map[string]*Message          `json:"messages"`
}

func (c *CanModel) Init() {
	for attName, att := range c.GeneralAttributes {
		att.initAttribute(attName)
	}
	for attName, att := range c.NodeAttributes {
		att.initNodeAttribute(attName)
	}

	_, hasFreqAtt := c.MessageAttributes[sym.MsgFrequencyAttribute]
	if hasFreqAtt {
		delete(c.MessageAttributes, sym.MsgFrequencyAttribute)
	} else {
		if c.MessageAttributes == nil {
			c.MessageAttributes = make(map[string]*MessageAttribute)
		}
		c.MessageAttributes[sym.MsgFrequencyAttribute] = &MessageAttribute{
			Attribute: &Attribute{
				Int: &AttributeInt{
					Default: 0,
					From:    0,
					To:      65535,
				},
			},
		}
	}
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
		msg.initMessage(msgName)
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
			}
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
