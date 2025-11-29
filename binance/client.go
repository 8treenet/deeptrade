package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"deeptrade/conf"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"time"
)

var fclient *FuturesClient
var once sync.Once

// GetOnceFuturesClient .
func GetOnceFuturesClient() *FuturesClient {
	// 获取期货客户端
	once.Do(func() {
		client, err := GetFuturesClient()
		if err != nil {
			log.Printf("[市场数据] 创建期货客户端失败: %v", err)
			panic(err)
		}
		fclient = client
	})
	return fclient
}

// RequestRateLimit 请求速率限制器
type RequestRateLimit struct {
	rateLimit int
	interval  int
	tokens    int
	lastTime  time.Time
}

// NewRequestRateLimit 创建新的速率限制器
func NewRequestRateLimit(rateLimit, interval int) *RequestRateLimit {
	return &RequestRateLimit{
		rateLimit: rateLimit,
		interval:  interval,
		tokens:    rateLimit,
		lastTime:  time.Now(),
	}
}

// Wait 等待直到可以发送请求
func (r *RequestRateLimit) Wait() {
	now := time.Now()
	elapsed := now.Sub(r.lastTime).Milliseconds()

	// 补充令牌
	if elapsed >= int64(r.interval) {
		r.tokens = r.rateLimit
		r.lastTime = now
	} else {
		tokensToAdd := int(elapsed * int64(r.rateLimit) / int64(r.interval))
		if tokensToAdd > 0 {
			r.tokens += tokensToAdd
			if r.tokens > r.rateLimit {
				r.tokens = r.rateLimit
			}
			r.lastTime = now
		}
	}

	// 如果没有令牌，等待
	if r.tokens <= 0 {
		waitTime := time.Duration(r.interval) * time.Millisecond
		time.Sleep(waitTime)
		r.tokens = r.rateLimit
		r.lastTime = time.Now()
	} else {
		r.tokens--
	}
}

// NewFuturesClientFromConfig 使用项目配置创建期货客户端
func NewFuturesClientFromConfig() (*FuturesClient, error) {
	envConfig := conf.Get().GetBinanceEnvironment()
	config := &ClientConfig{
		APIKey:             envConfig.APIKey,
		SecretKey:          envConfig.SecretKey,
		BaseURL:            envConfig.FuturesBaseURL,
		StreamURL:          envConfig.FuturesStreamURL,
		Timeout:            conf.Get().Binance.Timeout,
		Debug:              envConfig.Debug,
		MaxRetries:         conf.Get().Binance.MaxRetries,
		RateLimitRateLimit: 1200,
		RateLimitInterval:  60000,
		RecvWindow:         60000, // 增加时间戳窗口到60秒
		RetryDelay:         1000,
		RetryBackoff:       2,
	}

	// 设置代理
	proxyURL := conf.Get().Binance.DefaultProxy
	if proxyURL != "" {
		config.ProxyURL = proxyURL
	}

	return NewFuturesClientFromClientConfig(config)
}

// NewFuturesClientFromClientConfig 从ClientConfig创建期货客户端
func NewFuturesClientFromClientConfig(config *ClientConfig) (*FuturesClient, error) {
	if err := validateClientConfig(config); err != nil {
		return nil, err
	}

	client := &FuturesClient{
		clientConfig: config,
		httpClient:   getHTTPClient(config),
		rateLimit:    NewRequestRateLimit(config.RateLimitRateLimit, config.RateLimitInterval),
	}

	return client, nil
}

// validateClientConfig 验证客户端配置
func validateClientConfig(config *ClientConfig) error {
	if config.APIKey == "" {
		return NewError(ErrCodeInvalidRequest, "配置无效", "API密钥不能为空", "")
	}
	if config.SecretKey == "" {
		return NewError(ErrCodeInvalidRequest, "配置无效", "密钥不能为空", "")
	}
	if config.Timeout <= 0 {
		return NewError(ErrCodeInvalidRequest, "配置无效", "超时时间必须大于0", "")
	}
	if config.RecvWindow < 0 || config.RecvWindow > 60000 {
		return NewError(ErrCodeInvalidRequest, "配置无效", "接收窗口必须在0-60000毫秒之间", "")
	}
	if config.MaxRetries < 0 {
		return NewError(ErrCodeInvalidRequest, "配置无效", "最大重试次数不能为负数", "")
	}
	return nil
}

// getHTTPClient 获取HTTP客户端
func getHTTPClient(config *ClientConfig) *http.Client {
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// 设置代理
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			// 如果代理URL解析失败，记录警告但不影响客户端创建
			if config.Debug {
				fmt.Printf("[DEBUG] 代理URL解析失败: %v\n", err)
			}
			return client
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		if config.Debug {
			fmt.Printf("[DEBUG] 使用代理: %s\n", proxyURL.String())
		}
	}

	return client
}

// sign 签名请求
func (c *ClientConfig) sign(params string) string {
	if c.SecretKey == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(params))
	return hex.EncodeToString(h.Sum(nil))
}

// addAuthParams 添加认证参数
func (c *ClientConfig) addAuthParams(params map[string]string) {
	// 检查是否已经添加过认证参数（避免重试时重复添加）
	if params["_auth_added"] == "true" {
		return
	}

	// 创建参数副本以避免修改原始映射
	authParams := make(map[string]string)
	for k, v := range params {
		authParams[k] = v
	}

	if c.APIKey != "" {
		authParams["apiKey"] = c.APIKey
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	authParams["timestamp"] = strconv.FormatInt(timestamp, 10)

	if c.RecvWindow > 0 {
		authParams["recvWindow"] = strconv.Itoa(c.RecvWindow)
	}

	// 生成签名 - 使用原始参数（不进行预编码）
	if c.SecretKey != "" {
		queryString := c.buildQueryString(authParams)
		signature := c.sign(queryString)
		authParams["signature"] = signature
	}

	// 标记已添加认证参数
	authParams["_auth_added"] = "true"

	// 将认证参数复制回原始映射
	for k, v := range authParams {
		params[k] = v
	}
}

// buildQueryString 构建查询字符串
func (c *ClientConfig) buildQueryString(params map[string]string) string {
	var keys []string
	for k := range params {
		// 跳过内部参数（以下划线开头的参数）
		if !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}

	// 按字母顺序排序 - 使用sort包确保一致性
	sort.Strings(keys)

	var parts []string
	for _, key := range keys {
		value := params[key]
		if value != "" {
			// 对值进行URL编码以确保与币安API兼容
			encodedValue := url.QueryEscape(value)
			parts = append(parts, fmt.Sprintf("%s=%s", key, encodedValue))
		}
	}

	return strings.Join(parts, "&")
}
