package compatible

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

type oneofs map[string]*oneof

var _ descriptorProtoGroup = (oneofs)(nil)

func (os oneofs) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for i, o := range os {
		items[i] = o
	}
	return items
}

// oneof represents a *descriptor.OneofDescriptorProto.
type oneof struct {
	path location.Path
	name string
}

var _ descriptorProto = (*oneof)(nil)

func (o *oneof) Name() string        { return o.name }
func (o *oneof) Path() location.Path { return o.path }
func (o *oneof) Type() string        { return fmt.Sprintf("Oneof %q", o.name) }

func newOneof(od *descriptor.OneofDescriptorProto, p location.Path) *oneof {
	return &oneof{
		path: p,
		name: od.GetName(),
	}
}

// checkOneofs verifies that none of the oneof declarations were removed.
func (c *fileChecker) checkOneofs(original, updated oneofs) {
	c.checkRemovedItems(original, updated, location.Name)
}
