package cangoru

import (
	"fmt"
)

type Description struct {
	Description string
}

func (d *Description) SetDescription(desc string) {
	d.Description = desc
}

func (d *Description) GetDescription() string {
	return d.Description
}

func (d *Description) HasDescription() bool {
	return len(d.Description) > 0
}

type CAN struct {
	Description
	AttributeMap

	VersionString string
	Baudrate      uint
	Nodes         map[string]*Node
	Messages      map[MessageID]*Message
	Attributes    map[string]*Attribute
}

func NewCAN() *CAN {
	return &CAN{
		Nodes:      make(map[string]*Node),
		Messages:   make(map[MessageID]*Message),
		Attributes: make(map[string]*Attribute),
	}
}

func (c *CAN) SetVersionString(versionString string) {
	c.VersionString = versionString
}

func (c *CAN) SetBaudrate(baudrate uint) {
	c.Baudrate = baudrate
}

func (c *CAN) AddNode(node *Node) error {
	if _, ok := c.Nodes[node.Name]; ok {
		return fmt.Errorf("duplicated node: %s", node.Name)
	}
	c.Nodes[node.Name] = node
	return nil
}

func (c *CAN) GetNode(nodeName string) (*Node, error) {
	node, ok := c.Nodes[nodeName]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", nodeName)
	}
	return node, nil
}

func (c *CAN) AddMessage(msg *Message) error {
	if _, ok := c.Messages[msg.ID]; ok {
		return fmt.Errorf("duplicated message: %d", msg.ID)
	}
	c.Messages[msg.ID] = msg
	return nil
}

func (c *CAN) GetMessage(msgID MessageID) (*Message, error) {
	msg, ok := c.Messages[msgID]
	if !ok {
		return nil, fmt.Errorf("message not found: %d", msgID)
	}
	return msg, nil
}

func (c *CAN) AddAttribute(att *Attribute) error {
	if _, ok := c.Attributes[att.Name]; ok {
		return fmt.Errorf("duplicated attribute: %s", att.Name)
	}
	c.Attributes[att.Name] = att
	return nil
}

func (c *CAN) GetAttribute(attName string) (*Attribute, error) {
	att, ok := c.Attributes[attName]
	if !ok {
		return nil, fmt.Errorf("attribute not found: %s", attName)
	}
	return att, nil
}
