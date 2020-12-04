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

package main

import (
	"log"

	"github.com/cloudstateio/go-support/cloudstate"
	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/action"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/example/shoppingcart"
	actionTCK "github.com/cloudstateio/go-support/tck/action"
	"github.com/cloudstateio/go-support/tck/crdt2"
	tck "github.com/cloudstateio/go-support/tck/eventsourced"
)

// tag::shopping-cart-main[]
func main() {
	server, err := cloudstate.New(protocol.Config{
		ServiceName:    "cloudstate.tck.model.EventSourcedTckModel", // the servicename the proxy gets to know about
		ServiceVersion: "0.2.0",
	})
	if err != nil {
		log.Fatalf("cloudstate.New failed: %s", err)
	}

	err = server.RegisterEventSourced(
		&eventsourced.Entity{
			ServiceName:   "cloudstate.tck.model.EventSourcedTckModel",
			PersistenceID: "event-sourced-tck-model",
			SnapshotEvery: 5,
			EntityFunc:    tck.NewTestModel,
		}, protocol.DescriptorConfig{
			Service: "eventsourced.proto",
		},
	)
	if err != nil {
		log.Fatalf("Cloudstate failed to register entity: %s", err)
	}
	err = server.RegisterEventSourced(
		&eventsourced.Entity{
			ServiceName:   "cloudstate.tck.model.EventSourcedTwo",
			PersistenceID: "EventSourcedTwo",
			SnapshotEvery: 5,
			EntityFunc:    tck.NewTestModelTwo,
		}, protocol.DescriptorConfig{
			Service: "eventsourced.proto",
		},
	)
	if err != nil {
		log.Fatalf("Cloudstate failed to register entity: %s", err)
	}
	// tag::event-sourced-entity-type[]
	// tag::register[]
	err = server.RegisterEventSourced(&eventsourced.Entity{
		ServiceName:   "com.example.shoppingcart.ShoppingCart",
		PersistenceID: "ShoppingCart",
		EntityFunc:    shoppingcart.NewShoppingCart,
	}, protocol.DescriptorConfig{
		Service: "shoppingcart.proto",
	}.AddDomainDescriptor("domain.proto"))
	// end::register[]
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}

	err = server.RegisterAction(&action.Entity{
		ServiceName: "cloudstate.tck.model.action.ActionTckModel",
		EntityFunc:  actionTCK.NewTestModel,
	}, protocol.DescriptorConfig{
		Service: "tck_action.proto",
	})
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}
	err = server.RegisterAction(&action.Entity{
		ServiceName: "cloudstate.tck.model.action.ActionTwo",
		EntityFunc:  actionTCK.NewTestModelTwo,
	}, protocol.DescriptorConfig{
		Service: "tck_action.proto",
	})
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}

	// ServiceName: "cloudstate.tck.model.action.ActionTwo",

	// end::event-sourced-entity-type[]

	err = server.RegisterCRDT(&crdt.Entity{
		ServiceName: "cloudstate.tck.model.crdt.CrdtTckModel",
		EntityFunc:  crdt2.NewCrdtTckModelEntity,
	}, protocol.DescriptorConfig{
		Service: "tck_crdt2.proto",
	})
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}
	err = server.RegisterCRDT(&crdt.Entity{
		ServiceName: "cloudstate.tck.model.crdt.CrdtTwo",
		EntityFunc:  crdt2.NewCrdtTwoEntity,
	}, protocol.DescriptorConfig{
		Service: "tck_crdt2.proto",
	})
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}

	err = server.Run()
	if err != nil {
		log.Fatalf("Cloudstate failed to run: %v", err)
	}
}

// end::shopping-cart-main[]
