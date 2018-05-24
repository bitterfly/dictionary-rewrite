package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"local/ftransducer/transducer"
	"os"
	"strings"
)

func naiveReplace(dict map[string]string, text string) string {
	i := 0
	var outputText, currentText string
	for i < len(text) {
		// fmt.Printf("Current string: %s\nOutputString: %s\ni: %d\n=======\n", currentText, outputText, i)
		j := i
		for j <= len(text) {
			// fmt.Printf("(i=%d, j=%d) Searching for string: %s\n", i, j, text[i:j])
			if _, ok := dict[text[i:j]]; ok {
				currentText = dict[text[i:j]]
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
	return buf.String() == naiveReplace(dict, text), buf.String(), naiveReplace(dict, text)
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

	// t.StreamReplace(strings.NewReader("let's 4make 6-pack.\n"), os.Stdout)
	ok, first, second := checkNaive(t, dict, "let's 4make 6-pack.\n")
	if ok {
		fmt.Printf("True\n")
	} else {
		fmt.Printf("Transducer:\n%s\nNaive:\n%s\n", first, second)
	}

	ok, first, second = checkNaive(t, dict, "What about the penis?\n")
	if ok {
		fmt.Printf("True\n")
	} else {
		fmt.Printf("Transducer:\n%s\nNaive:\n%s\n", first, second)
	}

	ok, first, second = checkNaive(t, dict, "123456\n")
	if ok {
		fmt.Printf("True\n")
	} else {
		fmt.Printf("Transducer:\n%s\nNaive:\n%s\n", first, second)
	}

	ok, first, second = checkNaive(t, dict, "Go go 5 3\n")
	if ok {
		fmt.Printf("True\n")
	} else {
		fmt.Printf("Transducer:\n%s\nNaive:\n%s\n", first, second)
	}
}
