package resource

***REMOVED***
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type DNSDomain struct {
	client  *cmv1.DNSDomainsClient
	builder *cmv1.DNSDomainBuilder
	id      string
}

func NewDNSDomain(client *cmv1.DNSDomainsClient***REMOVED*** *DNSDomain {
	return &DNSDomain{
		client:  client,
		builder: cmv1.NewDNSDomain(***REMOVED***,
	}
}

func (d *DNSDomain***REMOVED*** GetDNSDomainBuilder(***REMOVED*** *cmv1.DNSDomainBuilder {
	return d.builder
}

func (d *DNSDomain***REMOVED*** Create(***REMOVED*** (*cmv1.DNSDomainsAddResponse, error***REMOVED*** {
	payload, err := d.builder.Build(***REMOVED***
	if err != nil {
		return nil, err
	}
	return d.client.Add(***REMOVED***.Body(payload***REMOVED***.Send(***REMOVED***
}

func (d *DNSDomain***REMOVED*** Delete(id string***REMOVED*** error {
	_, err := d.client.DNSDomain(id***REMOVED***.Delete(***REMOVED***.Send(***REMOVED***
	return err
}

func (d *DNSDomain***REMOVED*** Get(id string***REMOVED*** (*cmv1.DNSDomainGetResponse, error***REMOVED*** {
	return d.client.DNSDomain(id***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}
