// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

// Package gtrie provides a trie implementation based on a minimal acyclic
// finite-state automaton.
package gtrie

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

type nodeId int

type nodeIdGen struct {
	id nodeId
}

func (g *nodeIdGen) next() (next nodeId) {
	next = g.id
	g.id++
	return
}

// Represents a transition in the acyclic finite-state automaton. Each
// transition has one label and leads to one node.
type Transition struct {
	Child *Node
	Label rune
}

// Represents a node in the acyclic finite-state automaton.
type Node struct {
	id          nodeId
	Terminal    bool
	Transitions []Transition
}

// Checks whether the node has children.
func (n *Node) HasChildren() bool {
	return len(n.Transitions) > 0
}

// Checks whether the node has a child for the given letter.
func (n *Node) HasChild(letter rune) bool {
	return n.GetChild(letter) != nil
}

// Retrieves the child for the given letter. Returns nil if there is no child
// for this letter.
func (n *Node) GetChild(letter rune) (child *Node) {
	transitions := n.Transitions
	finder := func(i int) bool { return transitions[i].Label >= letter }
	// It is possible to use a binary search here because we know, by
	// construction, that the transitions are sorted by their labels.
	index := sort.Search(len(transitions), finder)
	if index < len(transitions) && transitions[index].Label == letter {
		child = transitions[index].Child
	}
	return
}

// Whether the node recognizes the given suffix. A suffix is accepted if there
// exists a path from the current node to a final node labeled with the suffix
// elements.
func (n *Node) Accepts(suffix string) bool {
	letters := []rune(suffix)
	current := n
	for i := 0; current != nil && i < len(letters); i++ {
		current = current.GetChild(letters[i])
	}
	return current != nil && current.Terminal
}

// Gets the number of nodes in the given automaton.
func Size(node *Node) int {
	ids := make(map[nodeId]bool)
	queue := []*Node{node}
	for len(queue) > 0 {
		node = queue[0]
		queue = queue[1:]
		ids[node.id] = true
		for _, t := range node.Transitions {
			queue = append(queue, t.Child)
		}
	}
	return len(ids)
}

func newNode(idGen *nodeIdGen) *Node {
	return &Node{id: idGen.next()}
}

func addTransition(node *Node, child *Node, letter rune) {
	node.Transitions = append(node.Transitions, Transition{child, letter})
}

func addChild(node *Node, letter rune, idGen *nodeIdGen) (child *Node) {
	child = node.GetChild(letter)
	if child == nil {
		child = newNode(idGen)
		addTransition(node, child, letter)
	}
	return
}

func getLastChild(node *Node) *Node {
	t := node.Transitions
	return t[len(t)-1].Child
}

func setLastChild(node *Node, last *Node) {
	t := node.Transitions
	t[len(t)-1].Child = last
}

type eqClass struct {
	terminal bool
	children string
}

// Obtains the equivalence class for this node, knowing that two nodes p and
// q belongs to the same class if and only if:
//  1. they are either both final or both nonfinal; and
//  2. they have the same number of outgoing transitions; and
//  3. corresponding outgoing transitions have the same labels; and
//  4. corresponding transitions lead to the same states.
func getEquivalenceClass(node *Node) (class eqClass) {
	children := []string{}
	for _, t := range node.Transitions {
		child := string(t.Label) + ":" + strconv.Itoa(int(t.Child.id))
		children = append(children, child)
	}
	class.children = strings.Join(children, ";")
	class.terminal = node.Terminal
	return
}

type registry struct {
	// Mapping from equivalence class to node.
	eqv map[eqClass]*Node
	// Set of nodes that are registered.
	nodes map[*Node]bool
}

func newRegistery() (reg *registry) {
	reg = new(registry)
	reg.eqv = make(map[eqClass]*Node)
	reg.nodes = make(map[*Node]bool)
	return
}

func (r *registry) find(class eqClass) *Node {
	return r.eqv[class]
}

func (r *registry) register(class eqClass, node *Node) {
	r.eqv[class] = node
	r.nodes[node] = true
}

func (r *registry) registered(node *Node) bool {
	return r.nodes[node]
}

// Creates an acyclic finite-state automaton from a sorted list of words and
// returns the root node. Words can contain any unicode chararcters. An error
// will be returned if the list of words is not lexicographically sorted.
func Create(words []string) (automaton *Node, err error) {
	reg := newRegistery()
	idGen := new(nodeIdGen)
	automaton = newNode(idGen)

	if !sort.StringsAreSorted(words) {
		err = errors.New("the words are not sorted")
		return
	}

	for _, word := range words {
		insertWord(word, automaton, reg, idGen)
	}

	replaceOrRegister(automaton, reg)
	return
}

func insertWord(word string, automaton *Node, reg *registry, idGen *nodeIdGen) {
	letters := []rune(word)
	var last *Node

	if len(letters) == 0 {
		return
	}

	// Find last common state.
	for current := automaton; current != nil && len(letters) > 0; {
		last = current
		current = last.GetChild(letters[0])
		if current != nil {
			letters = letters[1:]
		}
	}

	// Minimize.
	if last.HasChildren() {
		replaceOrRegister(last, reg)
	}

	// Add suffix.
	for len(letters) > 0 {
		last = addChild(last, letters[0], idGen)
		letters = letters[1:]
	}

	last.Terminal = true
}

func replaceOrRegister(node *Node, reg *registry) {
	var child = getLastChild(node)

	if reg.registered(child) {
		return
	}

	if child.HasChildren() {
		replaceOrRegister(child, reg)
	}

	class := getEquivalenceClass(child)

	if eq := reg.find(class); eq != nil {
		setLastChild(node, eq)
	} else {
		reg.register(class, child)
	}
}
