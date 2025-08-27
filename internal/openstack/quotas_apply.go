package openstack

import (
	"context"
	"fmt"

	cqs "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/quotasets"
	novaqs "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
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
