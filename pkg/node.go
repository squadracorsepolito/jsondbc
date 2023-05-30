package pkg

// Node represents a CAN node.
type Node struct {
	*AttributeAssignments
	Description string `json:"description,omitempty"`

	nodeName string
}

func (n *Node) initNode(nodeName string) {
	n.nodeName = nodeName

	if n.AttributeAssignments == nil {
		n.AttributeAssignments = &AttributeAssignments{
			Attributes: make(map[string]any),
		}
	}
}

// HasDescription returns true if the node has a description.
func (n *Node) HasDescription() bool {
	return len(n.Description) > 0
}
