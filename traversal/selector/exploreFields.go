package selector

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
)

// ExploreFields traverses named fields in a map (or equivalently, struct, if
// traversing on typed/schema nodes) and applies a next selector to the
// reached nodes.
//
// Note that a concept of "ExplorePath" (e.g. "foo/bar/baz") can be represented
// as a set of three nexted ExploreFields selectors, each specifying one field.
// (For this reason, we don't have a special "ExplorePath" feature; use this.)
//
// ExploreFields also works for selecting specific elements out of a list;
// if a "field" is a base-10 int, it will be coerced and do the right thing.
// ExploreIndex or ExploreRange is more appropriate, however, and should be preferred.
type ExploreFields struct {
	selections map[string]Selector
	interests  []PathSegment // keys of above; already boxed as that's the only way we consume them
}

// Interests for ExploreFields are the fields listed in the selector node
func (s ExploreFields) Interests() []PathSegment {
	return s.interests
}

// Explore returns the selector for the given path if it is a field in
// the selector node or nil if not
func (s ExploreFields) Explore(n ipld.Node, p PathSegment) Selector {
	return s.selections[p.String()]
}

// Decide always returns false because this is not a matcher
func (s ExploreFields) Decide(n ipld.Node) bool {
	return false
}

// ParseExploreFields assembles a Selector
// from a ExploreFields selector node
func ParseExploreFields(n ipld.Node, selectorContexts ...SelectorContext) (Selector, error) {
	if n.ReprKind() != ipld.ReprKind_Map {
		return nil, fmt.Errorf("selector spec parse rejected: selector body must be a map")
	}
	fields, err := n.TraverseField(fieldsKey)
	if err != nil {
		return nil, fmt.Errorf("selector spec parse rejected: fields in ExploreFields selector must be present")
	}
	if fields.ReprKind() != ipld.ReprKind_Map {
		return nil, fmt.Errorf("selector spec parse rejected: fields in ExploreFields selector must be a map")
	}
	x := ExploreFields{
		make(map[string]Selector, fields.Length()),
		make([]PathSegment, 0, fields.Length()),
	}
	for itr := fields.MapIterator(); !itr.Done(); {
		kn, v, err := itr.Next()
		if err != nil {
			return nil, fmt.Errorf("error during selector spec parse: %s", err)
		}

		kstr, _ := kn.AsString()
		x.interests = append(x.interests, PathSegmentString{kstr})
		x.selections[kstr], err = ParseSelector(v, selectorContexts...)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}
