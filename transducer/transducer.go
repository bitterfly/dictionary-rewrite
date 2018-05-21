package transducer

import "fmt"

// FTransition is the unique fail transition of each state if the delta function is undefined in the trie
type FTransition struct {
	failWord int
	state    *Node
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
	transitions map[rune]*Node
	output      int
	fTransition *FTransition
}

func NewNode() *Node {
	return &Node{transitions: make(map[rune]*Node), output: -1, fTransition: nil}
}

func (n *Node) processWord(word []rune) *Node {
	if len(word) == 0 {
		return n
	}

	if _, ok := n.transitions[word[0]]; !ok {
		node := NewNode()
		n.transitions[word[0]] = node
		return node.processWord(word[1:])
	} else {
		return n.transitions[word[0]].processWord(word[1:])
	}
}

// Transducer represents the fail transducer which is built in two stages
// first we make a trie from the dictionary
// then we traverse the trie with BFS
type Transducer struct {
	q0      *Node
	outputs []string
	outputStrings []OutputString
}

func (t *Transducer) Print() {
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
		return ""
	}

	if os.s1 == -1 && os.s2 != -1 {
		return t.outputs[os.s2]
	}

	if os.s1 != -1 && os.s2 == -1 {
		return string(rune(os.s1))
	}

	return t.getOutputString(os.s1) + t.getOutputString(os.s2)
}

func (t *Transducer) print(n *Node) {
	for letter, destination := range n.transitions {
		if destination.output == -1 {
			fmt.Printf("%p --- %c:ε ---> %p\n", n, letter, destination)
		} else {
			fmt.Printf("%p --- (%c:ε)[%s] ---> %p\n", n, letter, t.getOutputString(destination.output), destination)
		}
	}
}

func (t *Transducer) walkTransitions(n *Node, letter rune) (*Node, int) {
	if destination, ok := n.transitions[letter]; ok {
		// return epsilon outputString
		return destination, 0
	}

	if n == t.q0 {
		return n, t.newOutputStringFromChar(letter)
	}

	fstate, fword := t.walkTransitions(n.fTransition.state, letter)
	
	return fstate, t.newOutputStringConcatenate(fword, n.fTransition.failWord)
}

type DictionaryRecord struct {
	Input, Output string
}

func NewTransducer(dictionary chan DictionaryRecord) *Transducer {
	t := &Transducer{q0:NewNode(),outputs: make([]string, 0, 1), outputStrings:make([]OutputString, 1)}
	
	//Make the blank output string to be the first
	t.outputStrings[0] = OutputString{s1:-1, s2:-1}

	for record := range dictionary {
		lastState := t.q0.processWord([]rune(record.Input))
		lastState.output = t.newOutputStringFromString(record.Output)
		// fmt.Printf("Appending output: %s\n", t.getOutputString(lastState.output))	
	}


	// Put all reachable states from q1 in the queue and make their failtransition q0
	queue := make([]*Node, 0, len(t.q0.transitions))
	
	for letter, node := range t.q0.transitions {
		if node.output != -1 {
			node.fTransition = &FTransition{state: t.q0, failWord: node.output}
			// fmt.Printf("Adding fail transition from %p to %p with %s\n", node, t.q0, t.getOutputString(node.output))

		} else {
			node.fTransition = &FTransition{state: t.q0, failWord: t.newOutputStringFromChar(letter)}	
			// fmt.Printf("Adding fail transition from %p to %p with %c\n", node, t.q0, letter)
		
		}

		queue = append(queue, node)
	}

	// BFS to construct fail transitions"
	for len(queue) > 0 {
		current := queue[0]

		queue = queue[1:]
		for letter, destination := range current.transitions {
			// fmt.Printf(fmt.Sprintf("Looking up transition (%p, %c)\n", destination, transition.letter))
			if destination.output != -1 {
				// fmt.Printf(fmt.Sprintf("Putting failword: %s\n", *(destination.output)))
				destination.fTransition = &FTransition{state: t.q0, failWord: destination.output}
			} else {
				// fmt.Printf("Walking back from %p with letter %c\n", current.fTransition.state, letter)
				
				fstate, fword := t.walkTransitions(current.fTransition.state, letter)
				
				// fmt.Printf("This is state %p with word %s\n", fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))
				// fmt.Printf("Adding fail transition from %p to %p with %s\n", destination, fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))
				
				destination.fTransition = &FTransition{state: fstate, failWord: t.newOutputStringConcatenate(current.fTransition.failWord, fword)}
			}
			queue = append(queue, destination)
		}
	}

	return t
}

func (t *Transducer) Replace(word []rune) string {
	return t.replace(t.q0, word)
}

func (t *Transducer) replace(n *Node, word []rune) string {
	if len(word) == 0 {
		if n.fTransition == nil {
			// fmt.Printf("End\n")
			return ""
		} else {
			// fmt.Printf("Adding end fail word: %s\n", t.getOutputString(n.fTransition.failWord))
			return t.getOutputString(n.fTransition.failWord) + t.replace(n.fTransition.state, word)
		}
	}

	if destination, ok := n.transitions[word[0]]; ok {
		// fmt.Printf("Adding from trie (letter %c)\n", word[0])
		return t.replace(destination, word[1:])
	}

	if n == t.q0{
		// fmt.Printf("Adding from extra %c\n", word[0])
		return string(word[0]) + t.replace(n, word[1:])
	}

	// fmt.Printf("Adding fail word %s with letter %c\n", t.getOutputString(n.fTransition.failWord), word[0])

	return t.getOutputString(n.fTransition.failWord) + t.replace(n.fTransition.state, word)
}

