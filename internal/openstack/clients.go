package openstack

import (
	"fmt"

	"example.com/quotaapi/internal/config"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

type Clients struct {
	Provider       *gophercloud.ProviderClient
	Identity       *gophercloud.ServiceClient
	ComputeV2      *gophercloud.ServiceClient
	BlockStorageV3 *gophercloud.ServiceClient
	AdminProjectID string
	Region         string
}

func NewServiceClients(cfg *config.Config) (*Clients, error) {
	provider, ident, err := NewIdentity(cfg)
	if err != nil {
		return nil, fmt.Errorf("keystone auth: %w", err)
	}

	compute, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: cfg.RegionName})
	if err != nil {
		return nil, fmt.Errorf("new compute v2: %w", err)
	}

	block, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region:       cfg.RegionName,
		Type:         "block-storage",                // 카탈로그에 있는 type 과 맞춤
		Name:         "cinder",                       // (선택) name 일치시켜 매칭 더 확실
		Availability: gophercloud.AvailabilityPublic, // public endpoint 사용
	})
	if err != nil {
		return nil, fmt.Errorf("new blockstorage v3: %w", err)
	}

	return &Clients{
		Provider:       provider,
		Identity:       ident,
		ComputeV2:      compute,
		BlockStorageV3: block,
		AdminProjectID: cfg.AdminProjectID,
		Region:         cfg.RegionName,
	}, nil
}
