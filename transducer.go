package ftransducer

import "fmt"

// FTransition is the unique fail transition of each state if the delta function is undefined in the trie
type FTransition struct {
	failWord string
	state    *Node
}

// Transition type represents a transition in the transducer, where letter is the first tape and fromTrie indicates whether the second
// tape is epsilon (when true) or if it concides with the first (when false).
type Transition struct {
	letter   rune
	fromTrie bool
}

// Node is the node in the ftransducer
type Node struct {
	transitions map[Transition]*Node
	output      *string
	fTransition *FTransition
}

func NewNode() *Node {
	return &Node{transitions: make(map[Transition]*Node), output: nil, fTransition: nil}
}

func (n *Node) processWord(word []rune) *Node {
	if len(word) == 0 {
		return n
	}

	transition := Transition{word[0], true}

	if _, ok := n.transitions[transition]; !ok {
		node := NewNode()
		n.transitions[transition] = node
		node.processWord(word[1:])
	} else {
		return n.transitions[transition].processWord(word[1:])
	}

	return nil
}

// Transducer represents the fail transducer which is built in two stages
// first we make a trie from the dictionary
// then we traverse the trie with BFS
type Transducer struct {
	q0      *Node
	outputs []string
}

func (n *Node) Print() {
	for transition, destination := range n.transitions {
		if transition.fromTrie {
			fmt.Printf(fmt.Sprintf("%p --- %c:ε ---> %p\n", n, transition.letter, destination))
		} else {
			fmt.Printf(fmt.Sprintf("%p --- %c:%c ---> %p\n", n, transition.letter, transition.letter, destination))
		}
		destination.Print()
	}
}

func NewTransducer(alphabet []rune, dictionary map[string]string) *Transducer {
	q0 := NewNode()
	outputs := make([]string, 0, len(dictionary))

	for input, output := range dictionary {
		lastState := q0.processWord([]rune(input))
		outputs = append(outputs, output)
		lastState.output = &outputs[len(outputs)-1]
	}

	t := &Transducer{q0, outputs}
	return t
}
