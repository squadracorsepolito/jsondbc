package pkg

import (
	"errors"
	"fmt"
)

type attributeKind uint8

const (
	attributeKindNode attributeKind = iota
	attributeKindMessage
	attributeKindSignal
)

type attributeType uint8

const (
	attributeTypeInt attributeType = iota
	attributeTypeString
	attributeTypeEnum
)

type Attribute struct {
	Int    *AttributeInt    `json:"int"`
	String *AttributeString `json:"string"`
	Enum   *AttributeEnum   `json:"enum"`

	name          string
	attributeKind attributeKind
	attributeType attributeType
}

func (a *Attribute) validate(name string, kind attributeKind) error {
	a.name = name
	a.attributeKind = kind

	if a.Int != nil {
		a.attributeType = attributeTypeInt
		if err := a.Int.validate(); err != nil {
			return err
		}
	} else if a.String != nil {
		a.attributeType = attributeTypeString
	} else if a.Enum != nil {
		a.attributeType = attributeTypeEnum
		if err := a.Enum.validate(); err != nil {
			return err
		}
	} else {
		return errors.New("unset, set it to int, string or enum")
	}

	return nil
}

type AttributeInt struct {
	Default int `json:"default"`
	From    int `json:"from"`
	To      int `json:"to"`
}

func (ai *AttributeInt) validate() error {
	if ai.Default < ai.From || ai.Default > ai.To {
		return fmt.Errorf("default value %d is not in range [%d, %d]", ai.Default, ai.From, ai.To)
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

func (ae *AttributeEnum) validate() error {
	if ae.Default == "" {
		ae.defaultIdx = 0
		return nil
	}

	for idx, value := range ae.Values {
		if value == ae.Default {
			ae.defaultIdx = idx
			return nil
		}

	}

	return fmt.Errorf("default value %s is not part of the enum, valid values are %v", ae.Default, ae.Values)
}

type attributeAssignment struct {
	IntAttributes    map[string]int    `json:"int_attributes,omitempty"`
	StringAttributes map[string]string `json:"string_attributes,omitempty"`
	EnumAttributes   map[string]string `json:"enum_attributes,omitempty"`

	enumAttributeIdxs map[string]int
}

func (aa *attributeAssignment) validate(attributes map[string]*Attribute) error {
	for attName, attVal := range aa.IntAttributes {
		att, ok := attributes[attName]
		if !ok {
			return fmt.Errorf("int attribute %s doesn't exist", attName)
		}
		if att.attributeType != attributeTypeInt {
			return fmt.Errorf("int attribute %s is not of type int", attName)
		}
		if attVal < att.Int.From || attVal > att.Int.To {
			return fmt.Errorf("int attribute %s is out of range [%d, %d]", attName, att.Int.From, att.Int.To)
		}
	}

	for attName := range aa.StringAttributes {
		att, ok := attributes[attName]
		if !ok {
			return fmt.Errorf("string attribute %s doesn't exist", attName)
		}
		if att.attributeType != attributeTypeString {
			return fmt.Errorf("string attribute %s is not of type string", attName)
		}
	}

	aa.enumAttributeIdxs = make(map[string]int)
	for attName, attVal := range aa.EnumAttributes {
		att, ok := attributes[attName]
		if !ok {
			return fmt.Errorf("enum attribute %s doesn't exist", attName)
		}
		if att.attributeType != attributeTypeEnum {
			return fmt.Errorf("enum attribute %s is not of type enum", attName)
		}
		found := false
		for idx, value := range att.Enum.Values {
			if value == attVal {
				found = true
				aa.enumAttributeIdxs[attName] = idx
				break
			}
		}
		if !found {
			return fmt.Errorf("enum attribute %s has invalid value %s, should be one of %v", attName, attVal, att.Enum.Values)
		}
	}

	return nil
}

type attributeAssignmentValue struct {
	attType        attributeType
	attName        string
	intAttValue    int
	stringAttValue string
	enumAttValue   int
}

func (aa *attributeAssignment) getAttributeAssignmentValues() []*attributeAssignmentValue {
	values := []*attributeAssignmentValue{}

	for attName, intAttValue := range aa.IntAttributes {
		values = append(values, &attributeAssignmentValue{
			attName:     attName,
			attType:     attributeTypeInt,
			intAttValue: intAttValue,
		})
	}

	for attName, stringAttValue := range aa.StringAttributes {
		values = append(values, &attributeAssignmentValue{
			attName:        attName,
			attType:        attributeTypeString,
			stringAttValue: stringAttValue,
		})
	}

	for attName, enumIdx := range aa.enumAttributeIdxs {
		values = append(values, &attributeAssignmentValue{
			attName:      attName,
			attType:      attributeTypeEnum,
			enumAttValue: enumIdx,
		})
	}

	return values
}
