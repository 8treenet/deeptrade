package binance

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误代码类型
type ErrorCode int

const (
	// 通用错误代码
	ErrCodeUnknown ErrorCode = iota
	ErrCodeInvalidRequest
	ErrCodeInvalidJSON
	ErrCodeInvalidSymbol
	ErrCodeInvalidInterval
	ErrCodeInvalidOrderType
	ErrCodeInvalidTimeInForce
	ErrCodeInvalidSide
	ErrCodeInvalidQuantity
	ErrCodeInvalidPrice
	ErrCodeInvalidTimestamp
	ErrCodeDisconnected
	ErrCodeUnauthorized
	ErrCodeTooManyRequests
	ErrCodeInternalError
	ErrCodeServiceUnavailable
	ErrCodeUnknownOrder
	ErrCodeOrderRejected
	ErrCodeCancelRejected
	ErrCodeNoSuchOrder
	ErrCodeInsufficientFunds
	ErrCodeAccountInactive
	ErrCodeDuplicateOrder
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode `json:"code"`    // 错误代码
	Message string    `json:"message"` // 错误消息
	Details string    `json:"details"` // 错误详情
	Raw     string    `json:"raw"`     // 原始错误信息
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("binance error [%d]: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("binance error [%d]: %s", e.Code, e.Message)
}

// NewError 创建新的错误
func NewError(code ErrorCode, message, details, raw string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
		Raw:     raw,
	}
}

// APIError API响应错误
type APIError struct {
	Code    int    `json:"code"`    // API错误代码
	Message string `json:"msg"`     // API错误消息
	Raw     string `json:"-"`       // 原始响应
}

// Error 实现error接口
func (e *APIError) Error() string {
	return fmt.Sprintf("binance api error [%d]: %s", e.Code, e.Message)
}

// NewAPIError 创建API错误
func NewAPIError(code int, message, raw string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Raw:     raw,
	}
}

// ConvertAPIError 转换API错误为自定义错误
func ConvertAPIError(apiErr *APIError) *Error {
	switch apiErr.Code {
	case -1021:
		return NewError(ErrCodeInvalidTimestamp, "无效的时间戳", "时间戳与服务器时间相差过大", apiErr.Raw)
	case -1100:
		return NewError(ErrCodeInvalidJSON, "非法字符", "请求中包含非法字符", apiErr.Raw)
	case -1101:
		return NewError(ErrCodeInvalidRequest, "字符过多", "请求参数字符过多", apiErr.Raw)
	case -1102:
		return NewError(ErrCodeInvalidRequest, "必选参数缺失", "请求缺少必选参数", apiErr.Raw)
	case -1103:
		return NewError(ErrCodeInvalidRequest, "未知参数", "请求包含未知参数", apiErr.Raw)
	case -1104:
		return NewError(ErrCodeInvalidRequest, "参数过多", "请求参数过多", apiErr.Raw)
	case -1105:
		return NewError(ErrCodeInvalidRequest, "参数值异常", "参数值不在预期范围内", apiErr.Raw)
	case -1112:
		return NewError(ErrCodeInvalidRequest, "参数为空", "参数不能为空", apiErr.Raw)
	case -1114:
		return NewError(ErrCodeInvalidJSON, "无效的JSON", "JSON格式错误", apiErr.Raw)
	case -1121:
		return NewError(ErrCodeInvalidSymbol, "无效的交易对", "交易对不存在", apiErr.Raw)
	case -1120:
		return NewError(ErrCodeInvalidInterval, "无效的时间间隔", "K线时间间隔无效", apiErr.Raw)
	case -1116:
		return NewError(ErrCodeInvalidOrderType, "无效的订单类型", "订单类型无效", apiErr.Raw)
	case -1115:
		return NewError(ErrCodeInvalidSide, "无效的买卖方向", "订单买卖方向无效", apiErr.Raw)
	case -1111:
		return NewError(ErrCodeInvalidQuantity, "无效的数量", "订单数量无效", apiErr.Raw)
	case -1110:
		return NewError(ErrCodeInvalidPrice, "无效的价格", "订单价格无效", apiErr.Raw)
	case -1000:
		return NewError(ErrCodeUnknown, "未知错误", "发生未知错误", apiErr.Raw)
	case -1001:
		return NewError(ErrCodeDisconnected, "断开连接", "内部错误;无法处理请求", apiErr.Raw)
	case -1002:
		return NewError(ErrCodeUnauthorized, "未授权", "无权访问该命令", apiErr.Raw)
	case -1003:
		return NewError(ErrCodeTooManyRequests, "请求过多", "请求次数过多，超过速率限制", apiErr.Raw)
	case -1006:
		return NewError(ErrCodeServiceUnavailable, "服务不可用", "服务当前不可用", apiErr.Raw)
	case -1007:
		return NewError(ErrCodeInternalError, "超时", "等待超时", apiErr.Raw)
	case -1013:
		return NewError(ErrCodeInvalidQuantity, "无效的数量", "数量无效", apiErr.Raw)
	case -1015:
		return NewError(ErrCodeUnauthorized, "无效的API密钥", "API密钥无效", apiErr.Raw)
	case -1022:
		return NewError(ErrCodeInvalidRequest, "签名无效", "该请求的签名无效", apiErr.Raw)
	case -1106:
		return NewError(ErrCodeInvalidRequest, "参数重复", "参数已指定", apiErr.Raw)
	case -2010:
		return NewError(ErrCodeUnknownOrder, "订单不存在", "订单不存在", apiErr.Raw)
	case -2011:
		return NewError(ErrCodeUnknownOrder, "订单已取消", "订单已被取消", apiErr.Raw)
	case -2013:
		return NewError(ErrCodeNoSuchOrder, "无此订单", "订单不存在", apiErr.Raw)
	case -2014:
		return NewError(ErrCodeInvalidRequest, "错误的API密钥权限", "API密钥权限错误", apiErr.Raw)
	case -2015:
		return NewError(ErrCodeInvalidRequest, "无效的API密钥格式", "API密钥格式错误", apiErr.Raw)
	case -2016:
		return NewError(ErrCodeAccountInactive, "账户不活跃", "账户未激活", apiErr.Raw)
	case -2019:
		return NewError(ErrCodeInvalidRequest, "不允许交易", "不允许交易", apiErr.Raw)
	case -2021:
		return NewError(ErrCodeOrderRejected, "订单被拒绝", "订单被拒绝", apiErr.Raw)
	case -2022:
		return NewError(ErrCodeCancelRejected, "取消订单被拒绝", "取消订单被拒绝", apiErr.Raw)
	case -2012:
		return NewError(ErrCodeInsufficientFunds, "余额不足", "账户余额不足", apiErr.Raw)
	case -2018:
		return NewError(ErrCodeCancelRejected, "取消订单失败", "余额不足，无法取消订单", apiErr.Raw)
	default:
		return NewError(ErrCodeUnknown, "未知API错误", apiErr.Message, apiErr.Raw)
	}
}

// ConvertHTTPError 转换HTTP状态码为自定义错误
func ConvertHTTPError(statusCode int, response string) *Error {
	switch statusCode {
	case http.StatusBadRequest:
		return NewError(ErrCodeInvalidRequest, "请求无效", "HTTP 400 Bad Request", response)
	case http.StatusUnauthorized:
		return NewError(ErrCodeUnauthorized, "未授权", "HTTP 401 Unauthorized", response)
	case http.StatusForbidden:
		return NewError(ErrCodeUnauthorized, "禁止访问", "HTTP 403 Forbidden", response)
	case http.StatusTooManyRequests:
		return NewError(ErrCodeTooManyRequests, "请求过多", "HTTP 429 Too Many Requests", response)
	case http.StatusInternalServerError:
		return NewError(ErrCodeInternalError, "内部错误", "HTTP 500 Internal Server Error", response)
	case http.StatusBadGateway:
		return NewError(ErrCodeServiceUnavailable, "网关错误", "HTTP 502 Bad Gateway", response)
	case http.StatusServiceUnavailable:
		return NewError(ErrCodeServiceUnavailable, "服务不可用", "HTTP 503 Service Unavailable", response)
	case http.StatusGatewayTimeout:
		return NewError(ErrCodeServiceUnavailable, "网关超时", "HTTP 504 Gateway Timeout", response)
	default:
		return NewError(ErrCodeUnknown, "未知HTTP错误", fmt.Sprintf("HTTP %d", statusCode), response)
	}
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if binanceErr, ok := err.(*Error); ok {
		return binanceErr.Code == ErrCodeServiceUnavailable || 
			   binanceErr.Code == ErrCodeDisconnected ||
			   binanceErr.Code == ErrCodeInternalError
	}
	return false
}

// IsRateLimitError 检查是否为限流错误
func IsRateLimitError(err error) bool {
	if binanceErr, ok := err.(*Error); ok {
		return binanceErr.Code == ErrCodeTooManyRequests
	}
	return false
}

// IsAuthError 检查是否为认证错误
func IsAuthError(err error) bool {
	if binanceErr, ok := err.(*Error); ok {
		return binanceErr.Code == ErrCodeUnauthorized ||
			   binanceErr.Code == ErrCodeInvalidTimestamp ||
			   binanceErr.Code == ErrCodeAccountInactive
	}
	return false
}

// IsOrderError 检查是否为订单相关错误
func IsOrderError(err error) bool {
	if binanceErr, ok := err.(*Error); ok {
		return binanceErr.Code == ErrCodeUnknownOrder ||
			   binanceErr.Code == ErrCodeNoSuchOrder ||
			   binanceErr.Code == ErrCodeOrderRejected ||
			   binanceErr.Code == ErrCodeCancelRejected ||
			   binanceErr.Code == ErrCodeInsufficientFunds ||
			   binanceErr.Code == ErrCodeInvalidQuantity ||
			   binanceErr.Code == ErrCodeInvalidPrice
	}
	return false
}