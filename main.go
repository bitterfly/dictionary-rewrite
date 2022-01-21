package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bitterfly/dictionary-rewrite/transducer"
)

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

func main() {
	dictChan, err := chanFromFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file.")
		os.Exit(1)
	}
	t := transducer.NewTransducer(dictChan)

	if len(os.Args) > 2 {
		text, err := readPlain(os.Args[2])
		if err != nil {
			fmt.Printf("Could not read input text file")
			os.Exit(1)
		}
		t.StreamReplace(strings.NewReader(text), os.Stdout)
	} else {
		t.StreamReplace(os.Stdin, os.Stdout)
	}

	t.PrintStates()

	// t.Print(os.Stdout)

}
