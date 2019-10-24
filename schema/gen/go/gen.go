package gengo

import (
	"fmt"
	"io"
)

// you'll find a file in this package per kind
//  (schema level kind, not data model level reprkind)...
// sparse cross-product with their representation strategy (more or less)
//  (it's more... idunnoyet.  hopefully we have implstrats and reprstrats,
//   and those combine over an interface so it's not a triple cross product...
//    and hopefully that interface is nodebuilder,
//     because I dunno why it wouldn't be unless we goof on perf somehow).

// TypedNodeGenerator declares a standard names for a bunch of methods for generating
// code for our schema types.  There's still numerous places where other casts
// to more specific interfaces will be required (so, technically, it's not a
// very powerful interface; it's not so much that the abstractions leak as that
// the floodgates are outright open), but this at least forces consistency onto
// the parts where we can.
//
// All Emit{foo} methods should emit one trailing and one leading linebreak, or,
// nothing (e.g. string kinds don't need to produce a dummy map iterator, so
// such a method can just emit nothing, and the extra spacing between sections
// shouldn't accumulate).
//
// None of these methods return error values because we panic in this package.
//
type TypedNodeGenerator interface {
	// wip note: hopefully imports are a constant.  if not, we'll have to curry something with the writer.

	// -- the typed.Node.Type method and vars -->

	EmitTypedNodeMethodType(io.Writer) // these emit dummies for now

	// -- all node methods -->
	//   (and note that the nodeBuilder for this one should be the "semantic" one,
	//     e.g. it *always* acts like a map for structs, even if the repr is different.)

	nodeGenerator

	// -- and the representation and its node and nodebuilder -->

	EmitTypedNodeMethodRepresentation(io.Writer)
	GetRepresentationNodeGen() nodeGenerator // includes transitively the matched nodebuilderGenerator
}

type nodeGenerator interface {
	EmitNodeType(io.Writer)
	EmitNodeMethodReprKind(io.Writer)
	EmitNodeMethodLookupString(io.Writer)
	EmitNodeMethodLookup(io.Writer)
	EmitNodeMethodLookupIndex(io.Writer)
	EmitNodeMethodLookupSegment(io.Writer)
	EmitNodeMethodMapIterator(io.Writer)  // also iterator itself
	EmitNodeMethodListIterator(io.Writer) // also iterator itself
	EmitNodeMethodLength(io.Writer)
	EmitNodeMethodIsUndefined(io.Writer)
	EmitNodeMethodIsNull(io.Writer)
	EmitNodeMethodAsBool(io.Writer)
	EmitNodeMethodAsInt(io.Writer)
	EmitNodeMethodAsFloat(io.Writer)
	EmitNodeMethodAsString(io.Writer)
	EmitNodeMethodAsBytes(io.Writer)
	EmitNodeMethodAsLink(io.Writer)
	EmitNodeMethodNodeBuilder(io.Writer)

	GetNodeBuilderGen() nodebuilderGenerator
}

type nodebuilderGenerator interface {
	EmitNodebuilderType(io.Writer)
	EmitNodebuilderConstructor(io.Writer)

	EmitNodebuilderMethodCreateMap(io.Writer)
	EmitNodebuilderMethodAmendMap(io.Writer)
	EmitNodebuilderMethodCreateList(io.Writer)
	EmitNodebuilderMethodAmendList(io.Writer)
	EmitNodebuilderMethodCreateNull(io.Writer)
	EmitNodebuilderMethodCreateBool(io.Writer)
	EmitNodebuilderMethodCreateInt(io.Writer)
	EmitNodebuilderMethodCreateFloat(io.Writer)
	EmitNodebuilderMethodCreateString(io.Writer)
	EmitNodebuilderMethodCreateBytes(io.Writer)
	EmitNodebuilderMethodCreateLink(io.Writer)

	// TODO we'll soon also need all the child-nb-getters here too. // they're hucked in the CreateMap/CreateList methods for now.
}

func emitFileHeader(w io.Writer) {
	fmt.Fprintf(w, "package whee\n\n")
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, "\tipld \"github.com/ipld/go-ipld-prime\"\n")
	fmt.Fprintf(w, "\t\"github.com/ipld/go-ipld-prime/impl/typed\"\n")
	fmt.Fprintf(w, "\t\"github.com/ipld/go-ipld-prime/schema\"\n")
	fmt.Fprintf(w, ")\n\n")
}

// enums will have special methods
// maps will have special methods (namely, well typed getters
