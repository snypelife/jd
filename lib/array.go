package jd

import (
	"fmt"
	"reflect"
)

type jsonArray []JsonNode

var _ JsonNode = jsonArray(nil)

func (a jsonArray) Json() string {
	return renderJson(a)
}

func (a1 jsonArray) Equals(n JsonNode) bool {
	a2, ok := n.(jsonArray)
	if !ok {
		return false
	}
	if len(a1) != len(a2) {
		return false
	}
	return reflect.DeepEqual(a1, a2)
}

func (a jsonArray) Diff(n JsonNode) Diff {
	return a.diff(n, Path{})
}

func (a1 jsonArray) diff(n JsonNode, path Path) Diff {
	d := make(Diff, 0)
	a2, ok := n.(jsonArray)
	if !ok {
		// Different types
		e := DiffElement{
			Path:     path.clone(),
			OldValue: a1,
			NewValue: n,
		}
		return append(d, e)
	}
	maxLen := len(a1)
	if len(a1) < len(a2) {
		maxLen = len(a2)
	}
	for i := maxLen - 1; i >= 0; i-- {
		a1Has := i < len(a1)
		a2Has := i < len(a2)
		subPath := append(path.clone(), float64(i))
		if a1Has && a2Has {
			subDiff := a1[i].diff(a2[i], subPath)
			d = append(d, subDiff...)
		}
		if a1Has && !a2Has {
			e := DiffElement{
				Path:     subPath,
				OldValue: a1[i],
				NewValue: voidNode{},
			}
			d = append(d, e)
		}
		if !a1Has && a2Has {
			e := DiffElement{
				Path:     subPath,
				OldValue: voidNode{},
				NewValue: a2[i],
			}
			d = append(d, e)
		}
	}
	return d
}

func (a jsonArray) Patch(d Diff) (JsonNode, error) {
	return patchAll(a, d)
}

func (a jsonArray) patch(pathBehind, pathAhead Path, oldValue, newValue JsonNode) (JsonNode, error) {
	// Base case
	if len(pathAhead) == 0 {
		if !a.Equals(oldValue) {
			return nil, fmt.Errorf(
				"Found %v at %v. Expected %v.",
				a.Json(), pathBehind, oldValue.Json())
		}
		return newValue, nil
	}
	// Recursive case
	pe, ok := pathAhead[0].(float64)
	if !ok {
		return nil, fmt.Errorf(
			"Invalid path element %v. Expected float64.",
			pathAhead[0])
	}
	i := int(pe)
	var nextNode JsonNode = voidNode{}
	if len(a) > i {
		nextNode = a[i]
	}
	patchedNode, err := nextNode.patch(append(pathBehind, pe), pathAhead[1:], oldValue, newValue)
	if err != nil {
		return nil, err
	}
	if isVoid(patchedNode) {
		if i != len(a)-1 {
			return nil, fmt.Errorf(
				"Removal of a non-terminal element of an array.")
		}
		// Delete an element
		return a[:len(a)-1], nil
	}
	if i > len(a) {
		return nil, fmt.Errorf(
			"Addition beyond the terminal elemtn of an array.")
	}
	if i == len(a) {
		// Add an element
		return append(a, patchedNode), nil
	}
	// Replace an element
	a[i] = patchedNode
	return a, nil
}
