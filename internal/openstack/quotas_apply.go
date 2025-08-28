package openstack

import (
	"context"
	"fmt"

	cqs "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/quotasets"
	novaqs "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
	neutronqs "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/quotas"
)

func (c *Clients) ApplyNovaQuota(ctx context.Context, projectID string, cores, ramMB, instances *int) error {
	opts := novaqs.UpdateOpts{
		Cores:     cores,
		RAM:       ramMB,
		Instances: instances,
	}
	_, err := novaqs.Update(ctx, c.ComputeV2, projectID, opts).Extract()
	if err != nil {
		return fmt.Errorf("nova quota update: %w", err)
	}
	return nil
}

func (c *Clients) ApplyCinderQuota(ctx context.Context, projectID string, volumes, snapshots, gigabytes *int) error {
	opts := cqs.UpdateOpts{
		Volumes:   volumes,
		Snapshots: snapshots,
		Gigabytes: gigabytes,
	}

	_, err := cqs.Update(ctx, c.BlockStorageV3, projectID, opts).Extract()
	if err != nil {
		return fmt.Errorf("cinder quota update: %w", err)
	}
	return nil
}

// ApplyNeutronQuota applies Neutron quotas for ports and floating IPs
func (c *Clients) ApplyNeutronQuota(ctx context.Context, projectID string, ports, floatingIPs *int) error {
	opts := neutronqs.UpdateOpts{
		Port:       ports,
		FloatingIP: floatingIPs,
	}

	_, err := neutronqs.Update(ctx, c.NetworkV2, projectID, opts).Extract()
	if err != nil {
		return fmt.Errorf("neutron quota update: %w", err)
	}
	return nil
}
