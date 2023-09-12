package common

import "net/http"

type HttpClient interface {
	Get(url string***REMOVED*** (resp *http.Response, err error***REMOVED***
}

type DefaultHttpClient struct {
}

func (c DefaultHttpClient***REMOVED*** Get(url string***REMOVED*** (resp *http.Response, err error***REMOVED*** {
	return http.Get(url***REMOVED***
}
