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
	// dict := map[string]string{"a": "1", "ab": "2", "abcc": "3", "babc":"4", "c":"5"}

	// dictChan := make(chan transducer.DictionaryRecord)
	// go func() {
	// 	for k, v := range dict {
	// 		dictChan <- transducer.DictionaryRecord{Input:k, Output:v}
	// 	}
	// 	close(dictChan)
	// }()

	// t := transducer.NewTransducer(dictChan)
	// // t.Print()

	// fmt.Printf(fmt.Sprintf("%s, %s\n", "abcbbbabccb", t.Replace([]rune("abcbbbabccb"))))
	// fmt.Printf(fmt.Sprintf("%s, %s\n", "a", t.Replace([]rune("a"))))
	// fmt.Printf(fmt.Sprintf("%s, %s\n", "ab", t.Replace([]rune("ab"))))
	// fmt.Printf(fmt.Sprintf("%s, %s\n", "abcc", t.Replace([]rune("abcc"))))
	// fmt.Printf(fmt.Sprintf("%s, %s\n", "babc", t.Replace([]rune("babc"))))
	// fmt.Printf(fmt.Sprintf("%s, %s\n", "c", t.Replace([]rune("c"))))

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
