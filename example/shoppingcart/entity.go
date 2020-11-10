//
// Copyright 2019 Lightbend Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shoppingcart

import (
	"errors"
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	domain "github.com/cloudstateio/go-support/example/shoppingcart/persistence"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

// A Cloudstate event sourced entity implementing a shopping cart.
// tag::entity-type[]
// tag::entity-state[]
type ShoppingCart struct {
	// our domain object
	cart []*domain.LineItem
}

// end::entity-state[]
// end::entity-type[]

// NewShoppingCart returns a new and initialized instance of the ShoppingCart entity.
// tag::entity-func[]
func NewShoppingCart(eventsourced.EntityID) eventsourced.EntityHandler {
	return &ShoppingCart{
		cart: make([]*domain.LineItem, 0),
	}
}

// end::entity-func[]

// ItemAdded is a event handler function for the ItemAdded event.
// tag::item-added[]
func (sc *ShoppingCart) ItemAdded(added *domain.ItemAdded) error {
	if added.Item.GetName() == "FAIL" {
		return errors.New("boom: forced an unexpected error")
	}
	if item, _ := sc.find(added.Item.ProductId); item != nil {
		item.Quantity += added.Item.Quantity
		return nil
	}
	sc.cart = append(sc.cart, &domain.LineItem{
		ProductId: added.Item.ProductId,
		Name:      added.Item.Name,
		Quantity:  added.Item.Quantity,
	})
	return nil
}

// end::item-added[]

// ItemRemoved is a event handler function for the ItemRemoved event.
func (sc *ShoppingCart) ItemRemoved(removed *domain.ItemRemoved) error {
	if !sc.remove(removed.ProductId) {
		return errors.New("unable to remove product")
	}
	return nil
}

// HandleEvent lets us handle events by ourselves.
// tag::handle-event[]
func (sc *ShoppingCart) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	switch e := event.(type) {
	case *domain.ItemAdded:
		return sc.ItemAdded(e)
	case *domain.ItemRemoved:
		return sc.ItemRemoved(e)
	default:
		return nil
	}
}

// end::handle-event[]

// AddItem implements the AddItem command handling of the shopping cart service.
// tag::add-item[]
func (sc *ShoppingCart) AddItem(ctx *eventsourced.Context, li *AddLineItem) (*empty.Empty, error) {
	if li.GetQuantity() <= 0 {
		return nil, fmt.Errorf("cannot add negative quantity of to item %q", li.GetProductId())
	}
	ctx.Emit(&domain.ItemAdded{
		Item: &domain.LineItem{
			ProductId: li.ProductId,
			Name:      li.Name,
			Quantity:  li.Quantity,
		}})
	return &empty.Empty{}, nil
}

// end::add-item[]

// RemoveItem implements the RemoveItem command handling of the shopping cart service.
func (sc *ShoppingCart) RemoveItem(ctx *eventsourced.Context, li *RemoveLineItem) (*empty.Empty, error) {
	if item, _ := sc.find(li.GetProductId()); item == nil {
		return nil, fmt.Errorf("cannot remove item %s because it is not in the cart", li.GetProductId())
	}
	ctx.Emit(&domain.ItemRemoved{ProductId: li.ProductId})
	return &empty.Empty{}, nil
}

// GetCart implements the GetCart command handling of the shopping cart service.
// tag::get-cart[]
func (sc *ShoppingCart) GetCart(*eventsourced.Context, *GetShoppingCart) (*Cart, error) {
	cart := &Cart{}
	for _, item := range sc.cart {
		cart.Items = append(cart.Items, &LineItem{
			ProductId: item.ProductId,
			Name:      item.Name,
			Quantity:  item.Quantity,
		})
	}
	return cart, nil
}

// end::get-cart[]

// HandleCommand is the entities command handler implemented by the shopping cart.
// tag::handle-command[]
func (sc *ShoppingCart) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (proto.Message, error) {
	switch c := cmd.(type) {
	case *GetShoppingCart:
		return sc.GetCart(ctx, c)
	case *RemoveLineItem:
		return sc.RemoveItem(ctx, c)
	case *AddLineItem:
		return sc.AddItem(ctx, c)
	default:
		return nil, nil
	}
}

// end::handle-command[]

// Snapshot returns the current state of the shopping cart.
// tag::snapshot[]
func (sc *ShoppingCart) Snapshot(*eventsourced.Context) (snapshot interface{}, err error) {
	return &domain.Cart{
		Items: append(make([]*domain.LineItem, 0, len(sc.cart)), sc.cart...),
	}, nil
}

// end::snapshot[]

// HandleSnapshot applies given snapshot to be the current state.
// tag::handle-snapshot[]
func (sc *ShoppingCart) HandleSnapshot(ctx *eventsourced.Context, snapshot interface{}) error {
	switch value := snapshot.(type) {
	case *domain.Cart:
		sc.cart = append(sc.cart[:0], value.Items...)
		return nil
	default:
		return fmt.Errorf("unknown snapshot type: %v", value)
	}
}

// end::handle-snapshot[]

// find finds a product in the shopping cart by productId and returns it as a LineItem.
func (sc *ShoppingCart) find(productID string) (item *domain.LineItem, index int) {
	for i, item := range sc.cart {
		if productID == item.ProductId {
			return item, i
		}
	}
	return nil, 0
}

// remove removes a product from the shopping cart.
// An ok flag is returned to indicate that the product was present and removed.
func (sc *ShoppingCart) remove(productID string) (ok bool) {
	if item, i := sc.find(productID); item != nil {
		// remove and re-slice
		copy(sc.cart[i:], sc.cart[i+1:])
		sc.cart = sc.cart[:len(sc.cart)-1]
		return true
	}
	return false
}
