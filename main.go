package main

import (
	"os"

	"automata-utils/automata"
)

func main() {
	nfa := automata.ReadNfa(os.Stdin)
	nfa = automata.SplitLongWordsInNfa(nfa)
	automata.WriteNfa(os.Stdout, nfa)
}
