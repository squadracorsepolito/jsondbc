package pkg

import "strconv"

// CanModel represents the CAN model.
type CanModel struct {
	Version  string              `json:"version"`
	Nodes    map[string]*Node    `json:"nodes"`
	Messages map[string]*Message `json:"messages"`
}

// Validate validates the CAN model.
func (c *CanModel) Validate() error {
	for _, node := range c.Nodes {
		if err := node.validate(); err != nil {
			return err
		}
	}

	for _, message := range c.Messages {
		if err := message.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Node represents a CAN node.
type Node struct {
	Description string `json:"description,omitempty"`
}

func (n *Node) validate() error {
	return nil
}

// HasDescription returns true if the node has a description.
func (n *Node) HasDescription() bool {
	return len(n.Description) > 0
}

// Message represents a CAN message.
type Message struct {
	ID          uint32             `json:"id"`
	Description string             `json:"description,omitempty"`
	Length      uint32             `json:"length"`
	Sender      string             `json:"sender,omitempty"`
	Signals     map[string]*Signal `json:"signals"`

	name string
}

func (m *Message) validate() error {
	for _, signal := range m.Signals {
		if err := signal.validate(); err != nil {
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

// Signal represents a CAN signal in a message.
type Signal struct {
	Description string             `json:"description,omitempty"`
	MuxSwitch   uint32             `json:"mux_switch,omitempty"`
	StartBit    uint32             `json:"start_bit"`
	Size        uint32             `json:"size"`
	BigEndian   bool               `json:"big_endian,omitempty"`
	Signed      bool               `json:"signed,omitempty"`
	Unit        string             `json:"unit,omitempty"`
	Receivers   []string           `json:"receivers,omitempty"`
	Scale       float64            `json:"scale"`
	Offset      float64            `json:"offset"`
	Min         float64            `json:"min"`
	Max         float64            `json:"max"`
	Bitmap      map[string]uint32  `json:"bitmap,omitempty"`
	MuxGroup    map[string]*Signal `json:"mux_group,omitempty"`

	name          string
	isMultiplexor bool
	isMultiplexed bool
}

func (s *Signal) validate() error {
	if s.Scale == 0 {
		s.Scale = 1
	}

	return nil
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
