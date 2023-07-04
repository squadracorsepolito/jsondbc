package pkg

import "github.com/squadracorsepolito/jsondbc/pkg/sym"

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

func (a *Attribute) initAttribute(attName string) {
	a.attributeName = attName

	if a.Int != nil {
		a.attributeType = attributeTypeInt
		return
	}

	if a.String != nil {
		a.attributeType = attributeTypeString
		return
	}

	if a.Enum != nil {
		a.attributeType = attributeTypeEnum
		a.Enum.initAttributeEnum()
		return
	}

	if a.Float != nil {
		a.attributeType = attributeTypeFloat
		return
	}
}

type AttributeInt struct {
	Default int `json:"default"`
	From    int `json:"from"`
	To      int `json:"to"`
}

type AttributeFloat struct {
	Default float64 `json:"default"`
	From    float64 `json:"from"`
	To      float64 `json:"to"`
}

type AttributeString struct {
	Default string `json:"default"`
}

type AttributeEnum struct {
	Default string   `json:"default"`
	Values  []string `json:"values"`

	defaultIdx int
}

func (ae *AttributeEnum) initAttributeEnum() {
	ae.defaultIdx = 0

	if ae.Default == "" {
		return
	}

	for idx, value := range ae.Values {
		if value == ae.Default {
			ae.defaultIdx = idx
			return
		}
	}
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

func (na *NodeAttribute) initNodeAttribute(attName string) {
	na.attributeKind = attributeKindNode
	if na.assignedNodes == nil {
		na.assignedNodes = make(map[string]*Node)
	}
	na.initAttribute(attName)
}

func (na *NodeAttribute) asAttribute() *Attribute {
	return na.Attribute
}

func (na *NodeAttribute) assignNode(node *Node) {
	na.assignedNodes[node.nodeName] = node
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

func (ma *MessageAttribute) initMessageAttribute(attName string) {
	ma.attributeKind = attributeKindMessage
	if ma.assignedMessages == nil {
		ma.assignedMessages = make(map[uint32]*Message)
	}
	ma.initAttribute(attName)
}

func (ma *MessageAttribute) asAttribute() *Attribute {
	return ma.Attribute
}

func (ma *MessageAttribute) assignMessage(msg *Message) {
	ma.assignedMessages[msg.ID] = msg
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

func (sa *SignalAttribute) initSignalAttribute(attName string) {
	sa.attributeKind = attributeKindSignal
	if sa.assignedSignals == nil {
		sa.assignedSignals = make(map[uint32]map[string]*Signal)
	}
	sa.initAttribute(attName)
}

func (sa *SignalAttribute) asAttribute() *Attribute {
	return sa.Attribute
}

func (sa *SignalAttribute) assignSignal(msgID uint32, signal *Signal) {
	if msg, ok := sa.assignedSignals[msgID]; ok {
		msg[signal.signalName] = signal
		return
	}

	sa.assignedSignals[msgID] = make(map[string]*Signal)
	sa.assignedSignals[msgID][signal.signalName] = signal
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
		if attName == sym.MsgPeriodAttribute {
			return formatInt(int(att.(uint32)))
		}
		return formatInt(int(att.(float64)))

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
	}

	return ""
}
