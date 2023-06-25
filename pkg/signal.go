package pkg

import "fmt"

// Signal represents a CAN signal in a message.
type Signal struct {
	*AttributeAssignments
	Description string             `json:"description,omitempty"`
	MuxSwitch   uint32             `json:"mux_switch,omitempty"`
	StartBit    uint32             `json:"start_bit"`
	Size        uint32             `json:"size"`
	Endianness  string             `json:"endianness"`
	Signed      bool               `json:"signed,omitempty"`
	Unit        string             `json:"unit,omitempty"`
	Receivers   []string           `json:"receivers,omitempty"`
	Scale       float64            `json:"scale"`
	Offset      float64            `json:"offset"`
	Min         float64            `json:"min"`
	Max         float64            `json:"max"`
	Bitmap      map[string]uint32  `json:"bitmap,omitempty"`
	MuxGroup    map[string]*Signal `json:"mux_group,omitempty"`

	signalName    string
	isMultiplexor bool
	isMultiplexed bool
	isBigEndian   bool
}

func (s *Signal) initSignal(sigName string) {
	s.signalName = sigName

	if s.AttributeAssignments == nil {
		s.AttributeAssignments = &AttributeAssignments{
			Attributes: make(map[string]any),
		}
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

// IsBitmap returns true if the signal is a bitmap.
func (s *Signal) IsBitmap() bool {
	return len(s.Bitmap) > 0
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
