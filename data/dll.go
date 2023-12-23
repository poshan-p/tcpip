package data

type Dll struct {
	Previous *Dll
	Next     *Dll
}

func (dll *Dll) Init() {
	dll.Next = nil
	dll.Previous = nil
}

func (dll *Dll) AddNode(newNode *Dll) {
	if dll.Next == nil {
		dll.Next = newNode
		newNode.Previous = dll
		return
	}

	temp := dll.Next
	dll.Next = newNode
	newNode.Previous = dll
	newNode.Next = temp
	temp.Previous = newNode
}

func (dll *Dll) RemoveNode() {
	if dll.Previous == nil {
		if dll.Next != nil {
			dll.Next.Previous = nil
			dll.Next = nil
		}
		return
	}

	if dll.Next == nil {
		dll.Previous.Next = nil
		dll.Previous = nil
		return
	}

	dll.Previous.Next = dll.Next
	dll.Next.Previous = dll.Previous
	dll.Next = nil
	dll.Previous = nil
}

func (dll *Dll) DeleteAllNodes() {
	var next *Dll
	for node := dll; node != nil; node = next {
		next = node.Next
		node.RemoveNode()
	}
}

func (dll *Dll) IsDLLEmpty() bool {
	if dll.Next == nil && dll.Previous == nil {
		return true
	}
	return false
}
