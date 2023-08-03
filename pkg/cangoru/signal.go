package cangoru

import (
	"fmt"
	"math"
)

type SignalByteOrder uint

const (
	LittleEndian SignalByteOrder = iota
	BigEndian
)

type SignalValueType uint

const (
	Unsigned SignalValueType = iota
	Signed
)

func checkSignalValue[T uint | float64](sigSize uint, valueTyp SignalValueType, value T) bool {
	if valueTyp == Unsigned {
		return value <= T(math.Pow(2, float64(sigSize))-1)
	}
	return value >= T(-(math.Pow(2, float64(sigSize-1))-1)) && value <= T(math.Pow(2, float64(sigSize-1))-1)
}

type Signal struct {
	Description
	AttributeMap

	Name      string
	Size      uint
	StartBit  uint
	ByteOrder SignalByteOrder
	ValueType SignalValueType
	Factor    float64
	Offset    float64
	Min       float64
	Max       float64
	Unit      string
	MapValues map[uint]string

	IsMultiplexor bool
	IsMultiplexed bool
	MuxSignals    []*Signal
	MuxIndexes    []uint
}

func NewSignal(name string, size uint, startBit uint, byteOrder SignalByteOrder,
	valueType SignalValueType, factor float64, offset float64, min float64, max float64, unit string) (*Signal, error) {

	sig := &Signal{
		Name:      name,
		Size:      size,
		StartBit:  startBit,
		ByteOrder: byteOrder,
		ValueType: valueType,
		Factor:    factor,
		Offset:    offset,
		Min:       min,
		Max:       max,
		Unit:      unit,
		MapValues: make(map[uint]string),
	}

	if size == 0 {
		return nil, sig.errorf("size is zero")
	}

	if !checkSignalValue(size, valueType, factor) {
		return nil, sig.errorf("factor out of range: %f", factor)
	}
	if !checkSignalValue(size, valueType, offset) {
		return nil, sig.errorf("offset out of range: %f", offset)
	}
	if !checkSignalValue(size, valueType, min) {
		return nil, sig.errorf("min out of range: %f", min)
	}
	if !checkSignalValue(size, valueType, max) {
		return nil, sig.errorf("max out of range: %f", max)
	}

	return sig, nil
}

func (s *Signal) errorf(format string, arg ...any) error {
	placeholder := fmt.Sprintf(`[signal "%s"] `, s.Name)
	return fmt.Errorf(placeholder+format, arg...)
}

func (s *Signal) SetIsMultiplexor() {
	s.IsMultiplexor = true
}

func (s *Signal) SetIsMultiplexed() {
	s.IsMultiplexed = true
}

func (s *Signal) AddMapValue(index uint, value string) error {
	if !checkSignalValue(s.Size, s.ValueType, index) {
		return s.errorf("index out of range: %d", index)
	}
	if _, ok := s.MapValues[index]; ok {
		return s.errorf("duplicated index: %d", index)
	}
	s.MapValues[index] = value
	return nil
}

func (s *Signal) AddMuxSignal(muxSig *Signal, muxIndexes ...uint) error {
	if len(muxIndexes) == 0 {
		return muxSig.errorf("got no mux indexes")
	}
	muxSig.MuxIndexes = muxIndexes
	s.MuxSignals = append(s.MuxSignals, muxSig)
	return nil
}
