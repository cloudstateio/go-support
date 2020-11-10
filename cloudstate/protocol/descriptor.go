package protocol

import "github.com/golang/protobuf/descriptor"

// Config go get a CloudState instance configured.
type Config struct {
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

func (dc DescriptorConfig) AddDomainDescriptor(filename ...string) DescriptorConfig {
	dc.Domain = append(dc.Domain, filename...)
	return dc
}
