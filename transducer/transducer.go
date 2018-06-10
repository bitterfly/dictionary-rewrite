package transducer

import (
	"bufio"
	"io"
	"log"
	"time"
)

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
	output      int32
	fTransition *FTransition
	firstRune   rune
}

func (t *Transducer) NewNode() int32 {
	t.states = append(t.states, Node{output: -1, fTransition: nil})
	return int32(len(t.states) - 1)
}

func (t *Transducer) transitionBetween(from, to int32, letter rune) {

	oldRune := t.states[from].firstRune

	t.states[to].firstRune = 0
	t.states[from].firstRune = letter
	t.transitions[TransitionKey{from, letter}] = TransitionDestination{destState: to, nextRune: oldRune}
}

func (t *Transducer) processWord(n int32, word []rune) int32 {
	if len(word) == 0 {
		return n
	}

	if _, ok := t.transitions[TransitionKey{n, word[0]}]; !ok {
		newNodeIndex := t.NewNode()
		t.transitionBetween(n, newNodeIndex, word[0])
		return t.processWord(newNodeIndex, word[1:])
	} else {
		return t.processWord(t.transitions[TransitionKey{n, word[0]}].destState, word[1:])
	}
}

// Transducer represents the fail transducer which is built in two stages
// first we make a trie from the dictionary
// then we traverse the trie with BFS
type Transducer struct {
	states        []Node
	outputs       []string
	transitions   map[TransitionKey]TransitionDestination
	outputStrings []OutputString
}

type TransitionKey struct {
	fromState int32
	letter    rune
}

type TransitionDestination struct {
	destState int32
	nextRune  rune
}

// returns the index of the the fail state and the index of the fail word
func (t *Transducer) deltaFGamma(n int32, letter rune) (int32, int32) {
	if destination, ok := t.transitions[TransitionKey{n, letter}]; ok {
		// return epsilon outputString
		return destination.destState, 0
	}

	if n == 0 {
		return n, t.newOutputStringFromChar(letter)
	}

	fstate, fword := t.deltaFGamma(t.states[n].fTransition.state, letter)

	return fstate, t.newOutputStringConcatenate(t.states[n].fTransition.failWord, fword)
}

// DictionaryRecord is used for reading the dictionary in a channel
type DictionaryRecord struct {
	Input, Output string
}

// NewTransducer returns a fail transducer from the given dictionary
func NewTransducer(dictionary chan DictionaryRecord) *Transducer {
	defer timeTrack(time.Now(), "NewTransducer")
	t := &Transducer{states: make([]Node, 0, 1), outputs: make([]string, 0, 1), outputStrings: make([]OutputString, 1), transitions: make(map[TransitionKey]TransitionDestination)}
	t.NewNode()

	//Make the blank output string to be the first
	t.outputStrings[0] = OutputString{s1: -1, s2: -1}

	for record := range dictionary {
		lastStateIndex := t.processWord(0, []rune(record.Input))
		t.states[lastStateIndex].output = t.newOutputStringFromString(record.Output)
		// fmt.Printf("Appending output: %s\n", t.getOutputString(lastState.output))
	}

	// Put all reachable states from q1 in the queue and make their failtransition q0
	queue := make([]int32, 0, 1)

	currentKey := TransitionKey{fromState: 0, letter: t.states[0].firstRune}
	for currentKey.letter != 0 {
		node := t.transitions[currentKey].destState
		nextLetter := t.transitions[currentKey].nextRune
		if t.states[node].output != -1 {
			t.states[node].fTransition = &FTransition{state: 0, failWord: t.states[node].output}
			// fmt.Printf("Adding fail transition from %p to %p with %s\n", node, t.q0, t.getOutputString(node.output))

		} else {
			t.states[node].fTransition = &FTransition{state: 0, failWord: t.newOutputStringFromChar(currentKey.letter)}
			// fmt.Printf("Adding fail transition from %p to %p with %c\n", node, t.q0, letter)

		}

		queue = append(queue, node)
		currentKey.letter = nextLetter
	}
	// BFS to construct fail transitions"

	for len(queue) > 0 {
		currentKey.fromState = queue[0]

		queue = queue[1:]

		currentKey.letter = t.states[currentKey.fromState].firstRune
		for currentKey.letter != 0 {
			destination := t.transitions[currentKey].destState
			nextLetter := t.transitions[currentKey].nextRune
			if t.states[destination].output != -1 {
				t.states[destination].fTransition = &FTransition{state: 0, failWord: t.states[destination].output}
			} else {
				fstate, fword := t.deltaFGamma(t.states[currentKey.fromState].fTransition.state, currentKey.letter)
				t.states[destination].fTransition = &FTransition{state: fstate, failWord: t.newOutputStringConcatenate(t.states[currentKey.fromState].fTransition.failWord, fword)}
			}
			queue = append(queue, destination)
			currentKey.letter = nextLetter
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

		if destination, ok := t.transitions[TransitionKey{node, letter}]; ok {
			node = destination.destState
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

	err = t.followFTransitions(node, outputBuf)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transducer) followFTransitions(n int32, output *bufio.Writer) error {
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
