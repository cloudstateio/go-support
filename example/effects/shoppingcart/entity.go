package shoppingcart

import (
	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/example/crdt_shoppingcart/shoppingcart"
)

type ShoppingCart struct {
}

func (s *ShoppingCart) HandleCommand(ctx *crdt.CommandContext, name string, cmd interface{}) (reply interface{}, err error) {
	// tag::forward[]
	switch cmd {
	case "Forward":
		switch c := cmd.(type) {
		case *shoppingcart.AddLineItem:
			any, err := encoding.MarshalAny(&shoppingcart.Item{
				ProductId: c.GetProductId(),
				Name:      c.GetName(),
				Quantity:  c.GetQuantity(),
			})
			if err != nil {
				return nil, err
			}
			ctx.Forward(&protocol.Forward{
				ServiceName: "example.shoppingcart.HotItemsService",
				CommandName: "ItemAddedToCart",
				Payload:     any,
			})
		}
		return encoding.Empty, nil
	// end::forward[]
	// tag::effect[]
	case "Effect":
		switch c := cmd.(type) {
		case *shoppingcart.AddLineItem:
			any, err := encoding.MarshalAny(&shoppingcart.Item{
				ProductId: c.GetProductId(),
				Name:      c.GetName(),
				Quantity:  c.GetQuantity(),
			})
			if err != nil {
				return nil, err
			}
			ctx.SideEffect(&protocol.SideEffect{
				ServiceName: "example.shoppingcart.HotItemsService",
				CommandName: "ItemAddedToCart",
				Payload:     any,
				Synchronous: true,
			})
		}
		return encoding.Empty, nil
	// end::effect[]
	default:
		return encoding.Empty, nil
	}
}

func (s *ShoppingCart) HandleEvent(ctx *interface{}, event interface{}) error {
	return nil
}
