package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// CSVData 将数据结构转换为CSV格式，使用字段注释作为表头
func CSVData(data interface{}) string {
	v := reflect.ValueOf(data)

	// 处理切片类型
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return "无数据"
		}

		// 获取第一个元素来提取表头
		firstElem := v.Index(0)
		if firstElem.Kind() == reflect.Ptr {
			firstElem = firstElem.Elem()
		}

		// 生成表头
		headers := getHeaders(firstElem.Type())

		// 生成CSV内容
		var buf bytes.Buffer
		buf.WriteString(strings.Join(headers, ",") + "\n")

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			row := getRowValues(elem)
			buf.WriteString(strings.Join(row, ",") + "\n")
		}

		return buf.String()
	}

	// 处理单个结构体
	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// 生成表头
		headers := getHeaders(v.Type())

		// 生成CSV内容
		var buf bytes.Buffer
		buf.WriteString(strings.Join(headers, ",") + "\n")
		row := getRowValues(v)
		buf.WriteString(strings.Join(row, ",") + "\n")

		return buf.String()
	}

	jdata, _ := json.Marshal(data)
	// 其他类型直接使用JSON格式
	return string(jdata)
}

// getHeaders 从结构体类型中提取字段注释作为表头
func getHeaders(t reflect.Type) []string {
	typeName := t.Name()

	// 根据不同的结构体类型返回对应的中文表头
	switch typeName {
	case "Position":
		return []string{
			"交易对", "持仓数量", "开仓均价", "标记价格", "未实现盈亏", "强平价格",
			"杠杆倍数", "当前杠杆下用户可用的最大名义价值", "保证金模式", "逐仓保证金",
			"是否自动追加保证金", "持仓方向", "名义价值", "逐仓钱包余额", "更新时间",
		}
	case "FundingRateHistory":
		return []string{"交易对", "资金费率", "资金费率时间", "资金费对应标记价格"}
	case "BookTicker":
		return []string{"交易对", "最优买单价", "最优买单价挂单量", "最优卖单价", "最优卖单价挂单量", "撮合引擎时间"}
	case "Kline":
		return []string{
			"开盘时间", "收盘时间", "开盘价", "最高价", "最低价", "收盘价", "成交量",
			"成交额", "成交数量", "主动买入成交量", "主动买入成交额",
		}
	default:
		// 默认使用json标签作为表头
		var headers []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				headers = append(headers, parts[0])
			}
		}
		return headers
	}
}

// getRowValues 从结构体值中提取字段值作为CSV行
func getRowValues(v reflect.Value) []string {
	var values []string
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 获取字段的json标签
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			// 获取字段值
			var value string
			switch fieldValue.Kind() {
			case reflect.String:
				value = fieldValue.String()
			case reflect.Int, reflect.Int64:
				// 检查是否是时间戳字段，如果是则格式化为易读的时间格式
				fieldName := field.Name
				if isTimestampField(fieldName) || isTimeField(jsonTag) {
					timestamp := fieldValue.Int()
					if timestamp > 0 {
						// 格式化为 "MM-dd HH:mm:ss"
						value = time.Unix(0, timestamp*int64(time.Millisecond)).Format("01-02 15:04:05")
					} else {
						value = "0"
					}
				} else {
					value = fmt.Sprintf("%d", fieldValue.Int())
				}
			case reflect.Float64, reflect.Float32:
				value = fmt.Sprintf("%f", fieldValue.Float())
			case reflect.Bool:
				value = fmt.Sprintf("%t", fieldValue.Bool())
			default:
				// 对于其他类型，尝试转换为字符串
				value = fmt.Sprintf("%v", fieldValue.Interface())
			}

			// 处理可能包含逗号的值，用双引号包围
			if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") {
				value = strings.ReplaceAll(value, "\"", "\"\"")
				value = fmt.Sprintf("\"%s\"", value)
			}

			values = append(values, value)
		}
	}

	return values
}

// isTimestampField 判断字段名是否表示时间戳
func isTimestampField(fieldName string) bool {
	timestampFields := []string{
		"OpenTime", "CloseTime", "Time", "UpdateTime", "FundingTime",
		"NextFundingTime", "Timestamp", "updateTime",
	}
	for _, tf := range timestampFields {
		if fieldName == tf {
			return true
		}
	}
	return false
}

// isTimeField 根据json标签判断字段是否表示时间
func isTimeField(jsonTag string) bool {
	timeFields := []string{
		"openTime", "closeTime", "time", "updateTime", "fundingTime",
		"nextFundingTime", "timestamp",
	}
	parts := strings.Split(jsonTag, ",")
	fieldName := parts[0]
	for _, tf := range timeFields {
		if fieldName == tf {
			return true
		}
	}
	return false
}
