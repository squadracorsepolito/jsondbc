package pkg

import (
	"fmt"
	"strconv"
)

// Message represents a CAN message.
type Message struct {
	attributeAssignment
	ID          uint32             `json:"id"`
	Description string             `json:"description,omitempty"`
	Length      uint32             `json:"length"`
	Sender      string             `json:"sender,omitempty"`
	Signals     map[string]*Signal `json:"signals"`

	name string
}

func (m *Message) validate(msgName string, msgAtt map[string]*Attribute, sigAtt map[string]*Attribute) error {
	m.name = msgName

	if err := m.attributeAssignment.validate(msgAtt); err != nil {
		return fmt.Errorf("message %s: %w", m.name, err)
	}

	for _, signal := range m.Signals {
		if err := signal.validate(sigAtt); err != nil {
			return err
		}
	}

	return nil
}

// HasDescription returns true if the message has a description.
func (m *Message) HasDescription() bool {
	return len(m.Description) > 0
}

// FormatID returns the message ID as a string.
func (m *Message) FormatID() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}
