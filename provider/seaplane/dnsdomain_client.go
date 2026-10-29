// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplaneclientset "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
)

const (
	dnsReservationPollInterval = 5 * time.Second
	dnsReservationReadyTimeout = 10 * time.Minute
)

// DNSDomainSeaplane represents an account-scoped DNS reservation returned by
// the Seaplane Platform API. Name is the customer-selected metadata.name,
// while ID is the platform-assigned metadata.uid used by Cluster resources.
type DNSDomainSeaplane struct {
	Name       string
	ID         string
	BaseDomain string
	Phase      string
}

// DNSDomainSeaplaneClient manages account-scoped DNS reservations through the
// Seaplane Platform API.
type DNSDomainSeaplaneClient struct {
	reservations spplatform.DNSReservationInterface
}

// NewDNSDomainSeaplaneClient creates a DNS reservation client from a Seaplane
// clientset.
func NewDNSDomainSeaplaneClient(clientset seaplaneclientset.Interface) *DNSDomainSeaplaneClient {
	return &DNSDomainSeaplaneClient{
		reservations: clientset.SeaplaneV1alpha1().DNSReservations(),
	}
}

// Create reserves a DNS domain for the authenticated AWS account and waits
// until the platform assigns the reservation UID and base domain.
func (client *DNSDomainSeaplaneClient) Create(ctx context.Context, name string) (*DNSDomainSeaplane, error) {
	if name == "" {
		return nil, fmt.Errorf("DNS reservation name is required")
	}

	reservation := &v1alpha1.DNSReservation{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	if _, err := client.reservations.Create(ctx, reservation, spplatform.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("create DNS reservation %q: %w", name, err)
	}

	if err := client.waitReady(ctx, name); err != nil {
		return nil, err
	}
	result, err := client.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("DNS reservation %q disappeared after creation", name)
	}
	return result, nil
}

// Get returns the DNS reservation with the specified metadata.name. A missing
// reservation is reported as (nil, nil) so Terraform can remove it from state.
func (client *DNSDomainSeaplaneClient) Get(ctx context.Context, name string) (*DNSDomainSeaplane, error) {
	if name == "" {
		return nil, fmt.Errorf("DNS reservation name is required")
	}
	reservation, err := client.reservations.Get(ctx, name, spplatform.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read DNS reservation %q: %w", name, err)
	}
	return dnsDomainSeaplaneFromReservation(reservation), nil
}

// Delete removes a DNS reservation by metadata.name. The Platform API rejects
// deletion while a cluster still claims the reservation with HTTP 409.
func (client *DNSDomainSeaplaneClient) Delete(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("DNS reservation name is required")
	}
	if err := client.reservations.Delete(ctx, name, spplatform.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		if apierrors.IsConflict(err) {
			return fmt.Errorf(
				"DNS reservation %q is still claimed by a cluster; "+
					"destroy the cluster that uses it before deleting the reservation: %w",
				name, err,
			)
		}
		return fmt.Errorf("delete DNS reservation %q: %w", name, err)
	}
	return nil
}

func (client *DNSDomainSeaplaneClient) waitReady(ctx context.Context, name string) error {
	err := client.reservations.WaitUntil(
		ctx,
		name,
		func(reservation *v1alpha1.DNSReservation) bool {
			return reservation != nil && reservation.Status.Phase == v1alpha1.DNSReservationPhaseReady
		},
		dnsReservationPollInterval,
		dnsReservationReadyTimeout,
	)
	if err == nil {
		return nil
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("wait for DNS reservation %q to become Ready: %w", name, err)
	}

	last, getErr := client.Get(ctx, name)
	if getErr != nil || last == nil {
		return fmt.Errorf(
			"DNS reservation %q did not become Ready within %s: %w",
			name, dnsReservationReadyTimeout, err,
		)
	}
	return fmt.Errorf(
		"DNS reservation %q did not become Ready within %s (last phase %q)",
		name, dnsReservationReadyTimeout, last.Phase,
	)
}

func dnsDomainSeaplaneFromReservation(reservation *v1alpha1.DNSReservation) *DNSDomainSeaplane {
	if reservation == nil {
		return nil
	}
	return &DNSDomainSeaplane{
		Name:       reservation.Name,
		ID:         string(reservation.UID),
		BaseDomain: reservation.Status.BaseDomain,
		Phase:      string(reservation.Status.Phase),
	}
}
