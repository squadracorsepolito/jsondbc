package pkg

import (
	"fmt"

	"github.com/squadracorsepolito/jsondbc/pkg/sym"
)

// Signal represents a CAN signal in a message.
type Signal struct {
	*AttributeAssignments
	Description string `json:"description,omitempty"`

	// Custom attributes
	SendType string `json:"send_type,omitempty"`

	MuxSwitch  uint32             `json:"mux_switch,omitempty"`
	StartBit   uint32             `json:"start_bit"`
	Size       uint32             `json:"size"`
	Endianness string             `json:"endianness"`
	Signed     bool               `json:"signed,omitempty"`
	Unit       string             `json:"unit,omitempty"`
	Receivers  []string           `json:"receivers,omitempty"`
	Scale      float64            `json:"scale"`
	Offset     float64            `json:"offset"`
	Min        float64            `json:"min"`
	Max        float64            `json:"max"`
	Enum       map[string]uint32  `json:"enum,omitempty"`
	EnumRef    string             `json:"enum_ref,omitempty"`
	MuxGroup   map[string]*Signal `json:"mux_group,omitempty"`

	signalName    string
	isMultiplexor bool
	isMultiplexed bool
	isBigEndian   bool
	source        sourceType
}

func (s *Signal) initSignal(sigName string, source sourceType) {
	s.signalName = sigName
	s.source = source

	if s.AttributeAssignments == nil {
		s.AttributeAssignments = &AttributeAssignments{
			Attributes: make(map[string]any),
		}
	}

	s.handleCustomAttributes()

	if s.Scale == 0 {
		s.Scale = 1
	}

	if len(s.Endianness) > 0 {
		switch s.Endianness {
		case "big":
			s.isBigEndian = true
		case "little":
			s.isBigEndian = false
		default:
			s.Endianness = "little"
		}
	} else {
		s.Endianness = "little"
	}

	if len(s.MuxGroup) > 0 {
		s.isMultiplexor = true
	}
}

func (s *Signal) appendDescription(format string, a ...any) {
	s.Description = appendString(s.Description, format, a...)
}

func (s *Signal) handleCustomAttributes() {
	location := fmt.Sprintf("signal '%s'", s.signalName)

	switch s.source {
	case sourceTypeJSON:
		// SigSendType
		if len(s.SendType) > 0 {
			tmpST := checkCustomEnumAttribute(s.SendType, "signal.send_type", sym.SigSendTypeValues, location)
			s.AttributeAssignments.Attributes[sym.SigSendType] = tmpST
			s.appendDescription("(send_type: %s)", tmpST)
		}

	case sourceTypeDBC:
		// SigSendType
		stAtt, hasST := s.AttributeAssignments.Attributes[sym.SigSendType]
		if hasST {
			s.SendType = checkCustomEnumAttribute(stAtt.(string), sym.SigSendType, sym.SigSendTypeValues, location)
			delete(s.AttributeAssignments.Attributes, sym.SigSendType)
		}
	}
}

// IsBitmap returns true if the signal is a bitmap.
func (s *Signal) IsBitmap() bool {
	return len(s.Enum) > 0
}

// IsMultiplexor returns true if the signal is a multiplexor.
func (s *Signal) IsMultiplexor() bool {
	return len(s.MuxGroup) > 0
}

// HasDescription returns true if the signal has a description.
func (s *Signal) HasDescription() bool {
	return len(s.Description) > 0
}

func (s *Signal) validate() error {
	if s.Size == 0 {
		return fmt.Errorf("signal [%s] size cannot be 0", s.signalName)
	}

	return nil
}
