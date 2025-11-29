package utils

import (
	"net/http"
	"net/url"
	"time"
)

// GetProxyHTTPClient 获取HTTP客户端
func GetProxyHTTPClient(proxyURL string, timeout int) *http.Client {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	if proxyURL != "" {
		proxyobj, err := url.Parse(proxyURL)
		if err != nil {
			return client
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyobj),
		}
	}
	return client
}
