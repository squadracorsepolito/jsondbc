package pkg

import "fmt"

// Node represents a CAN node.
type Node struct {
	attributeAssignment
	Description string `json:"description,omitempty"`

	name string
}

func (n *Node) validate(nodeName string, nodeAtt map[string]*Attribute) error {
	n.name = nodeName
	if err := n.attributeAssignment.validate(nodeAtt); err != nil {
		return fmt.Errorf("node %s: %w", n.name, err)
	}

	return nil
}

// HasDescription returns true if the node has a description.
func (n *Node) HasDescription() bool {
	return len(n.Description) > 0
}
