package action

import (
	"github.com/golang/protobuf/proto"
)

type Entity struct {
	// ServiceName is the fully qualified name of the service that implements
	// this entities interface. Setting it is mandatory.
	ServiceName ServiceName
	// EntityFunc creates a new entity.
	EntityFunc func() EntityHandler
}

type EntityHandler interface {
	HandleCommand(ctx *Context, name string, msg proto.Message) error
}
