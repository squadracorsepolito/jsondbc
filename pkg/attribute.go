package pkg

import (
	"errors"
	"fmt"
	"math"

	"github.com/squadracorsepolito/acmelib"
)

type attributeKind uint8

const (
	attributeKindGeneral attributeKind = iota
	attributeKindNode
	attributeKindMessage
	attributeKindSignal
)

type attributeType uint8

const (
	attributeTypeInt attributeType = iota
	attributeTypeString
	attributeTypeEnum
	attributeTypeFloat
)

type Attribute struct {
	Int    *AttributeInt    `json:"int,omitempty"`
	String *AttributeString `json:"string,omitempty"`
	Enum   *AttributeEnum   `json:"enum,omitempty"`
	Float  *AttributeFloat  `json:"float,omitempty"`

	attributeName string
	attributeKind attributeKind
	attributeType attributeType
}

func (a *Attribute) initAttribute(attName string) error {
	a.attributeName = attName

	if a.Int != nil {
		a.attributeType = attributeTypeInt
		return a.Int.validate(attName)
	}

	if a.String != nil {
		a.attributeType = attributeTypeString
		return nil
	}

	if a.Enum != nil {
		a.attributeType = attributeTypeEnum
		return a.Enum.initAttributeEnum(attName)
	}

	if a.Float != nil {
		a.attributeType = attributeTypeFloat
		return a.Float.validate(attName)
	}
	return nil
}

func (a *Attribute) validateValue(value any) error {
	var acmeAtt acmelib.Attribute
	var err error
	switch a.attributeType {
	case attributeTypeInt:
		acmeAtt, err = acmelib.NewIntegerAttribute(a.attributeName, a.Int.Default, a.Int.From, a.Int.To)
	case attributeTypeFloat:
		acmeAtt, err = acmelib.NewFloatAttribute(a.attributeName, a.Float.Default, a.Float.From, a.Float.To)
	case attributeTypeEnum:
		acmeAtt, err = acmelib.NewEnumAttribute(a.attributeName, a.Enum.Values...)
	case attributeTypeString:
		acmeAtt = acmelib.NewStringAttribute(a.attributeName, a.String.Default)
	}
	if err != nil {
		return fmt.Errorf("error validating attribute %s: %w", a.attributeName, err)
	}
	host := acmelib.NewNode("attribute_value_validation", 0, 0)
	if err := host.AssignAttribute(acmeAtt, value); err != nil {
		return a.describeValueError(value, err)
	}
	return nil
}

// describeValueError unwraps the acmelib validation error and rephrases it
// against the attribute definition, dropping the throwaway-host entity details.
func (a *Attribute) describeValueError(value any, err error) error {
	var valErr *acmelib.AttributeValueError
	if !errors.As(err, &valErr) {
		return fmt.Errorf("error validating attribute %s: %w", a.attributeName, err)
	}

	switch {
	case errors.Is(valErr.Err, acmelib.ErrOutOfBounds) && a.attributeType == attributeTypeInt:
		return fmt.Errorf("attribute %q: value %v is out of bounds [%d, %d]",
			a.attributeName, value, a.Int.From, a.Int.To)

	case errors.Is(valErr.Err, acmelib.ErrOutOfBounds) && a.attributeType == attributeTypeFloat:
		return fmt.Errorf("attribute %q: value %v is out of bounds [%g, %g]",
			a.attributeName, value, a.Float.From, a.Float.To)

	case errors.Is(valErr.Err, acmelib.ErrNotFound) && a.attributeType == attributeTypeEnum:
		return fmt.Errorf("attribute %q: value %q is not one of the declared values %v",
			a.attributeName, value, a.Enum.Values)

	default:
		return fmt.Errorf("attribute %q: invalid value %v (%v)", a.attributeName, value, valErr.Err)
	}
}

type AttributeInt struct {
	Default int `json:"default"`
	From    int `json:"from"`
	To      int `json:"to"`
}

func (ai *AttributeInt) validate(name string) error {
	if _, err := acmelib.NewIntegerAttribute(name, ai.Default, ai.From, ai.To); err != nil {
		return err
	}
	return nil
}

type AttributeFloat struct {
	Default float64 `json:"default"`
	From    float64 `json:"from"`
	To      float64 `json:"to"`
}

func (af *AttributeFloat) validate(name string) error {
	if _, err := acmelib.NewFloatAttribute(name, af.Default, af.From, af.To); err != nil {
		return err
	}
	return nil
}

type AttributeString struct {
	Default string `json:"default"`
}

type AttributeEnum struct {
	Default string   `json:"default"`
	Values  []string `json:"values"`

	defaultIdx int
}

func (ae *AttributeEnum) initAttributeEnum(name string) error {

	acmelibEnumAttr, err := acmelib.NewEnumAttribute(name, ae.Values...)
	if err != nil {
		return fmt.Errorf("error validating enum attribute %s: %w", name, err)
	}

	ae.defaultIdx = 0

	if ae.Default == "" {
		return nil
	}

	for idx, value := range acmelibEnumAttr.Values() {
		if value == ae.Default {
			ae.defaultIdx = idx
			return nil
		}
	}
	return fmt.Errorf("attribute %s: default value %q is not one of the declared values %v", name, ae.Default, ae.Values)
}

type NodeAttribute struct {
	*Attribute

	assignedNodes map[string]*Node
}

func newNodeAttribute(att *Attribute) *NodeAttribute {
	return &NodeAttribute{
		Attribute: att,

		assignedNodes: make(map[string]*Node),
	}
}

func (na *NodeAttribute) initNodeAttribute(attName string) error {
	na.attributeKind = attributeKindNode
	if na.assignedNodes == nil {
		na.assignedNodes = make(map[string]*Node)
	}
	return na.initAttribute(attName)
}

func (na *NodeAttribute) asAttribute() *Attribute {
	return na.Attribute
}

func (na *NodeAttribute) assignNode(node *Node) error {
	val := node.Attributes[na.attributeName]
	value, err := normalizeValue(na.attributeName, val, na.attributeType)
	if err != nil {
		return err
	}
	node.Attributes[na.attributeName] = value
	if err := na.Attribute.validateValue(value); err != nil {
		return err
	}
	na.assignedNodes[node.nodeName] = node
	return nil
}

type MessageAttribute struct {
	*Attribute

	assignedMessages map[uint32]*Message
}

func newMessageAttribute(att *Attribute) *MessageAttribute {
	return &MessageAttribute{
		Attribute: att,

		assignedMessages: make(map[uint32]*Message),
	}
}

func (ma *MessageAttribute) initMessageAttribute(attName string) error {
	ma.attributeKind = attributeKindMessage
	if ma.assignedMessages == nil {
		ma.assignedMessages = make(map[uint32]*Message)
	}
	return ma.initAttribute(attName)
}

func (ma *MessageAttribute) asAttribute() *Attribute {
	return ma.Attribute
}

func (ma *MessageAttribute) assignMessage(msg *Message) error {
	val := msg.Attributes[ma.attributeName]
	value, err := normalizeValue(ma.attributeName, val, ma.attributeType)
	if err != nil {
		return err
	}
	msg.Attributes[ma.attributeName] = value
	if err := ma.Attribute.validateValue(value); err != nil {
		return err
	}
	ma.assignedMessages[msg.ID] = msg
	return nil
}

type SignalAttribute struct {
	*Attribute

	assignedSignals map[uint32]map[string]*Signal
}

func newSignalAttribute(att *Attribute) *SignalAttribute {
	return &SignalAttribute{
		Attribute: att,

		assignedSignals: make(map[uint32]map[string]*Signal),
	}
}

func (sa *SignalAttribute) initSignalAttribute(attName string) error {
	sa.attributeKind = attributeKindSignal
	if sa.assignedSignals == nil {
		sa.assignedSignals = make(map[uint32]map[string]*Signal)
	}
	return sa.initAttribute(attName)
}

func (sa *SignalAttribute) asAttribute() *Attribute {
	return sa.Attribute
}

func (sa *SignalAttribute) assignSignal(msgID uint32, signal *Signal) error {
	val := signal.Attributes[sa.attributeName]
	value, err := normalizeValue(sa.attributeName, val, sa.attributeType)
	if err != nil {
		return err
	}
	signal.Attributes[sa.attributeName] = value
	if err := sa.validateValue(value); err != nil {
		return err
	}
	if msg, ok := sa.assignedSignals[msgID]; ok {
		msg[signal.signalName] = signal
		return nil
	}

	sa.assignedSignals[msgID] = make(map[string]*Signal)
	sa.assignedSignals[msgID][signal.signalName] = signal
	return nil
}

type AttributeAssignments struct {
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (aa *AttributeAssignments) getAttributeValue(attName string, attType attributeType, enumAtt *AttributeEnum) string {
	att, ok := aa.Attributes[attName]
	if !ok {
		return ""
	}

	switch attType {
	case attributeTypeInt:
		return formatInt(att.(int))

	case attributeTypeString:
		return formatString(att.(string))

	case attributeTypeFloat:
		return formatFloat(att.(float64))

	case attributeTypeEnum:
		tmp := att.(string)
		for idx, val := range enumAtt.Values {
			if val == tmp {
				return formatInt(idx)
			}
		}
		return "0"
	}

	return ""
}

func normalizeValue(attName string, value any, attType attributeType) (any, error) {
	switch attType {
	case attributeTypeInt:
		switch v := value.(type) {
		case int:
			return v, nil
		case float64:
			if v != math.Trunc(v) {
				return nil, fmt.Errorf("attribute %q: value %v is not an integer", attName, v)
			}
			return int(v), nil
		default:
			return nil, fmt.Errorf("attribute %q: expected a number got %T (%v)", attName, value, value)
		}

	case attributeTypeFloat:
		if v, ok := value.(float64); ok {
			return v, nil
		}
		return nil, fmt.Errorf("attribute %q: expected a float, got %T (%v)", attName, value, value)
	case attributeTypeString, attributeTypeEnum:
		if v, ok := value.(string); ok {
			return v, nil
		}
		return nil, fmt.Errorf("attribute %q: expected a string, got %T (%v)", attName, value, value)
	}
	return nil, fmt.Errorf("attribute %q: unknown type %v", attName, attType)
}
