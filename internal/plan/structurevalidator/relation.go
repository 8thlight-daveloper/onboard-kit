package structurevalidator

import "sort"

type relation struct {
	identities []string
	edges      map[string]map[string]bool
}

func newRelation(identities []string) relation {
	return relation{identities: identities, edges: map[string]map[string]bool{}}
}

func (graph relation) add(from, to string) {
	if graph.edges[from] == nil {
		graph.edges[from] = map[string]bool{}
	}
	graph.edges[from][to] = true
}

func unionOf(left, right relation) relation {
	union := newRelation(left.identities)
	for _, graph := range []relation{left, right} {
		for from, targets := range graph.edges {
			for to := range targets {
				union.add(from, to)
			}
		}
	}
	return union
}

type componentSearch struct {
	graph      relation
	next       int
	indices    map[string]int
	lowest     map[string]int
	stack      []string
	stacked    map[string]bool
	components [][]string
}

func (graph relation) cyclicComponents() [][]string {
	search := componentSearch{
		graph: graph, indices: map[string]int{}, lowest: map[string]int{}, stacked: map[string]bool{},
	}
	for _, identity := range graph.identities {
		if _, visited := search.indices[identity]; !visited {
			search.visit(identity)
		}
	}
	return search.components
}

func (search *componentSearch) visit(identity string) {
	search.indices[identity] = search.next
	search.lowest[identity] = search.next
	search.next++
	search.stack = append(search.stack, identity)
	search.stacked[identity] = true
	for target := range search.graph.edges[identity] {
		if _, visited := search.indices[target]; !visited {
			search.visit(target)
			search.lowest[identity] = min(search.lowest[identity], search.lowest[target])
		} else if search.stacked[target] {
			search.lowest[identity] = min(search.lowest[identity], search.indices[target])
		}
	}
	if search.lowest[identity] != search.indices[identity] {
		return
	}
	var members []string
	for {
		last := len(search.stack) - 1
		member := search.stack[last]
		search.stack = search.stack[:last]
		search.stacked[member] = false
		members = append(members, member)
		if member == identity {
			break
		}
	}
	if len(members) > 1 {
		sort.Strings(members)
		search.components = append(search.components, members)
	}
}
