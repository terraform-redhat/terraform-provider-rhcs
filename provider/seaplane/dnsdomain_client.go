// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"encoding/json"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	seaplaneclientset "github.com/openshift-online/rosa-seaplane-api/clientset"
)

const dnsDomainClusterArchHCP = "hcp"

// DNSDomainSeaplane represents an HCP DNS-domain reservation returned by the
// Seaplane Platform API.
type DNSDomainSeaplane struct {
	ID          string
	ClusterArch string
}

type dnsDomainResponse struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	ClusterArch string `json:"cluster_arch"`
}

type dnsDomainListResponse struct {
	Items []dnsDomainResponse `json:"items"`
}

// DNSDomainSeaplaneClient manages account-scoped HCP DNS-domain reservations
// through the Seaplane Platform API.
type DNSDomainSeaplaneClient struct {
	client rest.Interface
}

// NewDNSDomainSeaplaneClient creates a DNS-domain client from a Seaplane
// clientset.
func NewDNSDomainSeaplaneClient(clientset seaplaneclientset.Interface) *DNSDomainSeaplaneClient {
	return &DNSDomainSeaplaneClient{client: clientset.SeaplaneV1alpha1().RESTClient()}
}

// Create reserves a DNS domain for the authenticated AWS account.
func (client *DNSDomainSeaplaneClient) Create(ctx context.Context, clusterArch string) (*DNSDomainSeaplane, error) {
	if clusterArch == "" {
		clusterArch = dnsDomainClusterArchHCP
	}
	if clusterArch != dnsDomainClusterArchHCP {
		return nil, fmt.Errorf("only HCP DNS domains are supported by the Seaplane Platform API")
	}

	result := client.client.Post().
		Resource("dns_domains").
		SetHeader("Content-Type", "application/json").
		Body([]byte(`{"cluster_arch":"hcp"}`)).
		Do(ctx)
	if err := result.Error(); err != nil {
		return nil, fmt.Errorf("create DNS domain: %w", err)
	}

	var domain dnsDomainResponse
	if err := decodeDNSDomainResponse(result, &domain); err != nil {
		return nil, err
	}
	if domain.ID == "" {
		return nil, fmt.Errorf("DNS domain response has no ID")
	}
	if domain.ClusterArch == "" {
		domain.ClusterArch = clusterArch
	}
	return &DNSDomainSeaplane{ID: domain.ID, ClusterArch: domain.ClusterArch}, nil
}

// List returns all HCP DNS-domain reservations owned by the authenticated AWS
// account.
func (client *DNSDomainSeaplaneClient) List(ctx context.Context) ([]DNSDomainSeaplane, error) {
	result := client.client.Get().Resource("dns_domains").Do(ctx)
	if err := result.Error(); err != nil {
		return nil, fmt.Errorf("list DNS domains: %w", err)
	}

	var response dnsDomainListResponse
	if err := decodeDNSDomainResponse(result, &response); err != nil {
		return nil, err
	}

	domains := make([]DNSDomainSeaplane, 0, len(response.Items))
	for _, item := range response.Items {
		if item.ID == "" {
			continue
		}
		if item.ClusterArch == "" {
			item.ClusterArch = dnsDomainClusterArchHCP
		}
		domains = append(domains, DNSDomainSeaplane{ID: item.ID, ClusterArch: item.ClusterArch})
	}
	return domains, nil
}

// Get returns the DNS domain with the specified ID. The Platform API exposes
// collection reads only, so Get matches the reservation from the account list.
func (client *DNSDomainSeaplaneClient) Get(ctx context.Context, id string) (*DNSDomainSeaplane, error) {
	domains, err := client.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range domains {
		if domains[i].ID == id {
			return &domains[i], nil
		}
	}
	return nil, nil
}

// Delete removes a DNS-domain reservation by its ID.
func (client *DNSDomainSeaplaneClient) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("DNS domain ID is required")
	}
	if err := client.client.Delete().Resource("dns_domains").Name(id).Do(ctx).Error(); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete DNS domain %q: %w", id, err)
	}
	return nil
}

func decodeDNSDomainResponse(result rest.Result, target any) error {
	data, err := result.Raw()
	if err != nil {
		return fmt.Errorf("read DNS domain response: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode DNS domain response: %w", err)
	}
	return nil
}
