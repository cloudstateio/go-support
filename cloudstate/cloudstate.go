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
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"

	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	filedescr "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	SupportLibraryVersion = "0.1.0"
	SupportLibraryName    = "cloudstate-go-support"
)

// CloudState is an instance of a CloudState User Function
type CloudState struct {
	server                *grpc.Server
	entityDiscoveryServer *EntityDiscoveryServer
	eventSourcedServer    *EventSourcedServer
}

// New returns a new CloudState instance.
func New(options Options) (*CloudState, error) {
	eds, err := newEntityDiscoveryServer(options)
	if err != nil {
		return nil, err
	}
	cs := &CloudState{
		server:                grpc.NewServer(),
		entityDiscoveryServer: eds,
		eventSourcedServer:    newEventSourcedServer(),
	}
	protocol.RegisterEntityDiscoveryServer(cs.server, cs.entityDiscoveryServer)
	protocol.RegisterEventSourcedServer(cs.server, cs.eventSourcedServer)
	return cs, nil
}

// Options go get a CloudState instance configured.
type Options struct {
	ServiceName    string
	ServiceVersion string
}

// DescriptorConfig configures service and dependent descriptors.
type DescriptorConfig struct {
	Service        string
	Domain         []string
	DomainMessages []descriptor.Message
}

func (dc DescriptorConfig) AddDomainMessage(m descriptor.Message) DescriptorConfig {
	dc.DomainMessages = append(dc.DomainMessages, m)
	return dc
}

func (dc DescriptorConfig) AddDomainDescriptor(filename string) DescriptorConfig {
	dc.Domain = append(dc.Domain, filename)
	return dc
}

// RegisterEventSourcedEntity registers an event sourced entity for CloudState.
func (cs *CloudState) RegisterEventSourcedEntity(ese *EventSourcedEntity, config DescriptorConfig) (err error) {
	ese.registerOnce.Do(func() {
		if err = ese.init(); err != nil {
			return
		}
		if err = cs.eventSourcedServer.registerEntity(ese); err != nil {
			return
		}
		if err = cs.entityDiscoveryServer.registerEntity(ese, config); err != nil {
			return
		}
	})
	return
}

// Run runs the CloudState instance.
func (cs *CloudState) Run() error {
	host, ok := os.LookupEnv("HOST")
	if !ok {
		return fmt.Errorf("unable to get environment variable \"HOST\"")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		return fmt.Errorf("unable to get environment variable \"PORT\"")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	if e := cs.server.Serve(lis); e != nil {
		return fmt.Errorf("failed to grpcServer.Serve for: %v", lis)
	}
	return nil
}

// EntityDiscoveryServer implements the Cloudstate discovery protocol.
type EntityDiscoveryServer struct {
	fileDescriptorSet *filedescr.FileDescriptorSet
	entitySpec        *protocol.EntitySpec
	message           *descriptor.Message
}

// newEntityDiscoveryServer returns a new and initialized EntityDiscoveryServer.
func newEntityDiscoveryServer(options Options) (*EntityDiscoveryServer, error) {
	svr := &EntityDiscoveryServer{}
	svr.entitySpec = &protocol.EntitySpec{
		Entities: make([]*protocol.Entity, 0),
		ServiceInfo: &protocol.ServiceInfo{
			ServiceName:           options.ServiceName,
			ServiceVersion:        options.ServiceVersion,
			ServiceRuntime:        fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			SupportLibraryName:    SupportLibraryName,
			SupportLibraryVersion: SupportLibraryVersion,
		},
	}
	svr.fileDescriptorSet = &filedescr.FileDescriptorSet{
		File: make([]*filedescr.FileDescriptorProto, 0),
	}
	return svr, nil
}

// Discover returns an entity spec for
func (r *EntityDiscoveryServer) Discover(_ context.Context, pi *protocol.ProxyInfo) (*protocol.EntitySpec, error) {
	log.Printf("Received discovery call from sidecar [%s w%s] supporting Cloudstate %v.%v\n",
		pi.ProxyName,
		pi.ProxyVersion,
		pi.ProtocolMajorVersion,
		pi.ProtocolMinorVersion,
	)
	log.Printf("Responding with: %v\n", r.entitySpec.GetServiceInfo())
	return r.entitySpec, nil
}

// ReportError logs any user function error reported by the Cloudstate proxy.
func (r *EntityDiscoveryServer) ReportError(_ context.Context, fe *protocol.UserFunctionError) (*empty.Empty, error) {
	log.Printf("ReportError: %v\n", fe)
	return &empty.Empty{}, nil
}

func (r *EntityDiscoveryServer) updateSpec() (err error) {
	protoBytes, err := proto.Marshal(r.fileDescriptorSet)
	if err != nil {
		return errors.New("unable to Marshal FileDescriptorSet")
	}
	r.entitySpec.Proto = protoBytes
	return nil
}

func (r *EntityDiscoveryServer) resolveFileDescriptors(dc DescriptorConfig) error {
	// service
	if dc.Service != "" {
		if err := r.registerFileDescriptorProto(dc.Service); err != nil {
			return err
		}
	}
	// and dependent domain descriptors
	for _, dp := range dc.Domain {
		if err := r.registerFileDescriptorProto(dp); err != nil {
			return err
		}
	}
	for _, dm := range dc.DomainMessages {
		if err := r.registerFileDescriptor(dm); err != nil {
			return err
		}
	}
	return nil
}

func (r *EntityDiscoveryServer) registerEntity(e *EventSourcedEntity, config DescriptorConfig) error {
	if err := r.resolveFileDescriptors(config); err != nil {
		return fmt.Errorf("failed to resolveFileDescriptor for DescriptorConfig: %+v: %w", config, err)
	}
	r.entitySpec.Entities = append(r.entitySpec.Entities, &protocol.Entity{
		EntityType:    EventSourced,
		ServiceName:   e.ServiceName,
		PersistenceId: e.PersistenceID,
	})
	return r.updateSpec()
}

func (r *EntityDiscoveryServer) hasRegistered(filename string) bool {
	for _, f := range r.fileDescriptorSet.File {
		if f.GetName() == filename {
			return true
		}
	}
	return false
}

func (r *EntityDiscoveryServer) registerFileDescriptorProto(filename string) error {
	if r.hasRegistered(filename) {
		return nil
	}
	descriptorProto, err := unpackFile(proto.FileDescriptor(filename))
	if err != nil {
		return fmt.Errorf("failed to registerFileDescriptorProto for filename: %s: %w", filename, err)
	}
	r.fileDescriptorSet.File = append(r.fileDescriptorSet.File, descriptorProto)
	for _, fn := range descriptorProto.Dependency {
		err := r.registerFileDescriptorProto(fn)
		if err != nil {
			return err
		}
	}
	return r.updateSpec()
}

func (r *EntityDiscoveryServer) registerFileDescriptor(msg descriptor.Message) error {
	fd, _ := descriptor.ForMessage(msg) // this can panic
	if r := recover(); r != nil {
		return fmt.Errorf("descriptor.ForMessage panicked (%v) for: %+v", r, msg)
	}
	if r.hasRegistered(fd.GetName()) {
		return nil
	}
	r.fileDescriptorSet.File = append(r.fileDescriptorSet.File, fd)
	return nil
}
