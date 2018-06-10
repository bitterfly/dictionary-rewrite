package transducer

import (
	"fmt"
	"io"
)

func (t *Transducer) Print(writer io.Writer) {
	fmt.Fprintf(writer, "digraph transducer {\n")
	for i, s := range t.states {
		fmt.Fprintf(writer, "  %d [label=\"\"];\n", i)
		if i == 0 {
			fmt.Fprintf(writer, " %d -> %d [label=\"%s\",color=red];\n", i, s.fTransition.state, t.getOutputString(s.fTransition.failWord))
		}
	}
	for from, to := range t.transitions {
		t.print(from.fromState, to.destState, from.letter, to.nextRune, writer)
	}
	fmt.Fprintf(writer, "}\n")
}

func (t *Transducer) print(from, to int32, letter rune, nextRune rune, writer io.Writer) {
	fmt.Fprintf(writer, " %d -> %d [label=\"%c,%d\"];\n", from, to, letter, int(nextRune))
}

func (t *Transducer) getOutputString(s int32) string {
	os := t.outputStrings[s]
	if os.s1 == -1 && os.s2 == -1 {
		return ""
	}

	if os.s1 == -1 && os.s2 != -1 {
		return t.outputs[os.s2]
	}

	if os.s1 != -1 && os.s2 == -1 {
		return string(rune(os.s1))
	}

	return t.getOutputString(os.s1) + t.getOutputString(os.s2)
}
