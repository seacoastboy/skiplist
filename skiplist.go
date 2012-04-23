// A golang Skip List Implementation.
// https://github.com/huandu/skiplist/
// 
// Copyright 2011, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/skiplist/blob/master/LICENSE

// Package skiplist provides a go implementation for skip list.
// About skip list: http://en.wikipedia.org/wiki/Skip_list
//
// Skip list is basically an ordered set.
// Following code creates a skip list with int key and adds some values.
//     list := skiplist.New(skiplist.Int)
//     
//     // adds some elements
//     list.Set(20, "Hello")
//     list.Set(10, "World")
//     list.Set(40, true)         // value can be any type
//     list.Set(40, 1000)         // replace last element with new value
//     
//     // try to find one
//     e := list.Get(10)          // value is the Element with key 10
//     _ = e.Value.(string)       // it's the "World". remember to do type cast
//     v, ok := list.GetValue(20) // directly get value. ok is false if not exists
//     notFound := list.Get(15)   // returns nil if key is not found
//     
//     // remove element
//     old := list.Remove(40)     // remove found element and returns its pointer
//                                // returns nil if key is not found
//     
//     // re-init list. it will make the list empty.
//     list.Init()
//
// Skip list elements have random number of next pointers. The max number (say
// "max level") is configurable.
//
// The variable skiplist.DefaultMaxLevel is controlling global default.
// Changing it will not affect created lists.
//     skiplist.DefaultMaxLevel = 24  // the global default is 24 at beginning
// Max level of a created list can also be changed even if it's not empty.
//     list.SetMaxLevel(10)
// Remember the side effect when changing this max level value.
// See its wikipedia page for more details.
//     
// Most comparable built-in types are pre-defined in skiplist, including
//     byte []byte float32 float64 int int16 int32 int64 int8
//     rune string uint uint16 uint32 uint64 uint8 uintptr
// Pre-defined compare function name is similar to the type name, e.g.
// skiplist.Float32 is for float32 key. A special case is skiplist.Bytes is for []byte.
// These functions order key from small to big (say "ascending order").
// There are also reserved order functions with name like skiplist.IntReversed.
// For key types out of the pre-defined list, one can write a custom compare function.
//     type GreaterThanFunc func (lhs, rhs interface{}) bool
// Such compare function returns true if lhs > rhs. Note that, if lhs == rhs, compare
// function (let the name is "foo") must act as following.
//     // if lhs == rhs, following expression must be true
//     foo(lhs, rhs) == false && foo(rhs, lhs) == false
package skiplist

import (
	"math/rand"
)

// Creates a new skiplist.
// keyFunc is a func checking the "greater than" logic.
// If k1 equals k2, keyFunc(k1, k2) and keyFunc(k2, k1) must both be false.
// For built-in types, keyFunc can be found in skiplist package.
// For instance, skiplist.Int is for the list with int keys.
// By default, the list with built-in type key is in ascend order.
// The keyFunc named as skiplist.IntReversed is for descend key order list.
func New(keyFunc GreaterThanFunc) *SkipList {
	if DefaultMaxLevel <= 0 {
		panic("skiplist default level must not be zero or negative")
	}

	return &SkipList{
		level:       DefaultMaxLevel,
		elementNode: elementNode{next: make([]*Element, DefaultMaxLevel)},
		keyFunc:     keyFunc,
	}
}

func randLevel(level int) []*Element {
	return make([]*Element, rand.Intn(level)+1)
}

// Resets a skiplist and discards all exists elements.
func (list *SkipList) Init() *SkipList {
	list.next = make([]*Element, list.level)
	list.length = 0
	return list
}

// Gets the first element.
func (list *SkipList) Front() *Element {
	return list.next[0]
}

// Gets list length.
func (list *SkipList) Len() int {
	return list.length
}

// Sets a value in the list with key.
// If the key exists, change element value to the new one.
// Returns new element pointer.
func (list *SkipList) Set(key, value interface{}) *Element {
	var element *Element

	prevs := list.getPrevElementNodes(key)

	// found an element with the same key, replace its value
	if element = prevs[0].next[0]; element != nil && !list.keyFunc(element.key, key) {
		element.Value = value
		return element
	}

	element = &Element{
		elementNode: elementNode{next: randLevel(list.level)},
		key:         key,
		Value:       value,
	}

	for i := range element.next {
		element.next[i], prevs[i].next[i] = prevs[i].next[i], element
	}

	list.length++
	return element
}

// Gets an element.
// Returns element pointer if found, nil if not found.
func (list *SkipList) Get(key interface{}) *Element {
	var prev *elementNode = &list.elementNode
	var next, last *Element

	for i := list.level - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != last && list.keyFunc(key, next.key) {
			prev, next = &next.elementNode, next.next[i]
		}

		last = next
	}

	if last != nil && !list.keyFunc(last.key, key) {
		return last
	}

	return nil
}

// Gets a value. It's a short hand for Get().Value.
// Returns value and its existence status.
func (list *SkipList) GetValue(key interface{}) (interface{}, bool) {
	element := list.Get(key)

	if element == nil {
		return nil, false
	}

	return element.Value, true
}

// Removes an element.
// Returns removed element pointer if found, nil if not found.
func (list *SkipList) Remove(key interface{}) *Element {
	prevs := list.getPrevElementNodes(key)

	// found the element, remove it
	if element := prevs[0].next[0]; element != nil && !list.keyFunc(element.key, key) {
		for k, v := range element.next {
			prevs[k].next[k] = v
		}

		list.length--
		return element
	}

	return nil
}

func (list *SkipList) getPrevElementNodes(key interface{}) []*elementNode {
	var prev *elementNode = &list.elementNode
	var next, last *Element

	prevs := make([]*elementNode, list.level)

	for i := list.level - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != last && list.keyFunc(key, next.key) {
			prev, next = &next.elementNode, next.next[i]
		}

		prevs[i], last = prev, next
	}

	return prevs
}

// Gets current max level value.
func (list *SkipList) MaxLevel() int {
	return list.level
}

// Changes skip list max level.
// If level is not greater than 0, just panic.
func (list *SkipList) SetMaxLevel(level int) (old int) {
	if level <= 0 {
		panic("invalid argument to SetLevel")
	}

	old, list.level = list.level, level

	if old == level {
		return
	}

	if old > level {
		list.next = list.next[:level]
		return
	}

	nils := make([]*Element, level-old)
	list.next = append(list.next, nils...)
	return
}