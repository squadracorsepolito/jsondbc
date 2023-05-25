package pkg

import (
	"fmt"
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
	Version           string                `json:"version"`
	BusSpeed          uint32                `json:"bus_speed"`
	Nodes             map[string]*Node      `json:"nodes"`
	Messages          map[string]*Message   `json:"messages"`
	NodeAttributes    map[string]*Attribute `json:"node_attributes"`
	MessageAttributes map[string]*Attribute `json:"message_attributes"`
	SignalAttributes  map[string]*Attribute `json:"signal_attributes"`
}

// Validate validates the CAN model.
func (c *CanModel) Validate() error {
	for attName, att := range c.NodeAttributes {
		if err := att.validate(attName, attributeKindNode); err != nil {
			return fmt.Errorf("node attribute %s: %w", attName, err)
		}
	}
	for attName, att := range c.MessageAttributes {
		if err := att.validate(attName, attributeKindMessage); err != nil {
			return fmt.Errorf("message attribute %s: %w", attName, err)
		}
	}
	for attName, att := range c.SignalAttributes {
		if err := att.validate(attName, attributeKindSignal); err != nil {
			return fmt.Errorf("signal attribute %s: %w", attName, err)
		}
	}

	for nodeName, node := range c.Nodes {
		if err := node.validate(nodeName, c.NodeAttributes); err != nil {
			return err
		}
	}

	for msgName, msg := range c.Messages {
		if err := msg.validate(msgName, c.MessageAttributes, c.SignalAttributes); err != nil {
			return err
		}
	}

	return nil
}

func (c *CanModel) getAttributes() []*Attribute {
	attributes := []*Attribute{}

	for _, att := range c.NodeAttributes {
		attributes = append(attributes, att)
	}
	for _, att := range c.MessageAttributes {
		attributes = append(attributes, att)
	}
	for _, att := range c.SignalAttributes {
		attributes = append(attributes, att)
	}

	return attributes
}

type nodeAttributeAssignment struct {
	attributeAssignmentValue
	nodeName string
}

func (c *CanModel) getNodeAttributeAssignments() []*nodeAttributeAssignment {
	assignments := []*nodeAttributeAssignment{}

	return assignments
}

type messageAttributeAssignment struct {
	attributeAssignmentValue
	messageID uint32
}

func (c *CanModel) getMessageAttributeAssignments() []*messageAttributeAssignment {
	assignments := []*messageAttributeAssignment{}

	for _, msg := range c.Messages {
		for _, ass := range msg.getAttributeAssignmentValues() {
			assignments = append(assignments, &messageAttributeAssignment{attributeAssignmentValue: *ass, messageID: msg.ID})
		}
	}

	return assignments
}

type signalAttributeAssignment struct {
	attributeAssignmentValue
	messageID  uint32
	signalName string
}

func (c *CanModel) getSignalAttributeAssignments() []*signalAttributeAssignment {
	assignments := []*signalAttributeAssignment{}

	for _, msg := range c.Messages {
		for sigName, sig := range msg.Signals {
			for _, ass := range sig.getAttributeAssignmentValues() {
				assignments = append(assignments, &signalAttributeAssignment{attributeAssignmentValue: *ass, messageID: msg.ID, signalName: sigName})
			}
		}
	}

	return assignments
}
