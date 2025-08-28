package openstack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	fips "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
)

type ProvisionOpts struct {
	Name           string
	ImageID        string
	FlavorID       string
	NetworkID      string
	KeyName        string
	SecurityGroups []string
	UserData       string

	AssignFIP     bool
	FloatingIP    string // 지정 시 그 FIP를 연결
	ExternalNetID string // 비어있으면 "public" 등 기본값을 내부에서 선택해도 됨
}

type ProvisionResult struct {
	ServerID   string
	Status     string
	FixedIP    string
	FloatingIP string
}

func (c *Clients) ProvisionServer(ctx context.Context, o ProvisionOpts) (*ProvisionResult, error) {
	// 1) 서버 생성
	create := servers.CreateOpts{
		Name:           o.Name,
		ImageRef:       o.ImageID,
		FlavorRef:      o.FlavorID,
		Networks:       []servers.Network{{UUID: o.NetworkID}},
		SecurityGroups: o.SecurityGroups,
		// KeyName 필드는 v2에서 다른 이름일 수 있음 - 일단 제거
	}
	if o.UserData != "" {
		create.UserData = []byte(o.UserData)
	}

	// v2 API에 맞게 수정 - SchedulerHintOpts는 nil로 전달
	s, err := servers.Create(ctx, c.ComputeV2, create, nil).Extract()
	if err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}

	// 2) ACTIVE 대기 (간단 폴링)
	if err := waitServerStatus(ctx, c.ComputeV2, s.ID, "ACTIVE", 300*time.Second); err != nil {
		return nil, err
	}

	// 최신 상태 조회
	s, err = servers.Get(ctx, c.ComputeV2, s.ID).Extract()
	if err != nil {
		return nil, fmt.Errorf("get server: %w", err)
	}

	// 3) Fixed IP 추출 (첫 IPv4 기준)
	fixed := firstIPv4FromAddresses(s.Addresses)

	res := &ProvisionResult{
		ServerID: s.ID,
		Status:   s.Status,
		FixedIP:  fixed,
	}

	// 4) FIP 연결 (옵션)
	if o.AssignFIP {
		// 대상 포트(ID) 찾기: device_id = 서버ID
		pp, err := ports.List(c.NetworkV2, ports.ListOpts{DeviceID: s.ID}).AllPages(ctx)
		if err != nil {
			return nil, fmt.Errorf("list ports: %w", err)
		}
		pl, err := ports.ExtractPorts(pp)
		if err != nil || len(pl) == 0 {
			return nil, fmt.Errorf("no port found for server %s", s.ID)
		}
		// 네트워크 일치하는 포트 우선 선택
		var portID string
		for _, p := range pl {
			if p.NetworkID == o.NetworkID {
				portID = p.ID
				break
			}
		}
		if portID == "" {
			portID = pl[0].ID
		}

		var fip *fips.FloatingIP
		if o.FloatingIP != "" {
			// 기존 FIP 찾기 - 정확한 필드명 확인 필요
			pg, err := fips.List(c.NetworkV2, fips.ListOpts{}).AllPages(ctx)
			if err != nil {
				return nil, fmt.Errorf("list floating ips: %w", err)
			}
			list, _ := fips.ExtractFloatingIPs(pg)
			// 수동으로 IP 주소 검색
			for _, f := range list {
				if f.FloatingIP == o.FloatingIP {
					fip = &f
					break
				}
			}
			if fip == nil {
				return nil, fmt.Errorf("floating ip %s not found", o.FloatingIP)
			}
		} else {
			ext := o.ExternalNetID
			if ext == "" {
				// 기본 external 네트워크 ID를 내부에서 정하는 로직을 넣어도 됨
				// TODO: Clients 구조체에 DefaultExternalNetID 필드 추가 필요
				return nil, errors.New("externalNetworkId is required (no default set)")
			}
			fip, err = fips.Create(ctx, c.NetworkV2, fips.CreateOpts{
				FloatingNetworkID: ext,
			}).Extract()
			if err != nil {
				return nil, fmt.Errorf("create floating ip: %w", err)
			}
		}

		// 포트에 연결 (Associate) → Update로 PortID 설정
		_, err = fips.Update(ctx, c.NetworkV2, fip.ID, fips.UpdateOpts{
			PortID: &portID,
		}).Extract()
		if err != nil {
			return nil, fmt.Errorf("associate floating ip: %w", err)
		}
		res.FloatingIP = fip.FloatingIP
	}

	return res, nil
}

func waitServerStatus(ctx context.Context, cc *gophercloud.ServiceClient, id, want string, timeout time.Duration) error {
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("wait for status %s timeout", want)
		}
		s, err := servers.Get(ctx, cc, id).Extract()
		if err != nil {
			return fmt.Errorf("get server while waiting: %w", err)
		}
		if s.Status == want {
			return nil
		}
		if s.Status == "ERROR" {
			return fmt.Errorf("server entered ERROR state")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

// Addresses 맵에서 첫 IPv4 추출 (네트워크명 → []{Addr,...})
func firstIPv4FromAddresses(addr any) string {
	type aRec struct {
		Addr    string
		Version int
	}
	// gophercloud는 map[string][]aRec 형태로 언마샬됨
	m, ok := addr.(map[string]any)
	if !ok {
		return ""
	}
	for _, v := range m {
		if list, ok := v.([]any); ok {
			for _, it := range list {
				if ar, ok := it.(map[string]any); ok {
					if ver, _ := ar["version"].(float64); int(ver) == 4 {
						if ip, _ := ar["addr"].(string); ip != "" {
							return ip
						}
					}
				}
			}
		}
	}
	return ""
}
