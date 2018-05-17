package transducer

import "fmt"

// FTransition is the unique fail transition of each state if the delta function is undefined in the trie
type FTransition struct {
	failWord int
	state    *Node
}

// Transition type represents a transition in the transducer, where letter is the first tape and fromTrie indicates whether the second
// tape is epsilon (when true) or if it concides with the first (when false).
type Transition struct {
	letter   rune
	fromTrie bool
}


// if s1 == -1 and s2 == -1 -> OutputString encodes epsilon
// if s1 != -1 and s2 == -1 -> OutputString encodes an unicode in s1
// if s1 == -1 and s2 != -1 -> OutputString encodes an index of a string in outputs array
// if s1 != -1 and s2 != -1 -> OutputString encodes two other output strings   
type OutputString struct {
	s1, s2 int
}

// Node is the node in the ftransducer
type Node struct {
	transitions map[Transition]*Node
	output      int
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
		n.output = -1
		n.transitions[transition] = node
		return node.processWord(word[1:])
	} else {
		return n.transitions[transition].processWord(word[1:])
	}
}

// Transducer represents the fail transducer which is built in two stages
// first we make a trie from the dictionary
// then we traverse the trie with BFS
type Transducer struct {
	q0      *Node
	alphabet []rune
	outputs []string
	outputStrings []OutputString
}

func (t *Transducer) Print() {
	fmt.Printf("Map: \n")
	for k, v := range CHECK {
		fmt.Printf("%s - %d\n", k, v)
	}
	t.print(t.q0)
}

func (t *Transducer) newOutputStringEpsilon()int {
	return 0
}

func (t *Transducer) newOutputStringFromChar(r rune) int {
	t.outputStrings = append(t.outputStrings, OutputString{s1:int(r), s2:-1})
	return len(t.outputStrings) -1
}

func (t *Transducer) newOutputStringFromString(s string) int {
	t.outputs = append(t.outputs, s)
	t.outputStrings = append(t.outputStrings, OutputString{s1:-1, s2:len(t.outputs) - 1})
	return len(t.outputStrings) -1
}

func (t *Transducer) newOutputStringConcatenate(s1, s2 int) int {
	t.outputStrings = append(t.outputStrings, OutputString{s1:s1, s2:s2})
	return len(t.outputStrings) -1	
}

func (t *Transducer) getOutputString(s int) string {
	os := t.outputStrings[s]
	if os.s1 == -1 && os.s2 == -1 {
		return "ε"
	}

	if os.s1 == -1 && os.s2 != -1 {
		return t.outputs[os.s1]
	}

	if os.s1 != -1 && os.s2 == -1 {
		return string(rune(os.s1))
	}

	return t.getOutputString(os.s1) + t.getOutputString(os.s2)
}

func (t *Transducer) print(n *Node) {
	for transition, destination := range n.transitions {
		if transition.fromTrie {
			if destination.output == -1 {
				fmt.Printf("%p --- %c:ε ---> %p\n", n, transition.letter, destination)
			} else {
				fmt.Printf("%p --- (%c:ε)[%s] ---> %p\n", n, transition.letter, t.getOutputString(destination.output), destination)
			}
		} else {
			if destination.output == -1 {
				fmt.Printf("%p --- %c:%c ---> %p\n", n, transition.letter, transition.letter, destination)
			} else {
				fmt.Printf("%p --- (%c:%c)[%s] ---> %p\n", n, transition.letter, transition.letter, t.getOutputString(destination.output), destination)
			}
		}
		if destination != n {
			t.print(destination)
		}
	}
}

func (n *Node) walkTransitions(letter rune) (*Node, int) {
	if destination, ok := n.transitions[Transition{letter, true}]; ok {
		// return epsilon outputString
		return destination, 0
	}

	if destination, ok := n.transitions[Transition{letter, false}]; ok {
		return destination, t.newOutputStringFromChar(letter)
	
	}

	fstate, fword := n.fTransition.state.walkTransitions(letter)
	
	return fstate, t.newOutputStringConcatenate(fword, n.fTransition.failWord)
}

func NewTransducer(alphabet []rune, dictionary map[string]string) *Transducer {
	t := &Transducer{q0:NewNode(), alphabet:alphabet, outputs: make([]string, 0, len(dictionary)), outputStrings:make([]OutputString, 1)}
	
	//Make the blank output string to be the first
	t.outputStrings[0] = OutputString{s1:-1, s2:-1}

	for input, output := range dictionary {
		lastState := t.q0.processWord([]rune(input))
		lastState.output = t.newOutputStringFromString(output)
		fmt.Printf("Appending output: %s\n", t.getOutputString(lastState.output))	
	}


	// Put all reachable states from q1 in the queue and make their failtransition q0
	queue := make([]*Node, 0, len(t.q0.transitions))
	
	for transition, node := range t.q0.transitions {
		if node.output != -1 {
			node.fTransition = &FTransition{state: t.q0, failWord: node.output}
			fmt.Printf("Adding fail transition from %p to %p with %s\n", node, t.q0, t.getOutputString(node.output))

		} else {
			node.fTransition = &FTransition{state: t.q0, failWord: t.newOutputStringFromChar(transition.letter)}	
			fmt.Printf("Adding fail transition from %p to %p with %s\n", node, t.q0, t.getOutputString(transition.letter))
		
		}

		queue = append(queue, node)
	}


	// Loop q0 with every missing letter in the alphabet
	// TODO: Remove this and check if state is q0
	for _, letter := range t.alphabet {
		if _, ok := t.q0.transitions[Transition{letter, true}]; !ok {
			t.q0.transitions[Transition{letter, false}] = t.q0
		}
	}


	// BFS to construct fail transitions"
	for len(queue) > 0 {
		current := queue[0]

		queue = queue[1:]
		for transition, destination := range current.transitions {
			// fmt.Printf(fmt.Sprintf("Looking up transition (%p, %c)\n", destination, transition.letter))
			if destination.output != -1 {
				// fmt.Printf(fmt.Sprintf("Putting failword: %s\n", *(destination.output)))
				destination.fTransition = &FTransition{state: q0, failWord: destination.output}
			} else {
				fmt.Printf("Walking back from %p with letter %c\n", current.fTransition.state, transition.letter)
				
				fstate, fword := current.fTransition.state.walkTransitions(transition.letter)
				
				fmt.Printf("This is state %p with word %s\n", fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))
				fmt.Printf("Adding fail transition from %p to %p with %s\n", destination, fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))
				
				destination.fTransition = &FTransition{state: fstate, failWord: t.newOutputStringConcatenate(current.fTransition.failWord, fword)}
			}
			queue = append(queue, destination)
		}
	}

	return t
}

func (t *Transducer) Replace(word []rune) string {
	return t.q0.replace(word)
}

func (n *Node) replace(word []rune) string {
	if len(word) == 0 {
		if n.fTransition == nil {
			fmt.Printf("End\n")
			return ""
		} else {
			fmt.Printf("Adding end fail word: %s\n", n.fTransition.failWord)
			return n.fTransition.failWord + n.fTransition.state.replace(word)
		}
	}

	if destination, ok := n.transitions[Transition{word[0], true}]; ok {
		fmt.Printf("Adding from trie (letter %c)\n", word[0])
		return destination.replace(word[1:])
	}

	if destination, ok := n.transitions[Transition{word[0], false}]; ok {
		fmt.Printf("Adding from extra %c\n", word[0])
	
		return string(word[0]) + destination.replace(word[1:])
	}

	fmt.Printf("Adding fail word %s with letter %c\n", n.fTransition.failWord, word[0])

	return n.fTransition.failWord + n.fTransition.state.replace(word)
}

