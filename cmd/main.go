package main

import (
	"encoding/json"
	"fmt"
	"local/ftransducer/transducer"
	"os"
	"strings"
)

func readJson(filename string) (map[string]string, error) {
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
	dict, err := readJson(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file.")
		os.Exit(1)
	}
	dictChan := chanFromDict(dict)

	t := transducer.NewTransducer(dictChan)
	t.StreamReplace(strings.NewReader("let's 4make 6-pack.\n"), os.Stdout)
	// fmt.Printf("%s\n", t.Replace([]rune("let's make 6-pack")))
}
