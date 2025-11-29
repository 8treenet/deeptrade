package utils

import (
	"fmt"
	"strconv"
)

// ParseFloatSafe 安全地解析浮点数，提供默认值
func ParseFloatSafe(s string, defaultValue float64) float64 {
	if value, err := strconv.ParseFloat(s, 64); err == nil {
		return value
	}
	return defaultValue
}

// FormatPrice 格式化价格显示
func FormatPrice(price float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, price)
}

// FormatPriceWithUnit 格式化价格并添加单位
func FormatPriceWithUnit(price float64, precision int, unit string) string {
	return fmt.Sprintf("%s %s", FormatPrice(price, precision), unit)
}

// CompareFloat 比较两个浮点数，考虑容差
func CompareFloat(a, b, tolerance float64) int {
	if a > b+tolerance {
		return 1 // a > b
	} else if a < b-tolerance {
		return -1 // a < b
	}
	return 0 // a == b (在容差范围内)
}

// IsPriceAbove 检查价格是否在参考值之上（考虑容差）
func IsPriceAbove(price, reference, tolerance float64) bool {
	return CompareFloat(price, reference, tolerance) > 0
}

// IsPriceBelow 检查价格是否在参考值之下（考虑容差）
func IsPriceBelow(price, reference, tolerance float64) bool {
	return CompareFloat(price, reference, tolerance) < 0
}

// IsPriceNear 检查价格是否接近参考值（在容差范围内）
func IsPriceNear(price, reference, tolerance float64) bool {
	return CompareFloat(price, reference, tolerance) == 0
}

// CalculatePercentageChange 计算百分比变化
func CalculatePercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return ((newValue - oldValue) / oldValue) * 100
}

// SafeDivide 安全除法，避免除零错误
func SafeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// Average 计算平均值
func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Min 返回最小值
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max 返回最大值
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Clamp 将值限制在指定范围内
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
