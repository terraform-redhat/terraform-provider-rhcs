// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	ocmConsts "github.com/openshift-online/ocm-common/pkg/ocm/consts"
)

// dateDisplayFormat renders a deprecation removal date in a human-friendly
// way for end users, e.g. "January 1, 2027".
const dateDisplayFormat = "January 2, 2006"

// AddDeprecationWarning inspects an OCM API response's headers for known CS
// deprecation headers and, if present, appends a user-focused
// diag.Diagnostics warning describing the migration action needed.
//
// This is deliberately narrow: it's wired only at the specific call sites CS
// is currently known to emit these headers from (ROSA Classic cluster
// creation and upgrade scheduling — see ROSAENG-62412). Generic coverage via
// an sdk.TransportWrapper is tracked as follow-up work.
func AddDeprecationWarning(diags *diag.Diagnostics, header http.Header) {
	if header == nil {
		return
	}
	deprecationDate := header.Get(ocmConsts.DeprecationHeader)
	deprecationMessage := header.Get(ocmConsts.OcmDeprecationMessage)
	fieldDeprecations := header.Get(ocmConsts.OcmFieldDeprecation)
	if deprecationDate == "" && deprecationMessage == "" && fieldDeprecations == "" {
		return
	}

	messages := collectDeprecationMessages(deprecationMessage, fieldDeprecations)

	var detail strings.Builder
	detail.WriteString("Deprecation warning: ")
	if len(messages) > 0 {
		detail.WriteString(strings.Join(messages, " "))
	} else {
		detail.WriteString("This OCM API is deprecated.")
	}
	if deprecationDate != "" {
		fmt.Fprintf(&detail, " Deprecation date: %s.", formatDeprecationDate(deprecationDate))
	}

	diags.AddWarning("OCM API deprecation notice", detail.String())
}

// collectDeprecationMessages gathers the user-facing migration messages from
// the endpoint-level deprecation message and any per-field deprecation
// messages, deduplicating identical text (CS commonly repeats the same
// migration guidance in both places) and ensuring each ends with punctuation.
func collectDeprecationMessages(deprecationMessage, fieldDeprecations string) []string {
	seen := map[string]bool{}
	var messages []string
	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			return
		}
		seen[m] = true
		if !strings.HasSuffix(m, ".") && !strings.HasSuffix(m, "!") && !strings.HasSuffix(m, "?") {
			m += "."
		}
		messages = append(messages, m)
	}

	add(deprecationMessage)

	if fieldDeprecations != "" {
		var fields map[string]string
		if err := json.Unmarshal([]byte(fieldDeprecations), &fields); err == nil {
			for _, fieldMessage := range fields {
				add(fieldMessage)
			}
		}
	}

	return messages
}

// formatDeprecationDate renders a deprecation date header (RFC3339 or
// RFC1123Z, per CS's usage) in a human-friendly form, falling back to the
// raw value if it can't be parsed.
func formatDeprecationDate(raw string) string {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.Format(dateDisplayFormat)
	}
	if t, err := time.Parse(time.RFC1123Z, raw); err == nil {
		return t.Format(dateDisplayFormat)
	}
	return raw
}
