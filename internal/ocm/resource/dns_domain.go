package resource

import (
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type DNSDomain struct {
	client  *cmv1.DNSDomainsClient
	builder *cmv1.DNSDomainBuilder
	id      string
}

func NewDNSDomain(client *cmv1.DNSDomainsClient) *DNSDomain {
	return &DNSDomain{
		client:  client,
		builder: cmv1.NewDNSDomain(),
	}
}

func (d *DNSDomain) GetDNSDomainBuilder() *cmv1.DNSDomainBuilder {
	return d.builder
}

func (d *DNSDomain) Create() (*cmv1.DNSDomainsAddResponse, error) {
	payload, err := d.builder.Build()
	if err != nil {
		return nil, err
	}
	return d.client.Add().Body(payload).Send()
}

func (d *DNSDomain) Delete(id string) error {
	_, err := d.client.DNSDomain(id).Delete().Send()
	return err
}

func (d *DNSDomain) Get(id string) (*cmv1.DNSDomainGetResponse, error) {
	return d.client.DNSDomain(id).Get().Send()
}
