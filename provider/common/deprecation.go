// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	ocmConsts "github.com/openshift-online/ocm-common/pkg/ocm/consts"
)

// AddDeprecationWarning inspects an OCM API response's headers for known CS
// deprecation headers and, if present, appends a diag.Diagnostics warning.
//
// This is deliberately narrow: it's wired only at the specific call sites CS
// is currently known to emit these headers from (ROSA Classic cluster
// creation and upgrade scheduling — see ROSAENG-62412). Generic coverage via
// an sdk.TransportWrapper is tracked as follow-up work.
func AddDeprecationWarning(diags *diag.Diagnostics, header http.Header) {
	if header == nil {
		return
	}
	deprecation := header.Get(ocmConsts.DeprecationHeader)
	message := header.Get(ocmConsts.OcmDeprecationMessage)
	fields := header.Get(ocmConsts.OcmFieldDeprecation)
	if deprecation == "" && message == "" && fields == "" {
		return
	}

	var detail strings.Builder
	switch {
	case deprecation != "" || message != "":
		detail.WriteString("This OCM API is deprecated. ")
		if deprecation != "" {
			fmt.Fprintf(&detail, "Removal date: %s. ", deprecation)
		}
		detail.WriteString(message)
	default:
		fmt.Fprintf(&detail, "This OCM API uses deprecated fields: %s", fields)
	}
	diags.AddWarning("OCM API deprecation notice", strings.TrimSpace(detail.String()))
}
