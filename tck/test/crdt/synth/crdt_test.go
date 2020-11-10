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

package synth

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/example/crdt_shoppingcart/shoppingcart"
)

// TestCRDT runs the TCK for the CRDT state model.
// As defined by the Cloudstate specification, each CRDT state model type
// has three state actions CRDTs can emit on state changes.
// - create
// - update
// - delete

func Command(i interface{}) string {
	// github.com/cloudstateio/go-support/example/crdt_shoppingcart/shoppingcart.ShoppingCartServiceServer.AddItem
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func TestCRDT(t *testing.T) {
	s := newServer(t)
	s.newClientConn()
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("huh", func(t *testing.T) {
		fmt.Println(Command(shoppingcart.ShoppingCartServiceServer.AddItem))
	})

	t.Run("entity discovery should find the service", func(t *testing.T) {
		edc := protocol.NewEntityDiscoveryClient(s.conn)
		discover, err := edc.Discover(ctx, &protocol.ProxyInfo{
			ProtocolMajorVersion: 0,
			ProtocolMinorVersion: 1,
			ProxyName:            "a-proxy",
			ProxyVersion:         "0.0.0",
			SupportedEntityTypes: []string{protocol.EventSourced, protocol.CRDT},
		})
		if err != nil {
			t.Fatal(err)
		}
		tr := tester{t}
		tr.expectedInt(len(discover.GetEntities()), 1)
		tr.expectedString(discover.GetEntities()[0].GetServiceName(), serviceName)
	})
}
