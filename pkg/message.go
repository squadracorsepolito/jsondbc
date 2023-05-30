package pkg

import (
	"strconv"
)

type attributes struct {
	Attributes map[string]any `json:"attributes"`
}

// Message represents a CAN message.
type Message struct {
	*AttributeAssignments
	ID          uint32             `json:"id"`
	Description string             `json:"description,omitempty"`
	Length      uint32             `json:"length"`
	Sender      string             `json:"sender,omitempty"`
	Signals     map[string]*Signal `json:"signals"`

	name         string
	childSignals map[string]*Signal
}

func (m *Message) initSignalRec(sigName string, sig *Signal) {
	sig.initSignal(sigName)
	m.childSignals[sigName] = sig
	if !sig.isMultiplexor {
		return
	}

	for muxedSigName, muxedSig := range sig.MuxGroup {
		m.initSignalRec(muxedSigName, muxedSig)
	}
}

func (m *Message) initMessage(msgName string) {
	m.name = msgName

	if m.AttributeAssignments == nil {
		m.AttributeAssignments = &AttributeAssignments{
			Attributes: make(map[string]any),
		}
	}

	m.childSignals = make(map[string]*Signal)

	for sigName, sig := range m.Signals {
		sig.initSignal(sigName)
		m.childSignals[sigName] = sig
		if !sig.isMultiplexor {
			continue
		}

		for muxedSigName, muxedSig := range sig.MuxGroup {
			m.initSignalRec(muxedSigName, muxedSig)
		}
	}

}

/*
func (m *Message) validate(msgName string, msgAtt map[string]*Attribute, sigAtt map[string]*Attribute) error {
	m.name = msgName

	if err := m.attributeAssignment.validate(msgAtt); err != nil {
		return fmt.Errorf("message %s: %w", m.name, err)
	}

	for _, signal := range m.Signals {
		if err := signal.validate(sigAtt); err != nil {
			return fmt.Errorf("message %s: %w", m.name, err)
		}
	}

	return nil
}*/

// HasDescription returns true if the message has a description.
func (m *Message) HasDescription() bool {
	return len(m.Description) > 0
}

// FormatID returns the message ID as a string.
func (m *Message) FormatID() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}
