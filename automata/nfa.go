package automata

import (
	"io"
	"strconv"
	"strings"

	"automata-utils/ioutils"
)

type State int

func (state State) String() string {
	return strconv.Itoa(int(state))
}

type Edge struct {
	From State
	To   State
	Word string
}

func (edge Edge) String() string {
	return edge.From.String() + "->" + edge.To.String() + " " + edge.Word + "\n"
}

type INondeterministicFiniteAutomaton interface {
	StateCount() int
	AddState() State

	Start() State
	SetStart(State)
	IsTerminal(State) bool
	SetTerminal(State, bool)

	Edges(State) []Edge
	AddEdge(Edge)
	DeleteEdge(Edge) bool
}

type NondeterministicFiniteAutomaton struct {
	start     State
	terminals []bool
	edges     [][]Edge
}

func NewNondeterministicFiniteAutomaton() INondeterministicFiniteAutomaton {
	return new(NondeterministicFiniteAutomaton)
}

func (nfa *NondeterministicFiniteAutomaton) StateCount() int {
	return len(nfa.edges)
}

func (nfa *NondeterministicFiniteAutomaton) AddState() State {
	nfa.terminals = append(nfa.terminals, false)
	nfa.edges = append(nfa.edges, nil)
	return State(nfa.StateCount() - 1)
}

func (nfa *NondeterministicFiniteAutomaton) Start() State {
	return nfa.start
}

func (nfa *NondeterministicFiniteAutomaton) SetStart(state State) {
	nfa.start = state
}

func (nfa *NondeterministicFiniteAutomaton) IsTerminal(state State) bool {
	return nfa.terminals[state]
}

func (nfa *NondeterministicFiniteAutomaton) SetTerminal(state State, isTerminal bool) {
	nfa.terminals[state] = isTerminal
}

func (nfa *NondeterministicFiniteAutomaton) Edges(state State) []Edge {
	return nfa.edges[state]
}

func (nfa *NondeterministicFiniteAutomaton) AddEdge(edge Edge) {
	nfa.edges[edge.From] = append(nfa.edges[edge.From], edge)
}

func (nfa *NondeterministicFiniteAutomaton) DeleteEdge(edge Edge) bool {
	edges := nfa.edges[edge.From]
	for index, currentEdge := range edges {
		if currentEdge == edge {
			edges[index] = edges[len(edges)-1]
			nfa.edges[edge.From] = edges[:len(edges)-1]
			return true
		}
	}
	return false
}

var _ INondeterministicFiniteAutomaton = (*NondeterministicFiniteAutomaton)(nil)

func ReadNfa(reader io.Reader) INondeterministicFiniteAutomaton {
	stateCount := ioutils.Read[int](reader)
	nfa := NewNondeterministicFiniteAutomaton()
	nfa.SetStart(ioutils.Read[State](reader))
	for state := range stateCount {
		nfa.AddState()
		nfa.SetTerminal(State(state), ioutils.Read[bool](reader))
	}
	edgeCount := ioutils.Read[int](reader)
	for range edgeCount {
		from := ioutils.Read[State](reader)
		to := ioutils.Read[State](reader)
		word := ioutils.Read[string](reader)
		if word == "eps" {
			word = ""
		}
		nfa.AddEdge(Edge{
			From: from,
			To:   to,
			Word: word,
		})
	}
	return nfa
}

func WriteNfa(writer io.Writer, nfa INondeterministicFiniteAutomaton) {
	ioutils.Write(writer, "Start: "+nfa.Start().String()+"\n")
	ioutils.Write(writer, "Terminals:")
	for state := range nfa.StateCount() {
		if nfa.IsTerminal(State(state)) {
			ioutils.Write(writer, " "+State(state).String())
		}
	}
	ioutils.Write(writer, "\n")
	for state := range nfa.StateCount() {
		for _, edge := range nfa.Edges(State(state)) {
			ioutils.Write(writer, edge)
		}
	}
}

func SplitLongWordsInNfa(nfa INondeterministicFiniteAutomaton) INondeterministicFiniteAutomaton {
	newNfa := NewNondeterministicFiniteAutomaton()
	for state := range nfa.StateCount() {
		newNfa.SetTerminal(newNfa.AddState(), nfa.IsTerminal(State(state)))
	}
	newNfa.SetStart(nfa.Start())

	for state := range nfa.StateCount() {
		for _, edge := range nfa.Edges(State(state)) {
			if len(edge.Word) <= 1 {
				newNfa.AddEdge(edge)
				continue
			}
			prevState := edge.From
			for i, char := range edge.Word {
				var nextState State
				if i+1 < len(edge.Word) {
					nextState = newNfa.AddState()
				} else {
					nextState = edge.To
				}
				newNfa.AddEdge(Edge{
					From: prevState,
					To:   nextState,
					Word: string(char),
				})
				prevState = nextState
			}
		}
	}
	return newNfa
}

func iterateOverReachableByEmptyWords(
	nfa INondeterministicFiniteAutomaton,
	from State,
	functor func(State),
) {
	visited := make([]bool, nfa.StateCount())
	var dfs func(state State)
	dfs = func(state State) {
		functor(state)
		visited[state] = true
		for _, edge := range nfa.Edges(state) {
			if edge.Word == "" && !visited[edge.To] {
				dfs(edge.To)
			}
		}
	}
	dfs(from)
}

func ConvertNfaToNfaWithWordsOfLen1(nfa1 INondeterministicFiniteAutomaton) INondeterministicFiniteAutomaton {
	nfa2 := SplitLongWordsInNfa(nfa1)
	nfa3 := NewNondeterministicFiniteAutomaton()
	for range nfa2.StateCount() {
		nfa3.AddState()
	}
	nfa3.SetStart(nfa2.Start())

	for from := range nfa2.StateCount() {
		iterateOverReachableByEmptyWords(nfa2, State(from),
			func(to State) {
				if nfa2.IsTerminal(to) {
					nfa3.SetTerminal(State(from), true)
				}
				for _, edge := range nfa2.Edges(to) {
					if edge.Word != "" {
						nfa3.AddEdge(Edge{
							From: State(from),
							To:   edge.To,
							Word: edge.Word,
						})
					}
				}
			},
		)
	}
	return nfa3
}

func ConvertNfaToNfaWithSingleTerminal(nfa INondeterministicFiniteAutomaton) State {
	terminal := nfa.AddState()
	nfa.SetTerminal(terminal, true)
	for state := range nfa.StateCount() {
		if nfa.IsTerminal(State(state)) {
			nfa.SetTerminal(State(state), false)
			nfa.AddEdge(Edge{
				From: State(state),
				To:   terminal,
				Word: "",
			})
		}
	}
	return terminal
}

func ConvertNfaToRegexp(nfa INondeterministicFiniteAutomaton) string {
	terminal := ConvertNfaToNfaWithSingleTerminal(nfa)

	ordinaryStates := make([]State, 0, nfa.StateCount()-2)
	for state := range nfa.StateCount() {
		if State(state) != nfa.Start() && State(state) != terminal {
			ordinaryStates = append(ordinaryStates, State(state))
		}
	}

	for _, curState := range ordinaryStates {
		var loops []Edge
		prevEdges := make(map[State][]Edge)
		nextEdges := make(map[State][]Edge)
		for state := range nfa.StateCount() {
			for _, edge := range nfa.Edges(State(state)) {
				if edge.From == curState && edge.To == curState {
					loops = append(loops, edge)
				} else if edge.To == curState {
					prevEdges[edge.From] = append(prevEdges[edge.From], edge)
				} else if edge.From == curState {
					nextEdges[edge.To] = append(nextEdges[edge.To], edge)
				}
			}
		}
		loop := reduceEdgesToRegexp(nfa, loops)
		prevEdgesReduced := reduceManyEdgesToRegexp(nfa, prevEdges)
		nextEdgesReduced := reduceManyEdgesToRegexp(nfa, nextEdges)

		for from, prevEdge := range prevEdgesReduced {
			for to, nextEdge := range nextEdgesReduced {
				nfa.AddEdge(Edge{
					From: from,
					To:   to,
					Word: "(" + prevEdge + loop + "*" + nextEdge + ")",
				})
			}
		}
	}

	var loopsStart []Edge
	var loopsTerminal []Edge
	var edgesFromStartToTerminal []Edge
	var edgesFromTerminalToStart []Edge
	for _, edge := range nfa.Edges(nfa.Start()) {
		if edge.To == nfa.Start() {
			loopsStart = append(loopsStart, edge)
		} else {
			edgesFromStartToTerminal = append(edgesFromStartToTerminal, edge)
		}
	}
	for _, edge := range nfa.Edges(terminal) {
		if edge.To == terminal {
			loopsTerminal = append(loopsTerminal, edge)
		} else {
			edgesFromTerminalToStart = append(edgesFromTerminalToStart, edge)
		}
	}
	loopStart := reduceEdgesToRegexp(nfa, loopsStart)
	loopTerminal := reduceEdgesToRegexp(nfa, loopsTerminal)
	edgeFromStartToTerminal := reduceEdgesToRegexp(nfa, edgesFromStartToTerminal)
	edgeFromTerminalToStart := reduceEdgesToRegexp(nfa, edgesFromTerminalToStart)

	var regexp strings.Builder
	regexp.WriteString("(")
	regexp.WriteString(loopStart)
	regexp.WriteString("*")
	regexp.WriteString(edgeFromStartToTerminal)
	regexp.WriteString(loopTerminal)
	regexp.WriteString("*")
	regexp.WriteString(edgeFromTerminalToStart)
	regexp.WriteString(")*")
	regexp.WriteString(loopStart)
	regexp.WriteString("*")
	regexp.WriteString(edgeFromStartToTerminal)
	regexp.WriteString(loopTerminal)
	regexp.WriteString("*")
	return regexp.String()
}

func reduceEdgesToRegexp(nfa INondeterministicFiniteAutomaton, edges []Edge) string {
	if len(edges) == 0 {
		return "1"
	}
	var regexp strings.Builder
	regexp.WriteString("(")
	for i, edge := range edges {
		if i > 0 {
			regexp.WriteString("+")
		}
		if edge.Word == "" {
			regexp.WriteString("1")
		} else {
			regexp.WriteString(edge.Word)
		}
		nfa.DeleteEdge(edge)
	}
	regexp.WriteString(")")
	return regexp.String()
}

func reduceManyEdgesToRegexp(nfa INondeterministicFiniteAutomaton, edgesMap map[State][]Edge) map[State]string {
	regexps := make(map[State]string, len(edgesMap))
	for state, edges := range edgesMap {
		regexps[state] = reduceEdgesToRegexp(nfa, edges)
	}
	return regexps
}
