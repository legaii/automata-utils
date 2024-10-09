package automata

import (
	"hash/fnv"
	"sort"
)

type DeterministicFiniteAutomaton struct {
	start     State
	terminals []bool
	edges     []map[uint8]State
}

func NewDeterministicFiniteAutomaton() *DeterministicFiniteAutomaton {
	return new(DeterministicFiniteAutomaton)
}

func (dfa *DeterministicFiniteAutomaton) StateCount() int {
	return len(dfa.edges)
}

func (dfa *DeterministicFiniteAutomaton) AddState() State {
	dfa.terminals = append(dfa.terminals, false)
	dfa.edges = append(dfa.edges, make(map[uint8]State))
	return State(dfa.StateCount() - 1)
}

func (dfa *DeterministicFiniteAutomaton) Start() State {
	return dfa.start
}

func (dfa *DeterministicFiniteAutomaton) SetStart(state State) {
	dfa.start = state
}

func (dfa *DeterministicFiniteAutomaton) IsTerminal(state State) bool {
	return dfa.terminals[state]
}

func (dfa *DeterministicFiniteAutomaton) SetTerminal(state State, isTerminal bool) {
	dfa.terminals[state] = isTerminal
}

func (dfa *DeterministicFiniteAutomaton) Edges(state State) []Edge {
	edges := make([]Edge, 0, len(dfa.edges[state]))
	for char, to := range dfa.edges[state] {
		edges = append(edges, Edge{
			From: state,
			To:   to,
			Word: string(char),
		})
	}

	sort.Slice(edges, func(i int, j int) bool {
		return edges[i].Word[0] < edges[j].Word[0]
	})
	return edges
}

func (dfa *DeterministicFiniteAutomaton) AddEdge(edge Edge) {
	if len(edge.Word) != 1 {
		panic("word length must be 1")
	}
	dfa.edges[edge.From][edge.Word[0]] = edge.To
}

func (dfa *DeterministicFiniteAutomaton) DeleteEdge(edge Edge) bool {
	if len(edge.Word) != 1 {
		panic("word length must be 1")
	}
	_, exists := dfa.edges[edge.From][edge.Word[0]]
	delete(dfa.edges[edge.From], edge.Word[0])
	return exists
}

var _ INondeterministicFiniteAutomaton = (*DeterministicFiniteAutomaton)(nil)

type MetaState struct {
	states []bool
	id     State
}

func (metaState MetaState) calcHash() uint64 {
	hash := fnv.New64a()
	for _, isInSet := range metaState.states {
		if isInSet {
			hash.Write([]byte{0})
		} else {
			hash.Write([]byte{1})
		}
	}
	return hash.Sum64()
}

func ConvertNfaToDfa(nfa1 INondeterministicFiniteAutomaton) *DeterministicFiniteAutomaton {
	nfa2 := ConvertNfaToNfaWithWordsOfLen1(nfa1)
	dfa := NewDeterministicFiniteAutomaton()
	metaStates := make(map[uint64]MetaState)

	var dfs func(metaFrom MetaState) State
	dfs = func(metaFrom MetaState) State {
		hash := metaFrom.calcHash()
		if metaState, visited := metaStates[hash]; visited {
			return metaState.id
		}
		metaFrom.id = dfa.AddState()
		metaStates[hash] = metaFrom

		metaEdges := make(map[uint8][]bool)
		for stateFrom, isInSet := range metaFrom.states {
			if !isInSet {
				continue
			}
			if nfa2.IsTerminal(State(stateFrom)) {
				dfa.SetTerminal(metaFrom.id, true)
			}
			for _, edge := range nfa2.Edges(State(stateFrom)) {
				char := edge.Word[0]
				if _, alreadyInited := metaEdges[char]; !alreadyInited {
					metaEdges[char] = make([]bool, nfa2.StateCount())
				}
				metaEdges[char][edge.To] = true
			}
		}

		for char, statesTo := range metaEdges {
			metaTo := MetaState{
				states: statesTo,
			}
			metaTo.id = dfs(metaTo)
			dfa.AddEdge(Edge{
				From: metaFrom.id,
				To:   metaTo.id,
				Word: string(char),
			})
		}
		return metaFrom.id
	}

	metaStart := MetaState{
		states: make([]bool, nfa2.StateCount()),
	}
	metaStart.states[nfa2.Start()] = true
	dfs(metaStart)
	return dfa
}

func MakeComplete(dfa *DeterministicFiniteAutomaton, alphabet []uint8) {
	dumpState := dfa.AddState()
	for _, edges := range dfa.edges {
		for _, char := range alphabet {
			if _, edgeExists := edges[char]; !edgeExists {
				edges[char] = dumpState
			}
		}
	}
}

func CdfaComplementInplace(cdfa *DeterministicFiniteAutomaton) {
	for state := range cdfa.StateCount() {
		cdfa.SetTerminal(State(state), !cdfa.IsTerminal(State(state)))
	}
}
