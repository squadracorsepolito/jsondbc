package pkg

import (
	"os"
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
	BoundRate         uint32                       `json:"bound_rate"`
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

/*
type nodeAttAssignment struct {
	attAssignmentVal
	nodeName string
}

func (c *CanModel) getNodeAttAssignments() []*nodeAttAssignment {
	assignments := []*nodeAttAssignment{}

	for _, node := range c.Nodes {
		for _, ass := range node.getAttAssignmentValues() {
			assignments = append(assignments, &nodeAttAssignment{
				attAssignmentVal: *ass,
				nodeName:         node.name,
			})
		}
	}

	return assignments
}

type messageAttAssignment struct {
	attAssignmentVal
	messageID uint32
}

func (c *CanModel) getMessageAttAssignments() []*messageAttAssignment {
	assignments := []*messageAttAssignment{}

	for _, msg := range c.Messages {
		for _, ass := range msg.getAttAssignmentValues() {
			assignments = append(assignments, &messageAttAssignment{
				attAssignmentVal: *ass,
				messageID:        msg.ID,
			})
		}
	}

	return assignments
}

type signalAttAssignment struct {
	attAssignmentVal
	messageID  uint32
	signalName string
}

func (c *CanModel) getSignalAttAssignments() []*signalAttAssignment {
	assignments := []*signalAttAssignment{}

	for _, msg := range c.Messages {
		for sigName, sig := range msg.Signals {
			for _, ass := range sig.getAttAssignmentValues() {
				assignments = append(assignments, &signalAttAssignment{
					attAssignmentVal: *ass,
					messageID:        msg.ID,
					signalName:       sigName,
				})
			}
		}
	}

	return assignments
}
*/
