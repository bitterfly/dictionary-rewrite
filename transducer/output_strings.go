package transducer

import "bufio"

//=================OutputString====================================
// OutputString describes the following:
// if s1 == -1 and s2 == -1 -> OutputString encodes epsilon
// if s1 != -1 and s2 == -1 -> OutputString encodes an unicode in s1
// if s1 == -1 and s2 != -1 -> OutputString encodes an index of a string in outputs array
// if s1 != -1 and s2 != -1 -> OutputString encodes two other output strings
type OutputString struct {
	s1, s2 int32
}

func (t *Transducer) newOutputStringEpsilon() int32 {
	return 0
}

func (t *Transducer) newOutputStringFromChar(r rune) int32 {
	t.outputStrings = append(t.outputStrings, OutputString{s1: int32(r), s2: -1})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) newOutputStringFromString(s string) int32 {
	t.outputs = append(t.outputs, s)
	t.outputStrings = append(t.outputStrings, OutputString{s1: -1, s2: int32(len(t.outputs) - 1)})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) newOutputStringConcatenate(s1, s2 int32) int32 {
	t.outputStrings = append(t.outputStrings, OutputString{s1: s1, s2: s2})
	return int32(len(t.outputStrings) - 1)
}

func (t *Transducer) processOutputString(s int32, output *bufio.Writer) error {
	var err error
	os := t.outputStrings[s]
	if os.s1 == -1 && os.s2 == -1 {
		return nil
	}

	if os.s1 == -1 && os.s2 != -1 {
		output.WriteString(t.outputs[os.s2])
		return nil
	}

	if os.s1 != -1 && os.s2 == -1 {
		output.WriteRune(rune(os.s1))
		return nil
	}

	err = t.processOutputString(os.s1, output)
	if err != nil {
		return err
	}
	err = t.processOutputString(os.s2, output)
	return err
}

//==================\OutputString====================================
