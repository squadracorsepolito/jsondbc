package pkg

import (
	"fmt"
	"strconv"
)

// Message represents a CAN message.
type Message struct {
	*AttributeAssignments
	ID          uint32             `json:"id"`
	Description string             `json:"description,omitempty"`
	Length      uint32             `json:"length"`
	Sender      string             `json:"sender,omitempty"`
	Signals     map[string]*Signal `json:"signals"`

	messageName  string
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
	m.messageName = msgName

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

// HasDescription returns true if the message has a description.
func (m *Message) HasDescription() bool {
	return len(m.Description) > 0
}

// FormatID returns the message ID as a string.
func (m *Message) FormatID() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}

func (m *Message) validate() error {
	if m.Length == 0 {
		return fmt.Errorf("message [%s] length cannot be 0", m.messageName)
	}
	if len(m.childSignals) == 0 {
		return fmt.Errorf("message [%s] has no signals", m.messageName)
	}

	for _, sig := range m.childSignals {
		if err := sig.validate(); err != nil {
			return err
		}
	}

	return nil
}
