package pkg

import "strconv"

type CanModel struct {
	Version  string             `json:"version"`
	Nodes    map[string]Node    `json:"nodes"`
	Messages map[string]Message `json:"messages"`
}

type Node struct {
	Description string `json:"description"`
}

func (n *Node) HasDescription() bool {
	return len(n.Description) > 0
}

type Message struct {
	ID          uint32            `json:"id"`
	Description string            `json:"description"`
	Length      uint32            `json:"length"`
	Sender      string            `json:"sender"`
	Signals     map[string]Signal `json:"signals"`
}

func (m *Message) HasDescription() bool {
	return len(m.Description) > 0
}

func (m *Message) FormatID() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}

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

func (s *Signal) Validate() {
	if s.Scale == 0 {
		s.Scale = 1
	}
}

func (s *Signal) IsBitmap() bool {
	return len(s.Bitmap) > 0
}

func (s *Signal) HasDescription() bool {
	return len(s.Description) > 0
}
