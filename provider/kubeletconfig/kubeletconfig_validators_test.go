/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeletconfig

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KubeletConfig Validators", func() {

	Context("Pod Pids Limit Validation", func() {

		pidsLimitValidator := PidsLimitValidator{}
		var ctx context.Context
		var resp *validator.Int64Response

		BeforeEach(func() {
			ctx = context.TODO()
			resp = &validator.Int64Response{}
		})

		It("Fails validation if podPidsLimit is below allowed minimum", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(4000),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeTrue())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(1))
		})

		It("Fails validation is podPidsLimit is above unsafe maximum", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(4000000),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeTrue())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(1))
		})

		It("Adds a warning if podPidsLimit is above default maximum", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(25000),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.WarningsCount()).To(Equal(1))
		})

		It("Passes validation if podPidsLimit is within the default, safe range", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(10000),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(0))
		})

		It("Passes validation if podPidsLimit is the minimum", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(MinPodPidsLimit),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(0))
		})

		It("Passes validation if podPidsLimit is the maximum", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Value(MaxPodPidsLimit),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)

			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(0))
		})

		It("Passes validation if podPidsLimit is currently unknown", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Unknown(),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(0))
		})

		It("Passes validation if podPidsLimit is currently null", func() {
			req := validator.Int64Request{
				Path:        path.Root("pod_pids_limit"),
				ConfigValue: types.Int64Null(),
			}

			pidsLimitValidator.ValidateInt64(ctx, req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.Diagnostics.Errors().ErrorsCount()).To(Equal(0))
		})
	})
})
