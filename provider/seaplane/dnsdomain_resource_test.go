// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	seaplaneclientset "github.com/openshift-online/rosa-seaplane-api/clientset"
	seaplanerest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"
)

const (
	testDNSReservationName   = "reservation-a"
	testDNSReservationUID    = "c8e1b8db-87f0-4a83-922d-ff5ab5d81f9f"
	testDNSReservationDomain = "f7a3.0.openshiftapps.com"
)

func TestDNSDomainSeaplaneResourceMetadata(t *testing.T) {
	resourceUnderTest := NewDNSDomain()

	var metadata resource.MetadataResponse
	resourceUnderTest.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "rhcs"}, &metadata)
	if got, want := metadata.TypeName, "rhcs_dns_domain_seaplane"; got != want {
		t.Fatalf("TypeName = %q, want %q", got, want)
	}

	var schema resource.SchemaResponse
	resourceUnderTest.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	if schema.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics: %v", schema.Diagnostics)
	}
	for _, attribute := range []string{"name", "id", "base_domain", "phase"} {
		if _, ok := schema.Schema.Attributes[attribute]; !ok {
			t.Errorf("schema is missing %q attribute", attribute)
		}
	}
	if _, ok := schema.Schema.Attributes["cluster_arch"]; ok {
		t.Error("schema still exposes obsolete cluster_arch attribute")
	}
}

func TestDNSDomainSeaplaneClientLifecycle(t *testing.T) {
	var requests []string
	getCount := 0
	deleteCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/stage/api/v0/dns_reservations":
			var body struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.Metadata.Name != testDNSReservationName {
				t.Fatalf("create metadata.name = %q, want %q", body.Metadata.Name, testDNSReservationName)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(dnsReservationResponse("Pending", "")))
		case r.Method == http.MethodGet && r.URL.Path == "/stage/api/v0/dns_reservations/"+testDNSReservationName:
			getCount++
			_, _ = w.Write([]byte(dnsReservationResponse("Ready", testDNSReservationDomain)))
		case r.Method == http.MethodGet && r.URL.Path == "/stage/api/v0/dns_reservations/missing":
			http.NotFound(w, r)
		case r.Method == http.MethodDelete && r.URL.Path == "/stage/api/v0/dns_reservations/"+testDNSReservationName:
			deleteCount++
			if deleteCount == 1 {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestDNSDomainClient(t, server.URL)

	created, err := client.Create(context.Background(), testDNSReservationName)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	assertDNSReservation(t, created)
	if getCount != 2 {
		t.Fatalf("Create GET count = %d, want 2 (readiness poll and final read)", getCount)
	}

	got, err := client.Get(context.Background(), testDNSReservationName)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	assertDNSReservation(t, got)

	missing, err := client.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Get missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("Get missing returned %+v", missing)
	}

	if err := client.Delete(context.Background(), testDNSReservationName); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := client.Delete(context.Background(), testDNSReservationName); err != nil {
		t.Fatalf("Delete already-missing reservation: %v", err)
	}

	wantRequests := []string{
		http.MethodPost + " /stage/api/v0/dns_reservations",
		http.MethodGet + " /stage/api/v0/dns_reservations/" + testDNSReservationName,
		http.MethodGet + " /stage/api/v0/dns_reservations/" + testDNSReservationName,
		http.MethodGet + " /stage/api/v0/dns_reservations/" + testDNSReservationName,
		http.MethodGet + " /stage/api/v0/dns_reservations/missing",
		http.MethodDelete + " /stage/api/v0/dns_reservations/" + testDNSReservationName,
		http.MethodDelete + " /stage/api/v0/dns_reservations/" + testDNSReservationName,
	}
	if strings.Join(requests, "\n") != strings.Join(wantRequests, "\n") {
		t.Fatalf("requests = %v, want %v", requests, wantRequests)
	}
}

func TestDNSDomainSeaplaneClientDeleteReportsClaimedReservation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete && r.URL.Path == "/stage/api/v0/dns_reservations/"+testDNSReservationName {
			http.Error(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Conflict","code":409}`, http.StatusConflict)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := newTestDNSDomainClient(t, server.URL)
	err := client.Delete(context.Background(), testDNSReservationName)
	if err == nil {
		t.Fatal("Delete succeeded for a reservation claimed by a cluster")
	}
	if !strings.Contains(err.Error(), "still claimed by a cluster") {
		t.Fatalf("Delete error = %q, want claim guidance", err)
	}
}

func newTestDNSDomainClient(t *testing.T, host string) *DNSDomainSeaplaneClient {
	t.Helper()
	clientset, err := seaplaneclientset.NewForConfig(&seaplanerest.Config{
		Host:      host + "/stage",
		Region:    "us-east-1",
		AccountID: "123456789012",
		AWSConfig: aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
		},
	})
	if err != nil {
		t.Fatalf("NewForConfig: %v", err)
	}
	return NewDNSDomainSeaplaneClient(clientset)
}

func dnsReservationResponse(phase, baseDomain string) string {
	status := `{"phase":"` + phase + `"}`
	if baseDomain != "" {
		status = `{"phase":"` + phase + `","baseDomain":"` + baseDomain + `"}`
	}
	return `{
		"kind":"DNSReservation",
		"apiVersion":"seaplane.io/v1alpha1",
		"metadata":{"name":"` + testDNSReservationName + `","uid":"` + testDNSReservationUID + `"},
		"status":` + status + `
	}`
}

func assertDNSReservation(t *testing.T, reservation *DNSDomainSeaplane) {
	t.Helper()
	if reservation == nil {
		t.Fatal("DNS reservation is nil")
	}
	if reservation.Name != testDNSReservationName ||
		reservation.ID != testDNSReservationUID ||
		reservation.BaseDomain != testDNSReservationDomain ||
		reservation.Phase != "Ready" {
		t.Fatalf("DNS reservation = %+v", reservation)
	}
}
