package transducer

import (
	"bytes"
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

func checkNaive(t *Transducer, dict map[string]string, text string) (bool, string, string) {
	buf := new(bytes.Buffer)
	t.StreamReplace(strings.NewReader(text), buf)
	replaced := naiveReplace(dict, []rune(text))
	return buf.String() == replaced, buf.String(), replaced
}
