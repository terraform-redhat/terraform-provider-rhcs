// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import "net/http"

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type DefaultHttpClient struct {
}

func (c DefaultHttpClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}
