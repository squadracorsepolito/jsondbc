package pkg

import "strconv"

// CanModel represents the CAN model
type CanModel struct {
	Version  string             `json:"version"`
	Nodes    map[string]Node    `json:"nodes"`
	Messages map[string]Message `json:"messages"`
}

// Node represents a CAN node
type Node struct {
	Description string `json:"description"`
}

// HasDescription returns true if the node has a description
func (n *Node) HasDescription() bool {
	return len(n.Description) > 0
}

// Message represents a CAN message
type Message struct {
	ID          uint32            `json:"id"`
	Description string            `json:"description"`
	Length      uint32            `json:"length"`
	Sender      string            `json:"sender"`
	Signals     map[string]Signal `json:"signals"`
}

// HasDescription returns true if the message has a description
func (m *Message) HasDescription() bool {
	return len(m.Description) > 0
}

// FormatID returns the message ID as a string
func (m *Message) FormatID() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}

// Signal represents a CAN signal in a message
type Signal struct {
	Description string            `json:"description"`
	StartBit    uint32            `json:"startBit"`
	Size        uint32            `json:"size"`
	BigEndian   bool              `json:"bigEndian"`
	Signed      bool              `json:"signed"`
	Unit        string            `json:"unit"`
	Receivers   []string          `json:"receivers"`
	Scale       float64           `json:"scale"`
	Offset      float64           `json:"offset"`
	Min         float64           `json:"min"`
	Max         float64           `json:"max"`
	Bitmap      map[string]uint32 `json:"bitmap"`
}

// Validate validates the signal
func (s *Signal) Validate() {
	if s.Scale == 0 {
		s.Scale = 1
	}
}

// IsBitmap returns true if the signal is a bitmap
func (s *Signal) IsBitmap() bool {
	return len(s.Bitmap) > 0
}

// HasDescription returns true if the signal has a description
func (s *Signal) HasDescription() bool {
	return len(s.Description) > 0
}
