package traversal_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/warpfork/go-wish"

	ipld "github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/encoding/dagjson"
	"github.com/ipld/go-ipld-prime/fluent"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	"github.com/ipld/go-ipld-prime/traversal"
	"github.com/ipld/go-ipld-prime/traversal/selector"
)

/* Remember, we've got the following fixtures in scope:
var (
	leafAlpha, leafAlphaLnk         = encode(fnb.CreateString("alpha"))
	leafBeta, leafBetaLnk           = encode(fnb.CreateString("beta"))
	middleMapNode, middleMapNodeLnk = encode(fnb.CreateMap(func(mb fluent.MapBuilder, knb fluent.NodeBuilder, vnb fluent.NodeBuilder) {
		mb.Insert(knb.CreateString("foo"), vnb.CreateBool(true))
		mb.Insert(knb.CreateString("bar"), vnb.CreateBool(false))
		mb.Insert(knb.CreateString("nested"), vnb.CreateMap(func(mb fluent.MapBuilder, knb fluent.NodeBuilder, vnb fluent.NodeBuilder) {
			mb.Insert(knb.CreateString("alink"), vnb.CreateLink(leafAlphaLnk))
			mb.Insert(knb.CreateString("nonlink"), vnb.CreateString("zoo"))
		}))
	}))
	middleListNode, middleListNodeLnk = encode(fnb.CreateList(func(lb fluent.ListBuilder, vnb fluent.NodeBuilder) {
		lb.Append(vnb.CreateLink(leafAlphaLnk))
		lb.Append(vnb.CreateLink(leafAlphaLnk))
		lb.Append(vnb.CreateLink(leafBetaLnk))
		lb.Append(vnb.CreateLink(leafAlphaLnk))
	}))
	rootNode, rootNodeLnk = encode(fnb.CreateMap(func(mb fluent.MapBuilder, knb fluent.NodeBuilder, vnb fluent.NodeBuilder) {
		mb.Insert(knb.CreateString("plain"), vnb.CreateString("olde string"))
		mb.Insert(knb.CreateString("linkedString"), vnb.CreateLink(leafAlphaLnk))
		mb.Insert(knb.CreateString("linkedMap"), vnb.CreateLink(middleMapNodeLnk))
		mb.Insert(knb.CreateString("linkedList"), vnb.CreateLink(middleListNodeLnk))
	}))
)
*/

// covers traverse using a variety of selectors.
// all cases here use one already-loaded Node; no link-loading exercised.

func TestTraverse(t *testing.T) {
	ssb := selector.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
	t.Run("traverse selecting true should visit the root", func(t *testing.T) {
		err := traversal.Traverse(fnb.CreateString("x"), selector.Matcher{}, func(tp traversal.TraversalProgress, n ipld.Node) error {
			Wish(t, n, ShouldEqual, fnb.CreateString("x"))
			Wish(t, tp.Path.String(), ShouldEqual, ipld.Path{}.String())
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
	})
	t.Run("traverse selecting true should visit only the root and no deeper", func(t *testing.T) {
		err := traversal.Traverse(middleMapNode, selector.Matcher{}, func(tp traversal.TraversalProgress, n ipld.Node) error {
			Wish(t, n, ShouldEqual, middleMapNode)
			Wish(t, tp.Path.String(), ShouldEqual, ipld.Path{}.String())
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
	})
	t.Run("traverse selecting fields should work", func(t *testing.T) {
		ss := ssb.ExploreFields(func(efsb selector.ExploreFieldsSpecBuilder) {
			efsb.Insert("foo", ssb.Matcher())
			efsb.Insert("bar", ssb.Matcher())
		})
		s, err := ss.Selector()
		Require(t, err, ShouldEqual, nil)
		var order int
		err = traversal.Traverse(middleMapNode, s, func(tp traversal.TraversalProgress, n ipld.Node) error {
			switch order {
			case 0:
				Wish(t, n, ShouldEqual, fnb.CreateBool(true))
				Wish(t, tp.Path.String(), ShouldEqual, "foo")
			case 1:
				Wish(t, n, ShouldEqual, fnb.CreateBool(false))
				Wish(t, tp.Path.String(), ShouldEqual, "bar")
			}
			order++
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
		Wish(t, order, ShouldEqual, 2)
	})
	t.Run("traverse selecting fields recursively should work", func(t *testing.T) {
		ss := ssb.ExploreFields(func(efsb selector.ExploreFieldsSpecBuilder) {
			efsb.Insert("foo", ssb.Matcher())
			efsb.Insert("nested", ssb.ExploreFields(func(efsb selector.ExploreFieldsSpecBuilder) {
				efsb.Insert("nonlink", ssb.Matcher())
			}))
		})
		s, err := ss.Selector()
		Require(t, err, ShouldEqual, nil)
		var order int
		err = traversal.Traverse(middleMapNode, s, func(tp traversal.TraversalProgress, n ipld.Node) error {
			switch order {
			case 0:
				Wish(t, n, ShouldEqual, fnb.CreateBool(true))
				Wish(t, tp.Path.String(), ShouldEqual, "foo")
			case 1:
				Wish(t, n, ShouldEqual, fnb.CreateString("zoo"))
				Wish(t, tp.Path.String(), ShouldEqual, "nested/nonlink")
			}
			order++
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
		Wish(t, order, ShouldEqual, 2)
	})
	t.Run("traversing across nodes should work", func(t *testing.T) {
		ss := ssb.ExploreRecursive(3, ssb.ExploreUnion(
			ssb.Matcher(),
			ssb.ExploreAll(ssb.ExploreRecursiveEdge()),
		))
		s, err := ss.Selector()
		var order int
		err = traversal.TraversalProgress{
			Cfg: &traversal.TraversalConfig{
				LinkLoader: func(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
					return bytes.NewBuffer(storage[lnk]), nil
				},
			},
		}.Traverse(middleMapNode, s, func(tp traversal.TraversalProgress, n ipld.Node) error {
			switch order {
			case 0:
				Wish(t, n, ShouldEqual, middleMapNode)
				Wish(t, tp.Path.String(), ShouldEqual, "")
			case 1:
				Wish(t, n, ShouldEqual, fnb.CreateBool(true))
				Wish(t, tp.Path.String(), ShouldEqual, "foo")
			case 2:
				Wish(t, n, ShouldEqual, fnb.CreateBool(false))
				Wish(t, tp.Path.String(), ShouldEqual, "bar")
			case 3:
				Wish(t, n, ShouldEqual, fnb.CreateMap(func(mb fluent.MapBuilder, knb fluent.NodeBuilder, vnb fluent.NodeBuilder) {
					mb.Insert(knb.CreateString("alink"), vnb.CreateLink(leafAlphaLnk))
					mb.Insert(knb.CreateString("nonlink"), vnb.CreateString("zoo"))
				}))
				Wish(t, tp.Path.String(), ShouldEqual, "nested")
			case 4:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "nested/alink")
			case 5:
				Wish(t, n, ShouldEqual, fnb.CreateString("zoo"))
				Wish(t, tp.Path.String(), ShouldEqual, "nested/nonlink")
			}
			order++
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
		Wish(t, order, ShouldEqual, 6)
	})
	t.Run("traversing lists should work", func(t *testing.T) {
		ss := ssb.ExploreRange(0, 3, ssb.Matcher())
		s, err := ss.Selector()
		var order int
		err = traversal.TraversalProgress{
			Cfg: &traversal.TraversalConfig{
				LinkLoader: func(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
					return bytes.NewBuffer(storage[lnk]), nil
				},
			},
		}.Traverse(middleListNode, s, func(tp traversal.TraversalProgress, n ipld.Node) error {
			switch order {
			case 0:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "0")
			case 1:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "1")
			case 2:
				Wish(t, n, ShouldEqual, fnb.CreateString("beta"))
				Wish(t, tp.Path.String(), ShouldEqual, "2")
			}
			order++
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
		Wish(t, order, ShouldEqual, 3)
	})
	t.Run("multiple layers of link traversal should work", func(t *testing.T) {
		ss := ssb.ExploreFields(func(efsb selector.ExploreFieldsSpecBuilder) {
			efsb.Insert("linkedList", ssb.ExploreAll(ssb.Matcher()))
			efsb.Insert("linkedMap", ssb.ExploreRecursive(3, ssb.ExploreFields(func(efsb selector.ExploreFieldsSpecBuilder) {
				efsb.Insert("foo", ssb.Matcher())
				efsb.Insert("nonlink", ssb.Matcher())
				efsb.Insert("alink", ssb.Matcher())
				efsb.Insert("nested", ssb.ExploreRecursiveEdge())
			})))
		})
		s, err := ss.Selector()
		var order int
		err = traversal.TraversalProgress{
			Cfg: &traversal.TraversalConfig{
				LinkLoader: func(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
					return bytes.NewBuffer(storage[lnk]), nil
				},
			},
		}.Traverse(rootNode, s, func(tp traversal.TraversalProgress, n ipld.Node) error {
			switch order {
			case 0:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedList/0")
			case 1:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedList/1")
			case 2:
				Wish(t, n, ShouldEqual, fnb.CreateString("beta"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedList/2")
			case 3:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedList/3")
			case 4:
				Wish(t, n, ShouldEqual, fnb.CreateBool(true))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedMap/foo")
			case 5:
				Wish(t, n, ShouldEqual, fnb.CreateString("zoo"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedMap/nested/nonlink")
			case 6:
				Wish(t, n, ShouldEqual, fnb.CreateString("alpha"))
				Wish(t, tp.Path.String(), ShouldEqual, "linkedMap/nested/alink")
			}
			order++
			return nil
		})
		Wish(t, err, ShouldEqual, nil)
		Wish(t, order, ShouldEqual, 7)
	})
}
