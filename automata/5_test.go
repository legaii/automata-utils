package automata

import (
	"testing"
)

func TestNfaToRegexp(t *testing.T) {
	nfa := NewNondeterministicFiniteAutomaton()
	q0 := nfa.AddState()
	q1 := nfa.AddState()
	nfa.SetTerminal(q0, true)
	nfa.AddEdge(Edge{From: q0, To: q1, Word: "a"})
	nfa.AddEdge(Edge{From: q1, To: q0, Word: "b"})
	regexp := ConvertNfaToRegexp(nfa)
	expected := "((((a)1*(b)))*(1)(1)*1)*(((a)1*(b)))*(1)(1)*" // same as (ab)*
	if regexp != expected {
		t.Fatalf("wrong regexp: %s ; expected: %s", regexp, expected)
	}
}
