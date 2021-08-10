package zjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	Log "github.com/wellmoon/go/logger"
)

type JSONObject struct {
	ItemMap map[string]interface{}
}

func NewObject() *JSONObject {
	itemMap := make(map[string]interface{})
	return &JSONObject{itemMap}
}

func (jsonObject *JSONObject) Put(key string, val interface{}) {
	jsonObject.ItemMap[key] = &val
}

func (jsonObject *JSONObject) Contains(key string) bool {
	var _, ok = jsonObject.ItemMap[key]
	return ok
}

func (jsonObject *JSONObject) GetInt(key string) int {
	value := jsonObject.ItemMap[key]
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			Log.Debug("GetInt by strconv.Atoi error: {}", err)
		}
		return r
	case int:
		return value
	case int64:
		return int(value)
	case int32:
		return int(value)
	default:
		r, err := strconv.Atoi(ToStr(value))
		if err != nil {
			Log.Debug("GetInt error for type {}", reflect.TypeOf(value))
		}
		return r
	}
}

func (jsonObject *JSONObject) GetInt64(key string) int64 {
	value := jsonObject.ItemMap[key]
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			Log.Debug("GetInt64 by strconv.Atoi error: {}", err)
		}
		return int64(r)
	case int:
		return int64(value)
	case int64:
		return value
	case int32:
		return int64(value)
	default:
		r, err := strconv.Atoi(ToStr(value))
		if err != nil {
			Log.Debug("GetInt64 error for type {}", reflect.TypeOf(value))
		}
		return int64(r)
	}
}

func (jsonObject *JSONObject) GetString(key string) string {
	value := jsonObject.ItemMap[key]
	switch value := value.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case int64:
		return ToStr(value)
	case float64:
		return decimal.NewFromFloat(value).String()
	default:
		Log.Debug("GetString for type {}", reflect.TypeOf(value))
		return ToStr(value)
	}
}

func (jsonObject *JSONObject) GetFloat(key string) float64 {
	value := jsonObject.ItemMap[key]
	v := value
	val, ok := v.(float64)
	if !ok {
		Log.Fatal("convert to float error, value is {}", v)
	}
	return val
}

func (jsonObject *JSONObject) GetBool(key string) bool {
	value := jsonObject.ItemMap[key]
	v := value
	val, ok := v.(bool)
	if ok {
		return val
	}

	str := ToStr(v)
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	case "0", "f", "F", "false", "FALSE", "False":
		return false
	}
	return false
}

func (jsonObject *JSONObject) Get(key string) interface{} {
	return jsonObject.ItemMap[key]
}

func (jsonObject *JSONObject) GetArray(key string) ([]interface{}, error) {
	value := jsonObject.ItemMap[key]
	v := value
	val, ok := v.([]interface{})
	if !ok {
		// 尝试把字符串转为slice
		str, sok := v.(string)
		if !sok {
			Log.Error("convert to array error, value is {}", v)
			return nil, errors.New("convert to array error")
		}
		slice := make([]interface{}, 0)
		err := ParseArray(str, &slice)
		if err != nil {
			return nil, err
		}
		return slice, nil
	}
	return val, nil
}

func (jsonObject *JSONObject) GetJSONObject(key string) (*JSONObject, error) {
	value := jsonObject.ItemMap[key]
	val, ok := value.(string)
	if !ok {
		Log.Fatal("convert to JSONObject error, value is {}", value)
	}
	result, err := ParseJSONObject(val)
	return result, err
}

func (jsonObject *JSONObject) Size() int {
	return len(jsonObject.ItemMap)
}

func ParseJSONObject(inter interface{}) (*JSONObject, error) {
	typeStr := reflect.TypeOf(inter).String()
	jsonObject := NewObject()
	if strings.HasPrefix(typeStr, "map") {
		// 如果是map
		mapRes, ok := inter.(map[string]interface{})
		if ok {
			jsonObject.ItemMap = mapRes
			return jsonObject, nil
		}
	}

	var str string
	var ok bool
	str, ok = inter.(string)
	if !ok {
		str = ToStr(inter)
	}
	err := json.Unmarshal([]byte(str), &jsonObject.ItemMap)
	if err != nil {
		Log.Error("string ParseJSONObject error, string is {}, err is {}", str, err)
		return nil, err
	}
	return jsonObject, nil
}

func ParseMap(inter interface{}) (map[string]interface{}, error) {
	typeStr := reflect.TypeOf(inter).String()
	if strings.HasPrefix(typeStr, "map") {
		// 如果是map
		mapRes, ok := inter.(map[string]interface{})
		if ok {
			return mapRes, nil
		}
	}

	var str string
	var ok bool
	str, ok = inter.(string)
	mapRes := make(map[string]interface{})
	if !ok {
		str = ToStr(inter)
	}
	err := json.Unmarshal([]byte(str), &mapRes)
	if err != nil {
		Log.Error("string ParseJSONObject error, string is {}, err is {}", str, err)
		return nil, err
	}
	return mapRes, nil
}

func Parse(str string, inter interface{}) {
	err := json.Unmarshal([]byte(str), inter)
	if err != nil {
		Log.Error("string Parse error, string is {}, err is {}", str, err)
	}
}

func ParseBytes(bytes []byte) *JSONObject {
	jsonObject := NewObject()
	err := json.Unmarshal(bytes, &jsonObject.ItemMap)
	if err != nil {
		Log.Error("string ParseBytes error,  err is {}", err)
	}
	return jsonObject
}

func ParseArray(str string, inter *[]interface{}) error {
	err := json.Unmarshal([]byte(str), inter)
	if err != nil {
		Log.Error("string ParseArray error, string is {}, err is {}", str, err)
		return err
	}
	return nil
}

func (jsonObject *JSONObject) ToJSONString() string {
	res, _ := json.Marshal(jsonObject.ItemMap)
	return string(res)
}

func ToJSONString(obj interface{}) string {
	res, _ := json.Marshal(obj)
	return string(res)
}

func (jsonObject *JSONObject) String() string {
	// res, _ := json.Marshal(jsonObject)
	// return string(res)
	return jsonObject.ToJSONString()
}

func ToStr(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
