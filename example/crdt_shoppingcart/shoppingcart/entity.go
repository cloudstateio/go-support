package shoppingcart

import (
	"errors"
	"reflect"
	"runtime"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

//go:generate protoc --go_out=plugins=grpc,paths=source_relative:./example/crdt_shoppingcart/shoppingcart --proto_path=protobuf/protocol --proto_path=protobuf/frontend --proto_path=protobuf/frontend/cloudstate --proto_path=protobuf/proxy --proto_path=example/crdt_shoppingcart/shoppingcart proto

type ShoppingCart struct {
	items *crdt.ORMap
}

func NewShoppingCart(id crdt.EntityID) crdt.EntityHandler {
	return &ShoppingCart{}
}

func (s *ShoppingCart) getCart() (*Cart, error) {
	items := &Cart{}
	for _, state := range s.items.Values() {
		var item LineItem
		if err := encoding.DecodeStruct(state.GetLwwregister().GetValue(), &item); err != nil {
			return nil, err
		}
		items.Items = append(items.Items, &item)
	}
	return items, nil
}

func Command(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// tag::command-handling-getcart-0[]
// tag::add-item-0[]
func (s *ShoppingCart) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	// end::command-handling-getcart-0[]
	// end::add-item-0[]

	// tag::watch-cart[]
	switch name {
	case "WatchCart":
		ctx.ChangeFunc(func(c *crdt.CommandContext) (*any.Any, error) {
			cart, err := s.getCart()
			if err != nil {
				return nil, err
			}
			return encoding.MarshalAny(cart)
		})
		cart, err := s.getCart()
		if err != nil {
			return nil, err
		}
		return encoding.MarshalAny(cart)
	}
	// end::watch-cart[]
	// tag::command-handling-getcart-1[]
	// tag::add-item-1[]
	switch m := msg.(type) {
	// end::add-item-1[]
	case *GetShoppingCart:
		cart, err := s.getCart()
		if err != nil {
			return nil, err
		}
		return encoding.MarshalAny(cart)
	// end::command-handling-getcart-1[]
	// tag::add-item-2[]
	case *AddLineItem:
		if m.GetQuantity() <= 0 {
			return nil, errors.New("cannot add a negative quantity of items")
		}

		key := encoding.String(m.GetProductId())
		reg, err := s.items.LWWRegister(key)
		if err != nil {
			return nil, err
		}
		item, err := encoding.MarshalAny(&LineItem{
			ProductId: m.GetProductId(),
			Name:      m.GetName(),
			Quantity:  m.GetQuantity(),
		})
		if err != nil {
			return nil, err
		}
		if reg != nil {
			reg.Set(item)
		} else {
			reg = crdt.NewLWWRegister(item)
		}
		s.items.Set(key, reg)
		return encoding.Empty, nil
	// end::add-item-2[]
	default:
		return nil, nil
	}
}

// tag::creation[]
func (s *ShoppingCart) Default(ctx *crdt.Context) (crdt.CRDT, error) {
	return crdt.NewORMap(), nil
}

func (s *ShoppingCart) Set(ctx *crdt.Context, state crdt.CRDT) error {
	s.items = state.(*crdt.ORMap)
	return nil
}

// end::creation[]
