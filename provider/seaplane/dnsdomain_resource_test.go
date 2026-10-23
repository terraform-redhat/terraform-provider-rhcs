// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"io"
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

const testDNSDomain = "abcd.0.example.com"

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
	for _, attribute := range []string{"id", "cluster_arch"} {
		if _, ok := schema.Schema.Attributes[attribute]; !ok {
			t.Errorf("schema is missing %q attribute", attribute)
		}
	}
}

func TestDNSDomainSeaplaneClientRejectsUnsupportedArchitecture(t *testing.T) {
	client := &DNSDomainSeaplaneClient{}
	if _, err := client.Create(context.Background(), "classic"); err == nil {
		t.Fatal("Create succeeded for unsupported architecture")
	}
}

func TestDNSDomainSeaplaneClientLifecycle(t *testing.T) {
	var requests []string
	deleteCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/stage/api/v0/dns_domains":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read create body: %v", err)
			}
			if got, want := string(body), `{"cluster_arch":"hcp"}`; got != want {
				t.Fatalf("create body = %q, want %q", got, want)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"kind":"DNSDomain","id":"` + testDNSDomain + `","cluster_arch":"hcp"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/stage/api/v0/dns_domains":
			_, _ = w.Write([]byte(`{"kind":"DNSDomainList","items":[{"id":"` + testDNSDomain + `","cluster_arch":"hcp"}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/stage/api/v0/dns_domains/"+testDNSDomain:
			deleteCount++
			if deleteCount == 1 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	clientset, err := seaplaneclientset.NewForConfig(&seaplanerest.Config{
		Host:      server.URL + "/stage",
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
	client := NewDNSDomainSeaplaneClient(clientset)

	created, err := client.Create(context.Background(), "hcp")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != testDNSDomain || created.ClusterArch != "hcp" {
		t.Fatalf("Create returned %+v", created)
	}

	got, err := client.Get(context.Background(), testDNSDomain)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil || got.ID != testDNSDomain || got.ClusterArch != "hcp" {
		t.Fatalf("Get returned %+v", got)
	}

	missing, err := client.Get(context.Background(), "missing.example.com")
	if err != nil {
		t.Fatalf("Get missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("Get missing returned %+v", missing)
	}

	if err := client.Delete(context.Background(), testDNSDomain); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := client.Delete(context.Background(), testDNSDomain); err != nil {
		t.Fatalf("Delete already-missing domain: %v", err)
	}

	wantRequests := []string{
		http.MethodPost + " /stage/api/v0/dns_domains",
		http.MethodGet + " /stage/api/v0/dns_domains",
		http.MethodGet + " /stage/api/v0/dns_domains",
		http.MethodDelete + " /stage/api/v0/dns_domains/" + testDNSDomain,
		http.MethodDelete + " /stage/api/v0/dns_domains/" + testDNSDomain,
	}
	if strings.Join(requests, "\n") != strings.Join(wantRequests, "\n") {
		t.Fatalf("requests = %v, want %v", requests, wantRequests)
	}
}
