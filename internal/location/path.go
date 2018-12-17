package location

// Path represents a Proto location path.
type Path []int32

// Scope adds field-specific scope to this Path.
// The path is scoped by adding a specific proto
// identifier, along with its index.
//
// Adding scope to a proto file's first
// message type, for example, is done by,
//
//  p.Scope(Message, 0)
func (p Path) Scope(typ ID, idx int) Path {
	return p.copyWith(int32(typ), int32(idx))
}

// Target adds the given target to this Path.
// The path receives a target to a specific
// proto type's component by explicitly
// specifying the target identifier.
//
// Targetting a proto file's package
// definition, for example, is done by,
//
//  p.Target(Package)
func (p Path) Target(target ID) Path {
	return p.copyWith(int32(target))
}

// copyWith creates a new Path, appending the
// given elements to those already found in p.
//
// We purposefully make a copy of the Path
// so that we don't accidentally truncate
// previously added elements.
//
// For details, read up at:
// https://blog.golang.org/go-slices-usage-and-internals
func (p Path) copyWith(elems ...int32) Path {
	path := make(Path, len(p)+len(elems))
	for i, e := range p {
		path[i] = e
	}
	for i, e := range elems {
		path[len(p)+i] = e
	}
	return path
}
