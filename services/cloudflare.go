package services

import (
	"cf-oidc-proxy/config"
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"net"
	"os"
	"strings"
	"time"
)

type CloudflareClient struct {
	api *cloudflare.API
	ctx context.Context
}

func NewCloudflareClient(config config.OidcProxy) (cloudflareClient *CloudflareClient, err error) {
	apiToken := os.Getenv("CF_GLOBAL_API_TOKEN")
	if config.Cloudflare != nil && config.Cloudflare.ApiToken != nil {
		apiToken = *config.Cloudflare.ApiToken
	}

	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("error creating cloudflare client: %s", err)
	}
	ctx := context.Background()
	return &CloudflareClient{api: api, ctx: ctx}, nil
}

// CreateShortLivedToken Request short-lived api tokens that are limited to scopes, ressources and ttl
// Optional client ip whitelisting can be enabled
func (cf *CloudflareClient) CreateShortLivedToken(role *config.Role, groupName string, oidcActorRequestIPs []string) (apiToken *cloudflare.APIToken, err error) {
	ttl := calculateTtlTime(role.TTL)
	condition := createConditions(role, oidcActorRequestIPs)
	request := cloudflare.APIToken{
		Name:      groupName,
		Value:     "empty",
		ExpiresOn: &ttl,
		Condition: condition,
		Policies:  createPolicies(role),
	}
	token, err := cf.api.CreateAPIToken(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("error creating token: %s", err)
	}
	return &token, nil
}

// Create Policies based on group config
func createPolicies(group *config.Role) (policies []cloudflare.APITokenPolicies) {
	for _, policy := range group.Policies {
		policies = append(policies, cloudflare.APITokenPolicies{
			Effect:           strings.ToLower(policy.Effect),
			Resources:        createResources(policy.Resources),
			PermissionGroups: createPermissionGroups(policy.Permissions),
		})
	}
	return policies
}

// Create Resources based on group config
func createResources(resources []config.Resource) (resourcesList map[string]interface{}) {
	resourcesList = make(map[string]interface{})
	for _, resource := range resources {
		resourcesList[resource.Name] = resource.Value
	}
	return resourcesList
}

// Create PermissionsGroup based on group config
func createPermissionGroups(permissions []config.Permission) (permissionGroups []cloudflare.APITokenPermissionGroups) {
	for _, permission := range permissions {
		permissionGroups = append(permissionGroups, cloudflare.APITokenPermissionGroups{
			ID: permission.Id,
		})
	}
	return permissionGroups
}

// Calculate the expired time (input ms)
// Cloudflare not supporting seconds so the token duration is minimal one minute.
func calculateTtlTime(ttl int) (expirationDate time.Time) {
	_ttl := time.Duration(ttl) * time.Second
	expirationDate = time.Now().UTC().Truncate(time.Second)
	if ttl > 0 {
		expirationDate = expirationDate.Add(_ttl).Add(time.Minute * 30)
	}
	return expirationDate
}

// Create Conditions for whitelist and blacklist client ips
// Add the oidc request ip to the whitelist if AllowOidcActor is enabled in the group config
func createConditions(group *config.Role, oidcActorRequestIPs []string) (conditions *cloudflare.APITokenCondition) {

	whitelist := group.Conditions.RequestIP.Whitelist
	blacklist := group.Conditions.RequestIP.Blacklist

	if group.Conditions.RequestIP.AllowOidcActor {
		for _, oidcActorRequestIp := range oidcActorRequestIPs {
			cidr, err := cidrFromIP(oidcActorRequestIp)
			if err == nil {
				whitelist = append(whitelist, cidr)
			}
		}
	}

	requestIp := &cloudflare.APITokenRequestIPCondition{
		In:    whitelist,
		NotIn: blacklist,
	}

	if len(whitelist) == 0 && len(blacklist) == 0 {
		return nil
	}

	return &cloudflare.APITokenCondition{
		RequestIP: requestIp,
	}
}

func cidrFromIP(ip string) (string, error) {
	var cidr string
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	if ipAddr.To4() != nil {
		cidr = ip + "/32"
	} else {
		cidr = ip + "/128"
	}
	return cidr, nil
}
