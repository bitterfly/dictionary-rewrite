package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bitterfly/dictionary-rewrite/transducer"
)

type transitionKey struct {
	fromState int32
	letter    rune
}

type Trie struct {
	outputs     []string
	finalStates []int32
	transitions map[transitionKey]int32
	numStates   int32
}

func (t *Trie) processWord(n int32, word []rune) int32 {
	if len(word) == 0 {
		return n
	}

	if _, ok := t.transitions[transitionKey{n, word[0]}]; !ok {
		newNodeIndex := t.numStates
		t.numStates++
		t.transitions[transitionKey{n, word[0]}] = newNodeIndex
		return t.processWord(newNodeIndex, word[1:])
	}

	return t.processWord(t.transitions[transitionKey{n, word[0]}], word[1:])
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func NewTrie(dictionary []transducer.DictionaryRecord, dictSize int) *Trie {
	defer timeTrack(time.Now(), "NewTrie")
	t := &Trie{outputs: make([]string, dictSize+1), finalStates: make([]int32, dictSize+1), transitions: make(map[transitionKey]int32), numStates: 1}

	for _, record := range dictionary {
		lastStateIndex := t.processWord(0, []rune(record.Input))
		t.finalStates[lastStateIndex] = 1
		t.outputs[lastStateIndex] = record.Output
	}
	return t
}

func readPlain(filename string) ([]rune, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []rune{}, err
	}
	defer func() {
		err = f.Close()
	}()

	text, err := ioutil.ReadAll(f)
	if err != nil {
		return []rune{}, err
	}
	return []rune(string(text)), nil
}

func getDictionary(filename string) ([]transducer.DictionaryRecord, int, error) {
	dictionary := make([]transducer.DictionaryRecord, 0, 1)
	dictSize := 0

	f, err := os.Open(filename)
	if err != nil {
		return nil, -1, err
	}

	defer func() {
		err = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var dicWords []string
	for scanner.Scan() {

		dicWords = strings.SplitN(scanner.Text(), "\t", 2)
		dictSize += len(dicWords[0])
		dictionary = append(dictionary, transducer.DictionaryRecord{Input: dicWords[0], Output: dicWords[1]})
	}

	if err != nil {
		return nil, -1, err
	}

	return dictionary, dictSize, nil
}

func (t *Trie) deltaDefined(state int32, letter rune) bool {
	_, ok := t.transitions[transitionKey{state, letter}]
	return ok
}

func (t *Trie) deltaDestination(state int32, letter rune) int32 {
	destination, _ := t.transitions[transitionKey{state, letter}]
	return destination
}

func (t *Trie) isFinal(state int32) bool {
	return t.finalStates[state] == 1
}

func (t *Trie) replace(text []rune, output io.Writer) {
	defer timeTrack(time.Now(), "StreamReplace")

	outputBuf := bufio.NewWriter(output)

	defer outputBuf.Flush()

	i := 0
	for i < len(text) {
		state := int32(0)
		wordLen := 0
		wordOutput := ""

		for j := i; j < len(text); j++ {
			if !t.deltaDefined(state, text[j]) {
				break
			}

			state = t.deltaDestination(state, text[j])

			if t.isFinal(state) {
				wordLen = j - i + 1
				wordOutput = t.outputs[state]
			}
		}

		if wordLen != 0 {
			outputBuf.WriteString(wordOutput)
			i += wordLen
		} else {
			outputBuf.WriteRune(text[i])
			i++
		}
	}

}

func main() {
	dictionary, dictSize, err := getDictionary(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file.")
		os.Exit(1)
	}
	t := NewTrie(dictionary, dictSize)

	text, _ := readPlain(os.Args[2])
	t.replace(text, os.Stdout)
}
