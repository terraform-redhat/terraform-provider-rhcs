package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// interface1: resource

type TFResource interface {
	CreateContext(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics)
	ReadContext(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics)
	UpdateContext(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics)
	DeleteContext(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics)
	GetImportContext(ctx context.Context) *schema.ResourceImporter
	Schema() map[string]*schema.Schema
	GetTimeout() *schema.ResourceTimeout
	FromResourceData(resourceData *schema.ResourceData, targetObject any)
	ToResourceData(resourceData *schema.ResourceData, targetObject any)
}

// interface2: nested resource

type TFNestedResource interface {
	//expand from resource data
	ExpandFromResourceData(resourceData *schema.ResourceData, targetObject any) error

	//expand from interface
	ExpandFromInterface(i interface{}, targetObject any) error

	//flat
	Flat(fromObject any, resourceData *schema.ResourceData) error
	// if err := resourceData.Set("metadata", k8s.FlattenMetadata(vm.ObjectMeta)); err != nil {
	//		return err
	//	}
}
