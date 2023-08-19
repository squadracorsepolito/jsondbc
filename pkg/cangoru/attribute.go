package cangoru

import (
	"fmt"

	"github.com/squadracorsepolito/jsondbc/pkg/cangoru/dbc"
	"golang.org/x/exp/slices"
)

type AttributeType uint

const (
	AttributeTypeInt AttributeType = iota
	AttributeTypeHex
	AttributeTypeFloat
	AttributeTypeString
	AttributeTypeEnum
)

type AttributeKind uint

const (
	AttributeKindGeneral AttributeKind = iota
	AttributeKindNode
	AttributeKindMessage
	AttributeKindSignal
)

func (ak AttributeKind) ToDBC() dbc.AttributeKind {
	switch ak {
	case AttributeKindNode:
		return dbc.AttributeNode
	case AttributeKindMessage:
		return dbc.AttributeMessage
	case AttributeKindSignal:
		return dbc.AttributeSignal
	default:
		return dbc.AttributeGeneral
	}
}

type AttributeDefault[T int | float64 | string] struct {
	Default T
}

type AttributeNumber[T int | float64] struct {
	AttributeDefault[T]

	From T
	To   T
}

func (an *AttributeNumber[T]) isInRange(value T) bool {
	return value >= an.From && value <= an.To
}

type AttributeEnum struct {
	AttributeDefault[string]

	Values []string
}

func (ae *AttributeEnum) isInRange(value int) bool {
	return value >= 0 && value < len(ae.Values)
}

type Attribute struct {
	Type   AttributeType
	Kind   AttributeKind
	Name   string
	Int    *AttributeNumber[int]
	Hex    *AttributeNumber[int]
	Float  *AttributeNumber[float64]
	String *AttributeDefault[string]
	Enum   *AttributeEnum
}

func (a *Attribute) errorf(format string, args ...any) error {
	kind := "general"
	switch a.Kind {
	case AttributeKindNode:
		kind = "node"
	case AttributeKindMessage:
		kind = "message"
	case AttributeKindSignal:
		kind = "signal"
	}
	placeholder := fmt.Sprintf(`[attribute "%s" kind "%s"] `, a.Name, kind)
	return fmt.Errorf(placeholder+format, args...)
}

func NewIntAttribute(kind AttributeKind, name string, from int, to int) *Attribute {
	return &Attribute{
		Type: AttributeTypeInt,
		Kind: kind,
		Name: name,
		Int: &AttributeNumber[int]{
			From: from,
			To:   to,
		},
	}
}

func (a *Attribute) SetIntDefault(def int) error {
	if a.Type != AttributeTypeInt {
		return a.errorf("is not an integer attribute")
	}

	if !a.Int.isInRange(def) {
		return a.errorf("default value is out of range")
	}

	a.Int.Default = def
	return nil
}

func NewFloatAttribute(kind AttributeKind, name string, from float64, to float64) *Attribute {
	return &Attribute{
		Type: AttributeTypeFloat,
		Kind: kind,
		Name: name,
		Float: &AttributeNumber[float64]{
			From: from,
			To:   to,
		},
	}
}

func (a *Attribute) SetFloatDefault(def float64) error {
	if a.Type != AttributeTypeFloat {
		return a.errorf("is not a float attribute")
	}

	if !a.Float.isInRange(def) {
		return a.errorf("default value is out of range")
	}

	a.Float.Default = def
	return nil
}

func NewHexAttribute(kind AttributeKind, name string, from int, to int) *Attribute {
	return &Attribute{
		Type: AttributeTypeHex,
		Kind: kind,
		Name: name,
		Hex: &AttributeNumber[int]{
			From: from,
			To:   to,
		},
	}
}

func (a *Attribute) SetHexDefault(def int) error {
	if a.Type != AttributeTypeHex {
		return a.errorf("is not an hex attribute")
	}

	if !a.Hex.isInRange(def) {
		return a.errorf("default value is out of range")
	}

	a.Hex.Default = def
	return nil
}

func NewStringAttribute(kind AttributeKind, name string) *Attribute {
	return &Attribute{
		Type:   AttributeTypeString,
		Kind:   kind,
		Name:   name,
		String: &AttributeDefault[string]{},
	}
}

func (a *Attribute) SetStringDefault(def string) error {
	if a.Type != AttributeTypeString {
		return a.errorf("is not a string attribute")
	}

	a.String.Default = def
	return nil
}

func NewEnumAttribute(kind AttributeKind, name string, values []string) *Attribute {
	return &Attribute{
		Type: AttributeTypeEnum,
		Kind: kind,
		Name: name,
		Enum: &AttributeEnum{
			Values: values,
		},
	}
}

func (a *Attribute) SetEnumDefault(def string) error {
	if a.Type != AttributeTypeEnum {
		return a.errorf("is not an enum attribute")
	}

	if !slices.Contains(a.Enum.Values, def) {
		return a.errorf("default value is not defined in enum")
	}

	a.Enum.Default = def
	return nil
}

func (a *Attribute) ToDBC() (*dbc.Attribute, *dbc.AttributeDefault) {
	att := &dbc.Attribute{
		Name: a.Name,
		Kind: a.Kind.ToDBC(),
	}

	attDef := &dbc.AttributeDefault{
		AttributeName: a.Name,
	}

	switch a.Type {
	case AttributeTypeInt:
		att.Type = dbc.AttributeInt
		att.MinInt = a.Int.From
		att.MaxInt = a.Int.To
		attDef.Type = dbc.AttributeDefaultInt
		attDef.ValueInt = a.Int.Default

	case AttributeTypeFloat:
		att.Type = dbc.AttributeFloat
		att.MinFloat = a.Float.From
		att.MaxFloat = a.Float.To
		attDef.Type = dbc.AttributeDefaultFloat
		attDef.ValueFloat = a.Float.Default

	case AttributeTypeHex:
		att.Type = dbc.AttributeHex
		att.MinHex = a.Hex.From
		att.MaxHex = a.Hex.To
		attDef.Type = dbc.AttributeDefaultHex
		attDef.ValueHex = a.Hex.Default

	case AttributeTypeString:
		att.Type = dbc.AttributeString
		attDef.Type = dbc.AttributeDefaultString
		attDef.ValueString = a.String.Default

	case AttributeTypeEnum:
		att.Type = dbc.AttributeEnum
		att.EnumValues = a.Enum.Values
		attDef.Type = dbc.AttributeDefaultString
		attDef.ValueString = a.Enum.Default
	}

	return att, attDef
}

type AttributeValue struct {
	Definition  *Attribute
	IntValue    int
	HexValue    int
	FloatValue  float64
	StringValue string
	EnumValue   int
}

func NewIntAttributeValue(att *Attribute, value int) (*AttributeValue, error) {
	if att.Type != AttributeTypeInt {
		return nil, att.errorf("is not an integer attribute")
	}

	if !att.Int.isInRange(value) {
		return nil, att.errorf("value is out of range")
	}

	return &AttributeValue{
		Definition: att,
		IntValue:   value,
	}, nil
}

func NewFloatAttributeValue(att *Attribute, value float64) (*AttributeValue, error) {
	if att.Type != AttributeTypeFloat {
		return nil, att.errorf("is not a float attribute")
	}

	if !att.Float.isInRange(value) {
		return nil, att.errorf("value is out of range")
	}

	return &AttributeValue{
		Definition: att,
		FloatValue: value,
	}, nil
}

func NewHexAttributeValue(att *Attribute, value int) (*AttributeValue, error) {
	if att.Type != AttributeTypeHex {
		return nil, att.errorf("is not an hex attribute")
	}

	if !att.Hex.isInRange(value) {
		return nil, att.errorf("value is out of range")
	}

	return &AttributeValue{
		Definition: att,
		HexValue:   value,
	}, nil
}

func NewStringAttributeValue(att *Attribute, value string) (*AttributeValue, error) {
	if att.Type != AttributeTypeString {
		return nil, att.errorf("is not a string attribute")
	}

	return &AttributeValue{
		Definition:  att,
		StringValue: value,
	}, nil
}

func NewEnumAttributeValue(att *Attribute, value int) (*AttributeValue, error) {
	if att.Type != AttributeTypeEnum {
		return nil, att.errorf("is not an enum attribute")
	}

	if !att.Enum.isInRange(value) {
		return nil, att.errorf("value is out of range")
	}

	return &AttributeValue{
		Definition: att,
		EnumValue:  value,
	}, nil
}

func (av *AttributeValue) ToDBC() *dbc.AttributeValue {
	attVal := &dbc.AttributeValue{
		AttributeName: av.Definition.Name,
		AttributeKind: av.Definition.Kind.ToDBC(),
	}

	switch av.Definition.Type {
	case AttributeTypeInt:
		attVal.Type = dbc.AttributeValueInt
		attVal.ValueInt = av.IntValue

	case AttributeTypeFloat:
		attVal.Type = dbc.AttributeValueFloat
		attVal.ValueFloat = av.FloatValue

	case AttributeTypeHex:
		attVal.Type = dbc.AttributeValueHex
		attVal.ValueHex = av.HexValue

	case AttributeTypeString:
		attVal.Type = dbc.AttributeValueString
		attVal.ValueString = av.StringValue

	case AttributeTypeEnum:
		attVal.Type = dbc.AttributeValueInt
		attVal.ValueInt = av.EnumValue
	}

	return attVal
}

type AttributeMap struct {
	Attributes map[string]*AttributeValue
}

func (am *AttributeMap) AssignAttribute(att *AttributeValue) {
	if am.Attributes == nil {
		am.Attributes = make(map[string]*AttributeValue)
	}
	am.Attributes[att.Definition.Name] = att
}

func (am *AttributeMap) GetAttributeValues() map[string]*AttributeValue {
	return am.Attributes
}
