package testauth

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

const (
	tenantIDEnv               = "K8S_AZURE_TENANTID"
	subscriptionEnv           = "K8S_AZURE_SUBSID"
	servicePrincipleIDEnv     = "K8S_AZURE_SPID"
	servicePrincipleSecretEnv = "K8S_AZURE_SPSEC"
	clusterLocationEnv        = "K8S_AZURE_LOCATION"
)

// AzureAuthConfig holds auth related part of cloud config
// Only consider servicePrinciple now
type AzureAuthConfig struct {
	// The AAD Tenant ID for the Subscription that the cluster is deployed in
	TenantID string
	// The ClientID for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientID string
	// The ClientSecret for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientSecret string
	// The ID of the Azure Subscription that the cluster is deployed in
	SubscriptionID string
}

// GetServicePrincipalToken creates a new service principal token based on the configuration
func GetServicePrincipalToken(config *AzureAuthConfig, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		return nil, fmt.Errorf("creating the OAuth config: %v", err)
	}

	if len(config.AADClientSecret) > 0 {
		glog.V(2).Infoln("azure: using client_id+client_secret to retrieve access token")
		return adal.NewServicePrincipalToken(
			*oauthConfig,
			config.AADClientID,
			config.AADClientSecret,
			env.ServiceManagementEndpoint)
	}

	return nil, fmt.Errorf("No credentials provided for AAD application %s", config.AADClientID)
}

// AzureAuthConfigFromTestProfile obtains azure config from Environment
func AzureAuthConfigFromTestProfile() *AzureAuthConfig {
	c := &AzureAuthConfig{
		TenantID:        os.Getenv(tenantIDEnv),
		AADClientID:     os.Getenv(servicePrincipleIDEnv),
		AADClientSecret: os.Getenv(servicePrincipleSecretEnv),
		SubscriptionID:  os.Getenv(subscriptionEnv),
	}
	return c
}
