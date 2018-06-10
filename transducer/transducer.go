package transducer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"
)

func (t *Transducer) Print(writer io.Writer) {
	fmt.Fprintf(writer, "digraph transducer {\n")
	for i, n := range t.states {
		t.print(i, n, writer)
	}
	fmt.Fprintf(writer, "}\n")
}

func (t *Transducer) print(i int, n Node, writer io.Writer) {
	fmt.Fprintf(writer, "  \"%d\" [label=\"\"];\n", i)

	for letter, dest := range n.transitions {
		fmt.Fprintf(writer, " \"%d\" -> \"%d\" [label=\"%c\"];\n", i, dest, letter)
		if n.fTransition != nil {
			fmt.Fprintf(writer, " \"%d\" -> \"%d\" [label=\"%s\",color=red];\n", i, n.fTransition.state, t.getOutputString(n.fTransition.failWord))
		}
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

// FTransition is the unique fail transition of each state if the delta function is undefined in the trie
type FTransition struct {
	failWord int32
	state    int32
}

// OutputString describes the following:
// if s1 == -1 and s2 == -1 -> OutputString encodes epsilon
// if s1 != -1 and s2 == -1 -> OutputString encodes an unicode in s1
// if s1 == -1 and s2 != -1 -> OutputString encodes an index of a string in outputs array
// if s1 != -1 and s2 != -1 -> OutputString encodes two other output strings
type OutputString struct {
	s1, s2 int32
}

// Node is the node in the ftransducer
type Node struct {
	transitions map[rune]int32
	output      int32
	fTransition *FTransition
}

func (t *Transducer) NewNode() int32 {
	t.states = append(t.states, Node{transitions: make(map[rune]int32), output: -1, fTransition: nil})
	return int32(len(t.states) - 1)
}

func (t *Transducer) processWord(n int32, word []rune) int32 {
	if len(word) == 0 {
		return n
	}

	if _, ok := t.states[n].transitions[word[0]]; !ok {
		newNodeIndex := t.NewNode()
		t.states[n].transitions[word[0]] = newNodeIndex
		return t.processWord(newNodeIndex, word[1:])
	} else {
		return t.processWord(t.states[n].transitions[word[0]], word[1:])
	}
}

// Transducer represents the fail transducer which is built in two stages
// first we make a trie from the dictionary
// then we traverse the trie with BFS
type Transducer struct {
	states        []Node
	outputs       []string
	outputStrings []OutputString
}

type bla struct {
}

func (t *Transducer) newOutputStringEpsilon() int32 {
	return 0
}

func (t *Transducer) newOutputStringFromChar(r rune) int32 {
	t.outputStrings = append(t.outputStrings, OutputString{s1: int32(r), s2: -1})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) newOutputStringFromString(s string) int32 {
	t.outputs = append(t.outputs, s)
	t.outputStrings = append(t.outputStrings, OutputString{s1: -1, s2: int32(len(t.outputs) - 1)})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) newOutputStringConcatenate(s1, s2 int32) int32 {
	t.outputStrings = append(t.outputStrings, OutputString{s1: s1, s2: s2})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) processOutputString(s int32, output *bufio.Writer) error {
	var err error
	os := t.outputStrings[s]
	if os.s1 == -1 && os.s2 == -1 {
		return nil
	}

	if os.s1 == -1 && os.s2 != -1 {
		// fmt.Printf("Writing output:|%s|\n", t.outputs[os.s2])
		output.WriteString(t.outputs[os.s2])
		return nil
	}

	if os.s1 != -1 && os.s2 == -1 {
		// fmt.Printf("Writing output:|%c|\n", rune(os.s1))
		output.WriteRune(rune(os.s1))
		return nil
	}

	err = t.processOutputString(os.s1, output)
	if err != nil {
		return err
	}
	err = t.processOutputString(os.s2, output)
	return err
}

func (t *Transducer) getOutputString(s int32) string {
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

// returns the index of the the fail state and the index of the fail word
func (t *Transducer) walkTransitions(n int32, letter rune) (int32, int32) {
	if destination, ok := t.states[n].transitions[letter]; ok {
		// return epsilon outputString
		return destination, 0
	}

	if n == 0 {
		return n, t.newOutputStringFromChar(letter)
	}

	fstate, fword := t.walkTransitions(t.states[n].fTransition.state, letter)

	return fstate, t.newOutputStringConcatenate(t.states[n].fTransition.failWord, fword)
}

// DictionaryRecord is used for reading the dictionary in a channel
type DictionaryRecord struct {
	Input, Output string
}

// NewTransducer returns a fail transducer from the given dictionary
func NewTransducer(dictionary chan DictionaryRecord) *Transducer {
	defer timeTrack(time.Now(), "NewTransducer")
	t := &Transducer{states: make([]Node, 0, 1), outputs: make([]string, 0, 1), outputStrings: make([]OutputString, 1)}
	t.NewNode()

	//Make the blank output string to be the first
	t.outputStrings[0] = OutputString{s1: -1, s2: -1}

	for record := range dictionary {
		lastStateIndex := t.processWord(0, []rune(record.Input))
		t.states[lastStateIndex].output = t.newOutputStringFromString(record.Output)
		// fmt.Printf("Appending output: %s\n", t.getOutputString(lastState.output))
	}

	// Put all reachable states from q1 in the queue and make their failtransition q0
	queue := make([]int32, 0, len(t.states[0].transitions))

	for letter, node := range t.states[0].transitions {
		if t.states[node].output != -1 {
			t.states[node].fTransition = &FTransition{state: 0, failWord: t.states[node].output}
			// fmt.Printf("Adding fail transition from %p to %p with %s\n", node, t.q0, t.getOutputString(node.output))

		} else {
			t.states[node].fTransition = &FTransition{state: 0, failWord: t.newOutputStringFromChar(letter)}
			// fmt.Printf("Adding fail transition from %p to %p with %c\n", node, t.q0, letter)

		}

		queue = append(queue, node)
	}

	// BFS to construct fail transitions"
	for len(queue) > 0 {
		current := queue[0]

		queue = queue[1:]
		for letter, destination := range t.states[current].transitions {
			// fmt.Printf(fmt.Sprintf("Looking up transition (%p, %c)\n", destination, transition.letter))
			if t.states[destination].output != -1 {
				// fmt.Printf(fmt.Sprintf("Putting failword: %s\n", *(destination.output)))
				t.states[destination].fTransition = &FTransition{state: 0, failWord: t.states[destination].output}
			} else {
				// fmt.Printf("Walking back from %p with letter %c\n", current.fTransition.state, letter)

				fstate, fword := t.walkTransitions(t.states[current].fTransition.state, letter)

				// fmt.Printf("This is state %p with word %s\n", fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))
				// fmt.Printf("Adding fail transition from %p to %p with %s\n", destination, fstate, t.getOutputString(current.fTransition.failWord) + t.getOutputString(fword))

				t.states[destination].fTransition = &FTransition{state: fstate, failWord: t.newOutputStringConcatenate(t.states[current].fTransition.failWord, fword)}
			}
			queue = append(queue, destination)
		}
	}

	return t
}

// StreamReplace uses the transducer to replace dictionary words from the text from the input stream using the transducer
// and writes the output into the output buffer
func (t *Transducer) StreamReplace(input io.Reader, output io.Writer) error {
	defer timeTrack(time.Now(), "StreamReplace")

	var err error
	inputBuf := bufio.NewReader(input)
	outputBuf := bufio.NewWriter(output)

	defer outputBuf.Flush()
	node := int32(0)

	for {
		letter, _, err := inputBuf.ReadRune()
		inputBuf.UnreadRune()
		letter, _, err = inputBuf.ReadRune()
		// fmt.Printf("Trying letter: %c\n", letter)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if destination, ok := t.states[node].transitions[letter]; ok {
			node = destination
			continue
		}

		if node == 0 {
			// fmt.Printf("Failing with %c from q0 ->|%c|\n", letter, letter)
			_, err = outputBuf.WriteRune(letter)
			if err != nil {
				return err
			}

			continue
		}

		err = t.processOutputString(t.states[node].fTransition.failWord, outputBuf)
		if err != nil {
			return err
		}

		node = t.states[node].fTransition.state
		err = inputBuf.UnreadRune()
		// fmt.Printf("Unreading rune: %c\n", letter)
		if err != nil {
			return err
		}
	}

	err = t.followTransitions(node, outputBuf)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transducer) followTransitions(n int32, output *bufio.Writer) error {
	var err error
	for n != 0 {
		err = t.processOutputString(t.states[n].fTransition.failWord, output)
		if err != nil {
			return err
		}
		n = t.states[n].fTransition.state
	}
	return nil
}
