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

package cloudstate

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cloudstateio/go-support/cloudstate/action"
	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/discovery"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/cloudstateio/go-support/cloudstate/value"
	"google.golang.org/grpc"
)

// CloudState is an instance of a Cloudstate User Function.
type CloudState struct {
	grpcServer            *grpc.Server
	entityDiscoveryServer *discovery.EntityDiscoveryServer
	eventSourcedServer    *eventsourced.Server
	crdtServer            *crdt.Server
	actionServer          *action.Server
	valueServer           *value.Server
}

// New returns a new CloudState instance.
func New(c protocol.Config) (*CloudState, error) {
	cs := &CloudState{
		grpcServer:            grpc.NewServer(),
		entityDiscoveryServer: discovery.NewServer(c),
		eventSourcedServer:    eventsourced.NewServer(),
		crdtServer:            crdt.NewServer(),
		actionServer:          action.NewServer(),
		valueServer:           value.NewServer(),
	}
	protocol.RegisterEntityDiscoveryServer(cs.grpcServer, cs.entityDiscoveryServer)
	entity.RegisterEventSourcedServer(cs.grpcServer, cs.eventSourcedServer)
	entity.RegisterCrdtServer(cs.grpcServer, cs.crdtServer)
	entity.RegisterValueEntityServer(cs.grpcServer, cs.valueServer)
	entity.RegisterActionProtocolServer(cs.grpcServer, cs.actionServer)
	return cs, nil
}

// RegisterEventSourced registers an event sourced entity.
func (cs *CloudState) RegisterEventSourced(entity *eventsourced.Entity, config protocol.DescriptorConfig, options ...eventsourced.Option) error {
	entity.Options(options...)
	if err := cs.eventSourcedServer.Register(entity); err != nil {
		return err
	}
	if err := cs.entityDiscoveryServer.RegisterEventSourcedEntity(entity, config); err != nil {
		return err
	}
	return nil
}

// RegisterCRDT registers a CRDT entity.
func (cs *CloudState) RegisterCRDT(entity *crdt.Entity, config protocol.DescriptorConfig, options ...crdt.Option) error {
	entity.Options(options...)
	if err := cs.crdtServer.Register(entity); err != nil {
		return err
	}
	if err := cs.entityDiscoveryServer.RegisterCRDTEntity(entity, config); err != nil {
		return err
	}
	return nil
}

// RegisterAction registers an action entity.
func (cs *CloudState) RegisterAction(entity *action.Entity, config protocol.DescriptorConfig) error {
	if err := cs.actionServer.Register(entity); err != nil {
		return err
	}
	if err := cs.entityDiscoveryServer.RegisterActionEntity(entity, config); err != nil {
		return err
	}
	return nil
}

// RegisterValueEntity registers a Value entity.
func (cs *CloudState) RegisterValueEntity(entity *value.Entity, config protocol.DescriptorConfig, options ...value.Option) error {
	entity.Options(options...)
	if err := cs.valueServer.Register(entity); err != nil {
		return err
	}
	if err := cs.entityDiscoveryServer.RegisterValueEntity(entity, config); err != nil {
		return err
	}
	return nil
}

// Run runs the CloudState instance on the interface and port defined by
// the HOST and PORT environment variable.
func (cs *CloudState) Run() error {
	host, ok := os.LookupEnv("HOST")
	if !ok {
		return errors.New("unable to get environment variable \"HOST\"")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		return errors.New("unable to get environment variable \"PORT\"")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	if err := cs.RunWithListener(lis); err != nil {
		return fmt.Errorf("failed to RunWithListener for: %v with: %w", lis, err)
	}
	return nil
}

// Run runs the CloudState instance with a listener provided.
func (cs *CloudState) RunWithListener(lis net.Listener) error {
	return cs.grpcServer.Serve(lis)
}

// Stop gracefully stops the Cloudstate instance.
func (cs *CloudState) Stop() {
	cs.grpcServer.GracefulStop()
	log.Println("CloudState stopped")
}
