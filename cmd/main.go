package main

import "local/ftransducer/transducer"
import "fmt"

func main() {
	dict := map[string]string{"a": "1", "ab": "2", "abcc": "3", "babc":"4", "c":"5"}
	t := transducer.NewTransducer([]rune{'a', 'b', 'c'}, dict)
	t.Print()

	fmt.Printf(fmt.Sprintf("%s, %s\n", "abcbbbabccb", t.Replace([]rune("abcbbbabccb"))))
	fmt.Printf(fmt.Sprintf("%s, %s\n", "a", t.Replace([]rune("a"))))
	fmt.Printf(fmt.Sprintf("%s, %s\n", "ab", t.Replace([]rune("ab"))))
	fmt.Printf(fmt.Sprintf("%s, %s\n", "abcc", t.Replace([]rune("abcc"))))
	fmt.Printf(fmt.Sprintf("%s, %s\n", "babc", t.Replace([]rune("babc"))))
	fmt.Printf(fmt.Sprintf("%s, %s\n", "c", t.Replace([]rune("c"))))
}
