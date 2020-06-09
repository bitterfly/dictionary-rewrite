package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/bitterfly/ftransducer/transducer"
	"log"
	"time"
	"os"
	"strings"
)


type transitionKey struct {
	fromState int32
	letter    rune
}

type Trie struct {
	outputs         []string
	transitions     map[transitionKey]int32
	num_states	int32
}

func (t *Trie) processWord(n int32, word []rune) int32 {
	if len(word) == 0 {
		return n
	}

	if _, ok := t.transitions[transitionKey{n, word[0]}]; !ok {
		newNodeIndex := t.num_states
		t.num_states++
		t.transitions[transitionKey{n, word[0]}] = newNodeIndex
		return t.processWord(newNodeIndex, word[1:])
	}

	return t.processWord(t.transitions[transitionKey{n, word[0]}], word[1:])
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func NewTrie(dictionary chan transducer.DictionaryRecord) *Trie {
	defer timeTrack(time.Now(), "NewTransducer")
	t := &Trie{outputs: make([]string, 0, 1), transitions: make(map[transitionKey]int32), num_states: 1}
	t.outputs[0] = ""

	for record := range dictionary {
		lastStateIndex := t.processWord(0, []rune(record.Input))
		t.outputs[lastStateIndex] = record.Output

	}
	return t
}

func readPlain(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
	}()

	text, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func readJSON(filename string) (map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
	}()

	decoder := json.NewDecoder(f)
	var dict map[string]string
	err = decoder.Decode(&dict)
	if err != nil {
		return nil, err
	}
	return dict, nil
}

func chanFromDict(dict map[string]string) chan transducer.DictionaryRecord {
	dictChan := make(chan transducer.DictionaryRecord)
	go func() {
		for k, v := range dict {
			dictChan <- transducer.DictionaryRecord{Input: k, Output: v}
		}
		close(dictChan)
	}()

	return dictChan
}

func chanFromFile(filename string) (chan transducer.DictionaryRecord, error) {
	dictChan := make(chan transducer.DictionaryRecord, 2000)
	var err error
	go func() {
		f, err := os.Open(filename)
		if err != nil {
			return
		}

		defer func() {
			err = f.Close()
		}()

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)

		i := 0

		var dicWords []string
		for scanner.Scan() {
			if i%100000 == 0 {
				log.Printf("%d ", i)
			}
			i++
			dicWords = strings.SplitN(scanner.Text(), "\t", 2)
			dictChan <- transducer.DictionaryRecord{Input: dicWords[0], Output: dicWords[1]}
		}
		close(dictChan)
	}()

	if err != nil {
		return nil, err
	}

	return dictChan, nil
}

func (t *Trie) StreamReplace(input io.Reader, output io.Writer) error {
	defer timeTrack(time.Now(), "StreamReplace")

	var err error
	inputBuf := bufio.NewReader(input)
	string text =  inputBuf.Text()
	outputBuf := bufio.NewWriter(output)

	defer outputBuf.Flush()
	node := int32(0)

	for {
		letter, _, err := inputBuf.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if destination, ok := t.transitions[transitionKey{node, letter}]; ok {
			node = destination.destState
			continue
		}

		if node == 0 {
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


func main() {
	dictChan, err := chanFromFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file.")
		os.Exit(1)
	}
	t := NewTrie(dictChan)
	fmt.Printf("%s\n", t.)
//	if len(os.Args) > 2 {
//		text, err := readPlain(os.Args[2])
//		if err != nil {
//			fmt.Printf("Could not read input text file")
//			os.Exit(1)
//		}
//		t.StreamReplace(strings.NewReader(text), os.Stdout)
//	} else {
//		t.StreamReplace(os.Stdin, os.Stdout)
//	}
//
//	t.PrintStates()
//
//	// t.Print(os.Stdout)

}
