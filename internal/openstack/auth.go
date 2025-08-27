package openstack

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"

	"example.com/quotaapi/internal/config"
)

// Keystone 로그인: Provider(토큰 보관) + Identity v3 클라이언트 생성
func NewIdentity(c *config.Config) (*gophercloud.ProviderClient, *gophercloud.ServiceClient, error) {
	ao := gophercloud.AuthOptions{
		IdentityEndpoint: c.AuthURL,
		Username:         c.Username,
		Password:         c.Password,
		DomainID:         c.UserDomainID,
		TenantName:       c.ProjectName, // 프로젝트 스코프
		AllowReauth:      true,
	}
	provider, err := openstack.AuthenticatedClient(context.Background(), ao)
	if err != nil {
		return nil, nil, err
	}
	ident, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{
		Region: c.RegionName,
	})
	if err != nil {
		return nil, nil, err
	}
	return provider, ident, nil
}

// 토큰 조회(X-Subject-Token). subjectToken이 비었으면 서비스 토큰 자체 조회.
func IntrospectToken(ctx context.Context, ident *gophercloud.ServiceClient, provider *gophercloud.ProviderClient, subjectToken string) (map[string]any, error) {

	if subjectToken == "" {
		subjectToken = provider.TokenID
	}
	var raw map[string]any
	_, err := ident.Get(ctx,
		ident.ServiceURL("auth", "tokens"),
		&raw,
		&gophercloud.RequestOpts{
			MoreHeaders: map[string]string{
				"X-Subject-Token": subjectToken,
			},
		},
	)
	return raw, err
}
