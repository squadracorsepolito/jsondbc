package cangoru

import (
	"fmt"
)

type MessageID uint32

func NewMessageID(id uint32) MessageID {
	return MessageID(id)
}

type bitmaskType uint8

const (
	empty bitmaskType = 1 << iota
	normal
	multiplexor
	multiplexed
)

type messageBitset struct {
	length uint
	mask   []bitmaskType
}

func newMessageBitset(length uint) *messageBitset {
	m := make([]bitmaskType, 8*length)
	for i := uint(0); i < 8*length; i++ {
		m[i] = empty
	}
	return &messageBitset{
		length: length,
		mask:   m,
	}
}

func (mb *messageBitset) checkMask(from, to uint, mask bitmaskType) error {
	for i := from; i <= to; i++ {
		m := mb.mask[i]
		if m != empty && m&multiplexed == 0 {
			return fmt.Errorf("overlapping signal bits: from %d to %d", from, to)
		}
		mb.mask[i] = mask
	}
	return nil
}

type Message struct {
	Description
	AttributeMap

	ID      MessageID
	Name    string
	Size    uint
	Period  uint
	Signals map[string]*Signal

	bitset             *messageBitset
	multiplexorSignals map[string]*Signal
	multiplexedSignals map[string]*Signal
}

func NewMessage(id MessageID, name string, size uint) *Message {
	bitMask := make([][]bool, 1)
	bitMask[0] = make([]bool, 8*size)

	return &Message{
		ID:      id,
		Name:    name,
		Size:    size,
		Signals: make(map[string]*Signal),

		bitset:             newMessageBitset(size),
		multiplexorSignals: make(map[string]*Signal),
		multiplexedSignals: make(map[string]*Signal),
	}
}

func (m *Message) errorf(format string, arg ...any) error {
	placeholder := fmt.Sprintf(`[message %d "%s"] `, m.ID, m.Name)
	return fmt.Errorf(placeholder+format, arg...)
}

func (m *Message) SetPeriod(period uint) {
	m.Period = period
}

func (m *Message) AddSignal(signal *Signal) error {
	if _, ok := m.Signals[signal.Name]; ok {
		return m.errorf("duplicated signal: %s", signal.Name)
	}

	if signal.StartBit+signal.Size > m.Size*8 {
		return m.errorf(`signal "%s" exceeds message size "%d": start bit %d, size %d`, signal.Name, m.Size, signal.StartBit, signal.Size)
	}

	sigMask := normal
	if signal.IsMultiplexor {
		m.multiplexorSignals[signal.Name] = signal
		sigMask = sigMask | multiplexor
	}
	if signal.IsMultiplexed {
		m.multiplexedSignals[signal.Name] = signal
		sigMask = sigMask | multiplexed
	}

	if err := m.bitset.checkMask(signal.StartBit, signal.Size+signal.StartBit-1, sigMask); err != nil {
		return m.errorf(`signal "%s" %v`, signal.Name, err)
	}

	m.Signals[signal.Name] = signal
	return nil
}

func (m *Message) GetSignal(name string) (*Signal, error) {
	sig, ok := m.Signals[name]
	if !ok {
		return nil, m.errorf("signal not found: %s", name)
	}
	return sig, nil
}

func (m *Message) HasMuxSignals() bool {
	return len(m.multiplexorSignals) > 0
}

func (m *Message) HasExtendedMuxSignals() bool {
	return len(m.multiplexorSignals) > 1
}
