package main

import "local/ftransducer/transducer"
import "fmt"

func main() {
	dict := map[string]string{"ab": "1", "abcd": "2", "d": "3"}
	t := transducer.NewTransducer([]rune{'a', 'b', 'c'}, dict)
	t.Print()

	fmt.Printf(fmt.Sprintf("%s\n", t.Replace([]rune("d"))))
}
