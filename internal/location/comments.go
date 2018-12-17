package location

// Comments represents a Proto location's comment details.
type Comments struct {
	// Leading is the comment that exists before
	// a proto type's definition.
	//
	//  /* I'm a leading comment. */
	//  message Foo {}
	//
	Leading string

	// Trailing is the comment that exists
	// immediately after a proto type's definition.
	//
	//  message Foo {} /* I'm a trailing comment. */
	//
	Trailing string

	// LeadingDetached are comments that exists
	// before a proto type's definition, separated
	// by at least one newline.
	//
	//  /* I'm a leading detached comment. */
	//
	//  message Foo {}
	//
	LeadingDetached []string
}
