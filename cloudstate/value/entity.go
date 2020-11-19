package value

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type Entity struct {
	// ServiceName is the fully qualified name of the service that implements
	// this entities interface. Setting it is mandatory.
	ServiceName ServiceName
	// EntityFunc creates a new entity.
	EntityFunc    func(EntityID) EntityHandler
	PersistenceID string
}

type EntityHandler interface {
	HandleCommand(ctx *Context, name string, msg proto.Message) (*any.Any, error)
	HandleState(ctx *Context, state *any.Any) error
}
