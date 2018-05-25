package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"local/ftransducer/transducer"
	"os"
	"strings"
)

func naiveReplace(dict map[string]string, text []rune) string {
	i := 0
	var outputText, currentText string
	for i < len(text) {
		// fmt.Printf("Current string: %s\nOutputString: %s\ni: %d\n=======\n", currentText, outputText, i)
		j := i
		for j <= len(text) {
			// fmt.Printf("(i=%d, j=%d) Searching for string: %s\n", i, j, text[i:j])
			if _, ok := dict[string(text[i:j])]; ok {
				currentText = dict[string(text[i:j])]
			}
			j++
		}

		if currentText == "" {
			outputText += string(text[i])
			i++
		} else {
			outputText += currentText
			i += len(currentText)
		}

		currentText = ""
	}
	return outputText
}

func checkNaive(t *transducer.Transducer, dict map[string]string, text string) (bool, string, string) {
	buf := new(bytes.Buffer)
	t.StreamReplace(strings.NewReader(text), buf)
	replaced := naiveReplace(dict, []rune(text))
	return buf.String() == replaced, buf.String(), replaced
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
func main() {
	dict, err := readJSON(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file.")
		os.Exit(1)
	}
	dictChan := chanFromDict(dict)
	t := transducer.NewTransducer(dictChan)

	f, err := os.Create("/tmp/graph.dot")
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		err = f.Close()
	}()

	t.Print(f)

	text, err := readPlain(os.Args[2])
	if err != nil {
		fmt.Printf("Could not read input text file")
		os.Exit(1)
	}

	ok, first, second := checkNaive(t, dict, text)
	if ok {
		fmt.Printf("%t\n", ok)
	} else {
		fmt.Printf("%t\n%s\n====\n%s\n=====\n", ok, first, second)
	}
	// t.StreamReplace(strings.NewReader(text), os.Stdout)

}
