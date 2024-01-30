package hcp

import "github.com/hashicorp/terraform-plugin-framework/types"

type DefaultIngress struct {
	Id              types.String `tfsdk:"id"`
	Cluster         types.String `tfsdk:"cluster"`
	ListeningMethod types.String `tfsdk:"listening_method"`
}
