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

package eventsourced

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/example/shoppingcart"
	domain "github.com/cloudstateio/go-support/example/shoppingcart/persistence"
)

const serviceName = "com.example.shoppingcart.ShoppingCart"

func TestEventsourcingShoppingCart(t *testing.T) {
	s := newServer(t)
	s.newClientConn()
	defer s.teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	edc := protocol.NewEntityDiscoveryClient(s.conn)
	discover, err := edc.Discover(ctx, &protocol.ProxyInfo{
		ProtocolMajorVersion: 0,
		ProtocolMinorVersion: 0,
		ProxyName:            "a-cs-proxy",
		ProxyVersion:         "0.0.0",
		SupportedEntityTypes: []string{protocol.EventSourced, protocol.CRDT},
	})
	if err != nil {
		t.Fatal(err)
	}
	// discovery
	if l := len(discover.GetEntities()); l != 1 {
		t.Fatalf("discover.Entities is:%d, should be: 1", l)
	}
	s.serviceName = discover.GetEntities()[0].GetServiceName()
	t.Run("entity discovery should find the shopping cart service", func(t *testing.T) {
		if s.serviceName != serviceName {
			t.Fatalf("discover.Entities[0].ServiceName is:%v, should be: %s", s.serviceName, serviceName)
		}
	})

	t.Run("calling GetShoppingCart should fail without an init message", func(t *testing.T) {
		p := newProxy(ctx, s)
		r := p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e1", Name: "GetShoppingCart"},
			m: &shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Failure:
		case *entity.EventSourcedStreamOut_Reply:
			t.Fatal("a message should not be allowed to be received without a init message sent before")
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("calling GetShoppingCart with an init message should succeed", func(t *testing.T) {
		p := newProxy(ctx, s)
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    "e2",
		})
		r := p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e2", Name: "GetShoppingCart"},
			m: &shoppingcart.GetShoppingCart{UserId: "user2"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("after added line items, the cart should return the same lines added", func(t *testing.T) {
		p := newProxy(ctx, s)
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    "e3",
		})
		// add line item
		addLineItem := &shoppingcart.AddLineItem{
			UserId: "user1", ProductId: "e-bike-1", Name: "e-Bike", Quantity: 2,
		}
		r := p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e3", Name: "AddLineItem"},
			m: addLineItem,
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			t.Run("the reply should have events", func(t *testing.T) {
				events := m.Reply.GetEvents()
				if got, want := len(events), 1; got != want {
					t.Fatalf("len(events) = %d; want %d", got, want)
				}
				itemAdded := &domain.ItemAdded{}
				err := encoding.UnmarshalAny(events[0], itemAdded)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := itemAdded.Item.ProductId, addLineItem.ProductId; got != want {
					t.Fatalf("itemAdded.Item.ProductId = %s; want; %s", got, want)
				}
				if got, want := itemAdded.Item.Name, addLineItem.Name; got != want {
					t.Fatalf("itemAdded.Item.Name = %s; want; %s", got, want)
				}
				if got, want := itemAdded.Item.Quantity, addLineItem.Quantity; got != want {
					t.Fatalf("itemAdded.Item.Quantity = %d; want; %d", got, want)
				}
			})
			t.Run("the reply should have a snapshot", func(t *testing.T) {
				snapshot := m.Reply.GetSnapshot()
				if snapshot == nil {
					t.Fatalf("snapshot was nil but should not")
				}
				cart := &domain.Cart{}
				err := encoding.UnmarshalAny(snapshot, cart)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := len(cart.Items), 1; got != want {
					t.Fatalf("len(cart.Items) = %d; want: %d", got, want)
				}
				item := cart.Items[0]
				if got, want := item.ProductId, addLineItem.ProductId; got != want {
					t.Fatalf("itemAdded.Item.ProductId = %s; want; %s", got, want)
				}
				if got, want := item.Name, addLineItem.Name; got != want {
					t.Fatalf("itemAdded.Item.Name = %s; want; %s", got, want)
				}
				if got, want := item.Quantity, addLineItem.Quantity; got != want {
					t.Fatalf("itemAdded.Item.Quantity = %d; want; %d", got, want)
				}
			})
		case *entity.EventSourcedStreamOut_Failure:
			p.checkCommandID(m)
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
		// get the shopping cart
		r = p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e3", Name: "GetShoppingCart"},
			m: &shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			payload := m.Reply.GetClientAction().GetReply().GetPayload()
			cart := &shoppingcart.Cart{}
			if err := encoding.UnmarshalAny(payload, cart); err != nil {
				t.Fatal(err)
			}
			if l := len(cart.Items); l != 1 {
				t.Fatalf("len(cart.Items) is: %d, but should be 1", l)
			}
			li, ai := cart.Items[0], addLineItem
			if got, want := li.Quantity, ai.Quantity; got != want {
				t.Fatalf("Quantity = %d; want: %d", got, want)
			}
			if got, want := li.Name, ai.Name; got != want {
				t.Fatalf("Name = %s; want: %s", got, want)
			}
			if got, want := li.ProductId, ai.ProductId; got != want {
				t.Fatalf("ProductId = %s; want: %s", got, want)
			}
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("the entity should have consistent state after a context failure", func(t *testing.T) {
		p := newProxy(ctx, s)
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    "e3",
		})
		// add line item
		addLineItem := &shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-1", Name: "e-Bike", Quantity: 2}
		r := p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e3", Name: "AddLineItem"},
			m: addLineItem,
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
		case *entity.EventSourcedStreamOut_Failure:
			p.checkCommandID(m)
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}

		// add BOOM line item
		r = p.sendRecvCmd(command{
			c: &protocol.Command{EntityId: "e3", Name: "AddLineItem"},
			m: &shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-1", Name: "FAIL", Quantity: 4},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			t.Fatalf("expected failure but got: %+v", m.Reply)
		case *entity.EventSourcedStreamOut_Failure:
			p.checkCommandID(m)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
		// get the shopping cart
		_, err = p.sendRecvCmdErr(command{
			c: &protocol.Command{
				EntityId: "e2",
				Name:     "GetShoppingCart",
			},
			m: &shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		if err == nil {
			t.Fatal(errors.New("expected error"))
		}
		if err != io.EOF {
			t.Fatal(errors.New("expected io.EOF error"))
		}
	})

	t.Run("an init message with an initial snapshot should initialise an entity", func(t *testing.T) {
		p := newProxy(ctx, s)
		cart := &domain.Cart{Items: make([]*domain.LineItem, 0)}
		cart.Items = append(cart.Items, &domain.LineItem{
			ProductId: "e-bike-2", Name: "Cross", Quantity: 3,
		})
		cart.Items = append(cart.Items, &domain.LineItem{
			ProductId: "e-bike-3", Name: "Cross TWO", Quantity: 1,
		})
		cart.Items = append(cart.Items, &domain.LineItem{
			ProductId: "e-bike-4", Name: "City", Quantity: 5,
		})
		any, err := encoding.MarshalAny(cart)
		if err != nil {
			t.Fatal(err)
		}
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    "e9",
			Snapshot: &entity.EventSourcedSnapshot{
				SnapshotSequence: 0,
				Snapshot:         any,
			},
		})
		r := p.sendRecvCmd(command{
			c: &protocol.Command{
				EntityId: "e9",
				Name:     "GetShoppingCart",
			},
			m: &shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			reply := &domain.Cart{}
			err := encoding.UnmarshalAny(m.Reply.GetClientAction().GetReply().GetPayload(), reply)
			if err != nil {
				t.Fatal(err)
			}
			if l := len(reply.Items); l != 3 {
				t.Fatalf("len(cart.Items) is: %d, but should be 1", l)
			}
			for i := 0; i < len(reply.Items); i++ {
				li := reply.Items[i]
				ci := cart.Items[i]
				if got, want := li.Quantity, ci.Quantity; got != want {
					t.Fatalf("Quantity = %d; want: %d", got, want)
				}
				if got, want := li.Name, ci.Name; got != want {
					t.Fatalf("Name = %s; want: %s", got, want)
				}
				if got, want := li.ProductId, ci.ProductId; got != want {
					t.Fatalf("ProductId = %s; want: %s", got, want)
				}
			}
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("an initialised entity with an event sent should return its applied state", func(t *testing.T) {
		p := newProxy(ctx, s)
		// init
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    "e20",
		})
		// send an event
		lineItem := &domain.LineItem{ProductId: "e-bike-100", Name: "AMP 100", Quantity: 1}
		event, err := encoding.MarshalAny(&domain.ItemAdded{Item: lineItem})
		if err != nil {
			t.Fatal(err)
		}
		p.sendEvent(&entity.EventSourcedEvent{Sequence: 0, Payload: event})
		r := p.sendRecvCmd(command{
			&protocol.Command{EntityId: "e20", Name: "GetShoppingCart"},
			&shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			cart := &shoppingcart.Cart{}
			if err := encoding.UnmarshalAny(m.Reply.GetClientAction().GetReply().GetPayload(), cart); err != nil {
				t.Fatal(err)
			}
			if l := len(cart.Items); l != 1 {
				t.Fatalf("len(cart.Items) is: %d, but should be 1", l)
			}
			li := cart.Items[0]
			if got, want := li.Quantity, lineItem.Quantity; got != want {
				t.Fatalf("Quantity = %d; want: %d", got, want)
			}
			if got, want := li.Name, lineItem.Name; got != want {
				t.Fatalf("Name = %s; want: %s", got, want)
			}
			if got, want := li.ProductId, lineItem.ProductId; got != want {
				t.Fatalf("ProductId = %s; want: %s", got, want)
			}
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("adding negative quantity should fail, and leave the entity in a conistent state", func(t *testing.T) {
		p := newProxy(ctx, s)
		entityID := "e23"
		userID := "user1"
		p.sendInit(&entity.EventSourcedInit{
			ServiceName: s.serviceName,
			EntityId:    entityID,
		})
		// add line item
		add := []command{
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: userID, ProductId: "e-bike-1", Name: "e-Bike", Quantity: 1},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: userID, ProductId: "e-bike-2", Name: "e-Bike 2", Quantity: 2},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: userID, ProductId: "e-bike-2", Name: "e-Bike 2", Quantity: -1},
			},
		}
		for i, cmd := range add {
			r := p.sendRecvCmd(cmd)
			switch m := r.Message.(type) {
			case *entity.EventSourcedStreamOut_Reply:
				p.checkCommandID(m)
				if i < 2 && m.Reply.ClientAction.GetFailure() != nil {
					t.Fatalf("unexpected ClientAction failure: %+v", m)
				}
				if i == 2 && m.Reply.ClientAction.GetFailure() == nil {
					t.Fatalf("expected ClientAction failure: %+v", m)
				}
			case *entity.EventSourcedStreamOut_Failure:
				p.checkCommandID(m)
				t.Fatalf("expected reply but got: %+v", m.Failure)
			default:
				t.Fatalf("unexpected message: %+v", m)
			}
		}
		// get the shopping chart
		r := p.sendRecvCmd(command{
			&protocol.Command{EntityId: entityID, Name: "GetShoppingCart"},
			&shoppingcart.GetShoppingCart{UserId: userID},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			cart := &shoppingcart.Cart{}
			if err := encoding.UnmarshalAny(m.Reply.GetClientAction().GetReply().GetPayload(), cart); err != nil {
				t.Fatal(err)
			}
			if got, want := len(cart.Items), 2; got != want {
				t.Fatalf("len(cart.Items) is: %d; want: %d", got, want)
			}
			for _, i := range cart.Items {
				if i.ProductId == "e-bike-2" && i.Quantity != 2 {
					t.Fatal("cart is in an inconsistent state after an emit-fail sequence")
				}
			}
			// li := cart.Items[0]
			// if got, want := li.Quantity, lineItem.Quantity; got != want {
			//	t.Fatalf("Quantity = %d; want: %d", got, want)
			// }
			// if got, want := li.Name, lineItem.Name; got != want {
			//	t.Fatalf("Name = %s; want: %s", got, want)
			// }
			// if got, want := li.ProductId, lineItem.ProductId; got != want {
			//	t.Fatalf("ProductId = %s; want: %s", got, want)
			// }
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("removing a non existent item should fail", func(t *testing.T) {
		p := newProxy(ctx, s)
		entityID := "e24"
		p.sendInit(&entity.EventSourcedInit{ServiceName: s.serviceName, EntityId: entityID})
		// add line item
		add := []command{
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-1", Name: "e-Bike", Quantity: 1},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-2", Name: "e-Bike 2", Quantity: 2},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "RemoveLineItem"},
				&shoppingcart.RemoveLineItem{UserId: "user1", ProductId: "e-bike-1"},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "RemoveLineItem"},
				&shoppingcart.RemoveLineItem{UserId: "user1", ProductId: "e-bike-1"},
			},
		}
		for i, cmd := range add {
			r := p.sendRecvCmd(cmd)
			switch m := r.Message.(type) {
			case *entity.EventSourcedStreamOut_Reply:
				p.checkCommandID(m)
				if i <= 2 && m.Reply.ClientAction.GetFailure() != nil {
					t.Fatalf("unexpected ClientAction failure: %+v", m)
				}
				if i == 3 && m.Reply.ClientAction.GetFailure() == nil {
					t.Fatalf("expected ClientAction failure: %+v", m)
				}
			case *entity.EventSourcedStreamOut_Failure:
				p.checkCommandID(m)
				t.Fatalf("expected reply but got: %+v", m.Failure)
			default:
				t.Fatalf("unexpected message: %+v", m)
			}
		}
	})

	t.Run("adding and removing line items should result in a consistent state", func(t *testing.T) {
		p := newProxy(ctx, s)
		entityID := "e22"
		p.sendInit(&entity.EventSourcedInit{ServiceName: s.serviceName, EntityId: entityID})
		// add line item
		add := []command{
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-1", Name: "e-Bike", Quantity: 1},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-2", Name: "e-Bike 2", Quantity: 2},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-3", Name: "e-Bike 3", Quantity: 3},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "AddLineItem"},
				&shoppingcart.AddLineItem{UserId: "user1", ProductId: "e-bike-3", Name: "e-Bike 3", Quantity: 4},
			},
		}
		for _, cmd := range add {
			r := p.sendRecvCmd(cmd)
			switch m := r.Message.(type) {
			case *entity.EventSourcedStreamOut_Reply:
				p.checkCommandID(m)
				if m.Reply.ClientAction.GetFailure() != nil {
					t.Fatalf("unexpected ClientAction failure: %+v", m)
				}
			case *entity.EventSourcedStreamOut_Failure:
				p.checkCommandID(m)
				t.Fatalf("expected reply but got: %+v", m.Failure)
			default:
				t.Fatalf("unexpected message: %+v", m)
			}
		}
		// get the shopping cart
		r := p.sendRecvCmd(command{
			&protocol.Command{
				EntityId: entityID,
				Name:     "GetShoppingCart",
			},
			&shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			payload := m.Reply.GetClientAction().GetReply().GetPayload()
			cart := &shoppingcart.Cart{}
			if err := encoding.UnmarshalAny(payload, cart); err != nil {
				t.Fatal(err)
			}
			if got, want := len(cart.Items), 3; got != want {
				t.Fatalf("len(cart.Items) = %d; want: %d ", got, want)
			}
			eBike3Count := int32(0)
			for _, c := range cart.Items {
				if c.ProductId == "e-bike-3" {
					eBike3Count += c.Quantity
				}
			}
			if got, want := eBike3Count, int32(7); got != want {
				t.Fatalf("eBike3Count = %d; want: %d ", got, want)
			}
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
		// remove item
		remove := []command{
			{
				&protocol.Command{EntityId: entityID, Name: "RemoveLineItem"},
				&shoppingcart.RemoveLineItem{UserId: "user1", ProductId: "e-bike-1"},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "RemoveLineItem"},
				&shoppingcart.RemoveLineItem{UserId: "user1", ProductId: "e-bike-2"},
			},
			{
				&protocol.Command{EntityId: entityID, Name: "RemoveLineItem"},
				&shoppingcart.RemoveLineItem{UserId: "user1", ProductId: "e-bike-3"},
			},
		}
		for _, cmd := range remove {
			r := p.sendRecvCmd(cmd)
			switch m := r.Message.(type) {
			case *entity.EventSourcedStreamOut_Reply:
				p.checkCommandID(m)
				if m.Reply.ClientAction.GetFailure() != nil {
					t.Fatalf("unexpected ClientAction failure: %+v", m)
				}
			case *entity.EventSourcedStreamOut_Failure:
				p.checkCommandID(m)
				t.Fatalf("expected reply but got: %+v", m.Failure)
			default:
				t.Fatalf("unexpected message: %+v", m)
			}
		}
		// get the shopping cart
		r = p.sendRecvCmd(command{
			&protocol.Command{EntityId: entityID, Name: "GetShoppingCart"},
			&shoppingcart.GetShoppingCart{UserId: "user1"},
		})
		switch m := r.Message.(type) {
		case *entity.EventSourcedStreamOut_Reply:
			p.checkCommandID(m)
			payload := m.Reply.GetClientAction().GetReply().GetPayload()
			cart := &shoppingcart.Cart{}
			if err := encoding.UnmarshalAny(payload, cart); err != nil {
				t.Fatal(err)
			}
			if got, want := len(cart.Items), 0; got != want {
				t.Fatalf("len(cart.Items) = %d; want: %d ", got, want)
			}
		case *entity.EventSourcedStreamOut_Failure:
			t.Fatalf("expected reply but got: %+v", m.Failure)
		default:
			t.Fatalf("unexpected message: %+v", m)
		}
	})

	t.Run("send the User Function an error message", func(t *testing.T) {
		reportError, err := edc.ReportError(ctx, &protocol.UserFunctionError{Message: "an error occured"})
		if err != nil {
			t.Fatal(err)
		}
		if reportError == nil {
			t.Fatalf("reportError was nil")
		}
	})
}
