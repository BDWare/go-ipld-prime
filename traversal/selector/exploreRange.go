package selector

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
)

// ExploreRange traverses a list, and for each element in the range specified,
// will apply a next selector to those reached nodes.
type ExploreRange struct {
	next     Selector // selector for element we're interested in
	start    int
	end      int
	interest []PathSegment // index of element we're interested in
}

// Interests for ExploreRange are all path segments within the iteration range
func (s ExploreRange) Interests() []PathSegment {
	return s.interest
}

// Explore returns the node's selector if
// the path matches an index in the range of this selector
func (s ExploreRange) Explore(n ipld.Node, p PathSegment) Selector {
	if n.ReprKind() != ipld.ReprKind_List {
		return nil
	}
	index, err := p.Index()
	if err != nil {
		return nil
	}
	if index < s.start || index >= s.end {
		return nil
	}
	return s.next
}

// Decide always returns false because this is not a matcher
func (s ExploreRange) Decide(n ipld.Node) bool {
	return false
}

// ParseExploreRange assembles a Selector
// from a ExploreRange selector node
func ParseExploreRange(n ipld.Node, selectorContexts ...SelectorContext) (Selector, error) {
	if n.ReprKind() != ipld.ReprKind_Map {
		return nil, fmt.Errorf("selector spec parse rejected: selector body must be a map")
	}
	startNode, err := n.TraverseField(startKey)
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: start field must be present in ExploreRange selector")
	}
	startValue, err := startNode.AsInt()
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: start field must be a number in ExploreRange selector")
	}
	endNode, err := n.TraverseField(endKey)
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: end field must be present in ExploreRange selector")
	}
	endValue, err := endNode.AsInt()
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: end field must be a number in ExploreRange selector")
	}
	if startValue >= endValue {
		return nil, fmt.Errorf("selector spec parse rejected: end field must be greater than start field in ExploreRange selector")
	}
	next, err := n.TraverseField(nextSelectorKey)
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: next field must be present in ExploreRange selector")
	}
	selector, err := ParseSelector(next, selectorContexts...)
	if err != nil {
		return nil, err
	}
	x := ExploreRange{
		selector,
		startValue,
		endValue,
		make([]PathSegment, 0, endValue-startValue),
	}
	for i := startValue; i < endValue; i++ {
		x.interest = append(x.interest, PathSegmentInt{I: i})
	}
	return x, nil
}
