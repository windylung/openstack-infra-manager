package openstack

import (
	"context"
	"fmt"

	cqs "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/quotasets"
	novaqs "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
)

type QuotaDetail struct {
	Limit int `json:"limit"`
	InUse int `json:"in_use"`
}

type NovaQuotaDetail struct {
	Cores     QuotaDetail `json:"cores"`
	RAMMB     QuotaDetail `json:"ramMB"` // MB
	Instances QuotaDetail `json:"instances"`
}

type CinderQuotaDetail struct {
	Gigabytes QuotaDetail `json:"gigabytes"` // GB
	Volumes   QuotaDetail `json:"volumes"`
	Snapshots QuotaDetail `json:"snapshots"`
}

func (c *Clients) GetNovaQuotaDetail(ctx context.Context, projectID string) (*NovaQuotaDetail, error) {
	q, err := novaqs.GetDetail(ctx, c.ComputeV2, projectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("nova get quota detail: %w", err)
	}
	return &NovaQuotaDetail{
		Cores:     QuotaDetail{Limit: q.Cores.Limit, InUse: q.Cores.InUse},
		RAMMB:     QuotaDetail{Limit: q.RAM.Limit, InUse: q.RAM.InUse},
		Instances: QuotaDetail{Limit: q.Instances.Limit, InUse: q.Instances.InUse},
	}, nil
}

// ----- Cinder 상세 조회 -----
func (c *Clients) GetCinderQuotaDetail(ctx context.Context, targetProjectID string) (*CinderQuotaDetail, error) {
	// v3: admin project 경로 사용
	q, err := cqs.GetUsage(ctx, c.BlockStorageV3, targetProjectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("cinder get quota detail: %w", err)
	}
	return &CinderQuotaDetail{
		Gigabytes: QuotaDetail{Limit: q.Gigabytes.Limit, InUse: q.Gigabytes.InUse},
		Volumes:   QuotaDetail{Limit: q.Volumes.Limit, InUse: q.Volumes.InUse},
		Snapshots: QuotaDetail{Limit: q.Snapshots.Limit, InUse: q.Snapshots.InUse},
	}, nil
}
