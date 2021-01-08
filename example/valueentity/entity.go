package valueentity

import (
	"errors"
	"fmt"
	"sort"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/cloudstate/value"
	domain "github.com/cloudstateio/go-support/example/valueentity/persistence"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
)

// A Cloudstate event sourced entity implementing a shopping cart.
type ShoppingCart struct {
	// our domain object
	cart []*domain.LineItem
}

// NewShoppingCart returns a new and initialized instance of the ShoppingCart entity.
func NewShoppingCart(value.EntityID) value.EntityHandler {
	return &ShoppingCart{
		cart: make([]*domain.LineItem, 0),
	}
}

type sortedCart []*domain.LineItem

func (s sortedCart) Len() int           { return len(s) }
func (s sortedCart) Less(i, j int) bool { return s[i].ProductId < s[j].ProductId }
func (s sortedCart) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// AddItem implements the AddItem command handling of the shopping cart service.
func (sc *ShoppingCart) AddItem(ctx *value.Context, item *AddLineItem) (*any.Any, error) {
	if item.GetQuantity() <= 0 {
		return nil, protocol.ClientError{Err: fmt.Errorf("cannot add negative quantity of to item %q", item.GetProductId())}
	}

	if i, _ := sc.find(item.ProductId); i != nil {
		i.Quantity += item.Quantity
		sort.Sort(sortedCart(sc.cart))
		c := &domain.Cart{Items: sc.cart}
		if err := ctx.Update(encoding.MarshalAny(c)); err != nil {
			return nil, err
		}
		return encoding.MarshalAny(&empty.Empty{})
	}
	sc.cart = append(sc.cart, &domain.LineItem{
		ProductId: item.ProductId,
		Name:      item.Name,
		Quantity:  item.Quantity,
	})
	sort.Sort(sortedCart(sc.cart))
	c := &domain.Cart{Items: sc.cart}
	if err := ctx.Update(encoding.MarshalAny(c)); err != nil {
		return nil, err
	}
	return encoding.MarshalAny(&empty.Empty{})
}

// RemoveItem implements the RemoveItem command handling of the shopping cart service.
func (sc *ShoppingCart) RemoveItem(ctx *value.Context, item *RemoveLineItem) (*any.Any, error) {
	if item, _ := sc.find(item.GetProductId()); item == nil {
		return nil, protocol.ClientError{Err: fmt.Errorf("cannot remove item %s because it is not in the cart", item.GetProductId())}
	}
	if !sc.remove(item.ProductId) {
		return nil, protocol.ClientError{errors.New("unable to remove product")}
	}
	sort.Sort(sortedCart(sc.cart))
	c := &domain.Cart{Items: sc.cart}
	if err := ctx.Update(encoding.MarshalAny(c)); err != nil {
		return nil, err
	}
	return encoding.MarshalAny(&empty.Empty{})
}

// GetCart implements the GetCart command handling of the shopping cart service.
func (sc *ShoppingCart) GetCart(*value.Context, *GetShoppingCart) (*any.Any, error) {
	cart := &Cart{}
	for _, item := range sc.cart {
		cart.Items = append(cart.Items, &LineItem{
			ProductId: item.ProductId,
			Name:      item.Name,
			Quantity:  item.Quantity,
		})
	}
	return encoding.MarshalAny(cart)
}

// HandleCommand is the entities command handler implemented by the shopping cart.
func (sc *ShoppingCart) HandleCommand(ctx *value.Context, name string, cmd proto.Message) (*any.Any, error) {
	switch c := cmd.(type) {
	case *GetShoppingCart:
		return sc.GetCart(ctx, c)
	case *RemoveLineItem:
		return sc.RemoveItem(ctx, c)
	case *AddLineItem:
		return sc.AddItem(ctx, c)
	case *RemoveShoppingCart:
		ctx.Delete()
		sc.cart = nil
		return encoding.MarshalAny(&empty.Empty{})
	default:
		return nil, nil
	}
}

func (sc *ShoppingCart) HandleState(ctx *value.Context, state *any.Any) error {
	return nil
}

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
