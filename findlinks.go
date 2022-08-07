package rapide

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	linkcid "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/traversal/selector"

	"github.com/ipfs/go-cid"
)

// findLinks execute a selector and tries to find all the links it points to.
// it call the callback on each link found.
func findLinks(s selector.Selector, node datamodel.Node, callback func(cid.Cid, selector.Selector) error) error {
	// FIXME: limit recursion depth in this function
	if s == nil {
		return nil // no need to explore further
	}
	switch node.Kind() {
	case datamodel.Kind_Map, datamodel.Kind_List:
		// Iterate recursively
		iterator := selector.NewSegmentIterator(node)
		todo := s.Interests()
		if todo != nil { // not wildcard
			iterator = &segmentsFilterIterator{iterator, todo}
		}
		for !iterator.Done() {
			path, newNode, err := iterator.Next()
			if err != nil {
				return err
			}
			newSelector, err := s.Explore(node, path)
			if err != nil {
				return nil
			}
			err = findLinks(newSelector, newNode, callback)
			if err != nil {
				return err
			}
		}
	case datamodel.Kind_Link:
		// callback link
		link, err := node.AsLink()
		if err != nil {
			panic(fmt.Errorf("assert failed, links kind should never fail as link call: %w", err))
		}
		// FIXME: deal with non CID links
		return callback(link.(linkcid.Link).Cid, s)
	}
	return nil // other scallar value we don't care about
}

var _ selector.SegmentIterator = (*segmentsFilterIterator)(nil)

type segmentsFilterIterator struct {
	orig     selector.SegmentIterator
	segments []datamodel.PathSegment
}

func (it *segmentsFilterIterator) Next() (datamodel.PathSegment, datamodel.Node, error) {
	o := it.orig
	segments := it.segments
	for !o.Done() {
		p, n, err := o.Next()
		if err != nil {
			return datamodel.PathSegment{}, nil, err
		}

		if i := findItem(segments, p); i >= 0 {
			it.segments = removeItem(segments, i)
			return p, n, nil
		}
	}
	return datamodel.PathSegment{}, nil, fmt.Errorf("segment iterator called while done")
}

func (it *segmentsFilterIterator) Done() bool {
	return len(it.segments) == 0 || it.orig.Done()
}
