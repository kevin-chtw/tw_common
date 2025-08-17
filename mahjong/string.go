package mahjong

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ToString 将各种类型转换为字符串
func ToString(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case []string:
		return strings.Join(v, ",")
	case []int:
		var strs []string
		for _, i := range v {
			strs = append(strs, strconv.Itoa(i))
		}
		return strings.Join(strs, ",")
	case []float64:
		var strs []string
		for _, f := range v {
			strs = append(strs, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return strings.Join(strs, ",")
	case map[string]string:
		var pairs []string
		for k, v := range v {
			pairs = append(pairs, k+":"+v)
		}
		return strings.Join(pairs, ",")
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		if b, err := json.Marshal(value); err == nil {
			return string(b)
		}
		return ""
	}
}

// ToLower 转换为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper 转换为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLowerInplace 原地转换为小写
func ToLowerInplace(s *string) {
	*s = strings.ToLower(*s)
}

// ToUpperInplace 原地转换为大写
func ToUpperInplace(s *string) {
	*s = strings.ToUpper(*s)
}

// ToQuoteString 添加引号
func ToQuoteString(s string) string {
	return "\"" + s + "\""
}

// ToNumber 字符串转数字
func ToNumber[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(s)
		return T(v), err
	case int8:
		v, err := strconv.ParseInt(s, 10, 8)
		return T(v), err
	case int16:
		v, err := strconv.ParseInt(s, 10, 16)
		return T(v), err
	case int32:
		v, err := strconv.ParseInt(s, 10, 32)
		return T(v), err
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return T(v), err
	case uint:
		v, err := strconv.ParseUint(s, 10, 0)
		return T(v), err
	case uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		return T(v), err
	case uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		return T(v), err
	case uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		return T(v), err
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return T(v), err
	case float32:
		v, err := strconv.ParseFloat(s, 32)
		return T(v), err
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return T(v), err
	default:
		return zero, nil
	}
}

// ToNumbers 字符串切片转数字切片
func ToNumbers[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ss []string) ([]T, error) {
	var result []T
	for _, s := range ss {
		n, err := ToNumber[T](s)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

// SplitByAny 按任意分隔符分割
func SplitByAny(s, delims string) []string {
	if s == "" {
		return nil
	}
	if delims == "" {
		return []string{s}
	}

	return strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune(delims, r)
	})
}

// SplitByString 按字符串分割
func SplitByString(s, sep string) []string {
	return strings.Split(s, sep)
}

// TrimString 去除前后空白字符
func TrimString(s string) string {
	return strings.TrimSpace(s)
}

// GetPureFileName 获取纯文件名
func GetPureFileName(path string) string {
	return filepath.Base(path)
}

// TimeStringToTime 时间字符串转Time
func TimeStringToTime(str string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", str)
	return t
}

// TimeToString Time转字符串
func TimeToString(t time.Time, digitOnly bool) string {
	if digitOnly {
		return t.Format("20060102150405")
	}
	return t.Format("2006-01-02 15:04:05")
}

// Formator 格式化输出
type Formator struct {
	msg   string
	datas []struct {
		name  string
		value string
	}
}

func NewFormator(msg string) *Formator {
	return &Formator{msg: msg}
}

func (f *Formator) Push(name string, value interface{}) *Formator {
	f.datas = append(f.datas, struct {
		name  string
		value string
	}{name: name, value: ToString(value)})
	return f
}

func (f *Formator) Clear() {
	f.msg = ""
	f.datas = nil
}

func (f *Formator) String() string {
	if len(f.datas) == 0 {
		return f.msg
	}

	var pairs []string
	for _, d := range f.datas {
		pairs = append(pairs, d.name+":"+d.value)
	}
	return f.msg + "[" + strings.Join(pairs, ",") + "]"
}

// Base64Encode Base64编码
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode Base64解码
func Base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PadLeft 左侧填充字符串
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(padChar), length-len(s)) + s
}

// PadRight 右侧填充字符串
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(padChar), length-len(s))
}

// Truncate 截断字符串
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}