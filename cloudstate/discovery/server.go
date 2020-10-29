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

// Package discovery implements the Cloudstate entity discovery server.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	filedescr "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/empty"
)

const (
	SupportLibraryName    = "cloudstate-go-support"
	SupportLibraryVersion = "0.2.0"
	ProtocolMajorVersion  = 0
	ProtocolMinorVersion  = 2
)

// EntityDiscoveryServer implements the Cloudstate discovery protocol.
type EntityDiscoveryServer struct {
	mu                sync.RWMutex
	fileDescriptorSet *filedescr.FileDescriptorSet
	entitySpec        *protocol.EntitySpec
}

// NewServer returns a new and initialized EntityDiscoveryServer.
func NewServer(config protocol.Config) *EntityDiscoveryServer {
	return &EntityDiscoveryServer{
		entitySpec: &protocol.EntitySpec{
			Entities: make([]*protocol.Entity, 0),
			ServiceInfo: &protocol.ServiceInfo{
				ServiceName:           config.ServiceName,
				ServiceVersion:        config.ServiceVersion,
				ServiceRuntime:        fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
				SupportLibraryName:    SupportLibraryName,
				SupportLibraryVersion: SupportLibraryVersion,
				ProtocolMajorVersion:  ProtocolMajorVersion,
				ProtocolMinorVersion:  ProtocolMinorVersion,
			},
		},
		fileDescriptorSet: &filedescr.FileDescriptorSet{
			File: make([]*filedescr.FileDescriptorProto, 0),
		},
	}
}

// Discover returns an entity spec for registered entities.
func (s *EntityDiscoveryServer) Discover(_ context.Context, info *protocol.ProxyInfo) (*protocol.EntitySpec, error) {
	log.Printf("Received discovery call from sidecar [%s %s] supporting Cloudstate %v.%v\n",
		info.ProxyName,
		info.ProxyVersion,
		info.ProtocolMajorVersion,
		info.ProtocolMinorVersion,
	)
	log.Printf("Responding with: %v\n", s.entitySpec.GetServiceInfo())
	// TODO: s.entitySpec can be written potentially but should not after we started to run the server;
	//  check how to enforce that after protocol.Run has started.
	return s.entitySpec, nil
}

// ReportError logs any user function error reported by the Cloudstate proxy.
func (s *EntityDiscoveryServer) ReportError(_ context.Context, error *protocol.UserFunctionError) (*empty.Empty, error) {
	log.Printf("ReportError: %v\n", error)
	return &empty.Empty{}, nil
}

func (s *EntityDiscoveryServer) updateSpec() (err error) {
	protoBytes, err := proto.Marshal(s.fileDescriptorSet)
	if err != nil {
		return errors.New("unable to Marshal FileDescriptorSet")
	}
	s.entitySpec.Proto = protoBytes
	return nil
}

func (s *EntityDiscoveryServer) resolveFileDescriptors(config protocol.DescriptorConfig) error {
	if config.Service != "" {
		if err := s.registerFileDescriptorProto(config.Service); err != nil {
			return err
		}
	}
	// Add dependent domain descriptors.
	for _, dp := range config.Domain {
		if err := s.registerFileDescriptorProto(dp); err != nil {
			return err
		}
	}
	for _, dm := range config.DomainMessages {
		if err := s.registerFileDescriptor(dm); err != nil {
			return err
		}
	}
	return nil
}

func (s *EntityDiscoveryServer) RegisterEventSourcedEntity(entity *eventsourced.Entity, config protocol.DescriptorConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.resolveFileDescriptors(config); err != nil {
		return fmt.Errorf("failed to resolve FileDescriptor for DescriptorConfig: %+v: %w", config, err)
	}
	s.entitySpec.Entities = append(s.entitySpec.Entities, &protocol.Entity{
		EntityType:    protocol.EventSourced,
		ServiceName:   entity.ServiceName.String(),
		PersistenceId: entity.PersistenceID,
	})
	return s.updateSpec()
}

func (s *EntityDiscoveryServer) RegisterCRDTEntity(entity *crdt.Entity, config protocol.DescriptorConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.resolveFileDescriptors(config); err != nil {
		return fmt.Errorf("failed to resolveFileDescriptor for DescriptorConfig: %+v: %w", config, err)
	}
	s.entitySpec.Entities = append(s.entitySpec.Entities, &protocol.Entity{
		EntityType:  protocol.CRDT,
		ServiceName: entity.ServiceName.String(),
	})
	return s.updateSpec()
}

func (s *EntityDiscoveryServer) hasRegistered(filename string) bool {
	for _, f := range s.fileDescriptorSet.File {
		if f.GetName() == filename {
			return true
		}
	}
	return false
}

func (s *EntityDiscoveryServer) registerFileDescriptorProto(filename string) error {
	if s.hasRegistered(filename) {
		return nil
	}
	desc, err := unpackFile(proto.FileDescriptor(filename))
	if err != nil {
		return fmt.Errorf("failed to registerFileDescriptorProto for filename: %s: %w", filename, err)
	}
	s.fileDescriptorSet.File = append(s.fileDescriptorSet.File, desc)
	for _, dep := range desc.Dependency {
		err := s.registerFileDescriptorProto(dep)
		if err != nil {
			return err
		}
	}
	return s.updateSpec()
}

func (s *EntityDiscoveryServer) registerFileDescriptor(msg descriptor.Message) error {
	fd, _ := descriptor.ForMessage(msg) // this can panic.
	if r := recover(); r != nil {
		return fmt.Errorf("descriptor.ForMessage panicked (%v) for: %+v", r, msg)
	}
	if s.hasRegistered(fd.GetName()) {
		return nil
	}
	s.fileDescriptorSet.File = append(s.fileDescriptorSet.File, fd)
	return nil
}
