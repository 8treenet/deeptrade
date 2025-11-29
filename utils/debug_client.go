package utils

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func NewDebugHTTPClient(apikey ...string) *http.Client {
	key := ""
	if len(apikey) > 0 {
		key = apikey[0]
	}
	return &http.Client{ //nolint:gochecknoglobals
		Transport: &logTransport{
			Transport: http.DefaultTransport,
			key:       key,
		},
	}
}

type logTransport struct {
	key       string
	Transport http.RoundTripper
}

// RoundTrip logs the request and response with full contents using httputil.DumpRequest and httputil.DumpResponse.
func (t *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n------------------------------------------------------------\n")
	// zhaiyao := string(dump)
	// if len(zhaiyao) > 500 {
	// 	zhaiyao = zhaiyao[0:500]
	// }
	fmt.Printf("httpclient 请求, 长度:%v 摘要:%s \n\n", len(dump), dump)
	if t.key != "" {
		newReq := req.Clone(req.Context())
		vals := newReq.URL.Query()
		vals.Set("key", t.key)
		newReq.URL.RawQuery = vals.Encode()
		req = newReq
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("httpclient 结果: %s \n", string(dump))
	fmt.Printf("------------------------------------------------------------\n")
	return resp, nil
}
