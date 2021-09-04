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
	kind := reflect.TypeOf(val).Kind()
	if kind == reflect.Ptr {
		jsonObject.ItemMap[key] = val
	} else {
		jsonObject.ItemMap[key] = &val
	}
}

func (jsonObject *JSONObject) Contains(key string) bool {
	var _, ok = jsonObject.ItemMap[key]
	return ok
}

func (jsonObject *JSONObject) GetInt(key string) int {
	value := jsonObject.ItemMap[key]
	res, err := ToInt(value)
	if err != nil {
		panic(err)
	}
	return res
}

func (jsonObject *JSONObject) GetInt64(key string) int64 {
	value := jsonObject.ItemMap[key]
	res, err := ToInt64(value)
	if err != nil {
		panic(err)
	}
	return res
}

func ToInt(value interface{}) (int, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			Log.Debug("GetInt by strconv.Atoi error: {}", err)
		}
		return r, err
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case int32:
		return int(value), nil
	default:
		r, err := strconv.Atoi(ToStr(value))
		if err != nil {
			Log.Debug("GetInt error for type {}", reflect.TypeOf(value))
		}
		return r, err
	}
}

func ToInt64(value interface{}) (int64, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			Log.Debug("GetInt64 by strconv.Atoi error: {}", err)
		}
		return int64(r), err
	case int:
		return int64(value), nil
	case int64:
		return value, nil
	case int32:
		return int64(value), nil
	default:
		r, err := strconv.Atoi(ToStr(value))
		if err != nil {
			Log.Debug("GetInt64 error for type {}", reflect.TypeOf(value))
		}
		return int64(r), err
	}
}

func ToFloat64(value interface{}) (float64, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.ParseFloat(value, 64)
		if err != nil {
			Log.Debug("ToFloat64 error : {}", err)
		}
		return r, err
	case int:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	default:
		r, err := strconv.ParseFloat(ToStr(value), 64)
		if err != nil {
			Log.Debug("ToFloat64 error : {}", err)
		}
		return r, err
	}
}

func (jsonObject *JSONObject) GetString(key string) string {
	value := jsonObject.ItemMap[key]
	return ToStr(value)
}

func ToStr(inter interface{}) string {
	if inter == nil {
		return ""
	}
	switch value := inter.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case int64:
		return fmt.Sprintf("%v", value)
	case float64:
		return decimal.NewFromFloat(value).String()
	case *interface{}:
		s, _ := interfaceToString(inter)
		return s
	default:
		s, err := interfaceToString(inter)
		if err != nil {
			Log.Debug("GetString for type {}", reflect.TypeOf(value))
			json.Marshal(value)
			return fmt.Sprintf("%v", value)
		}
		return s
	}
}

func interfaceToString(inter interface{}) (string, error) {
	s, _ := json.Marshal(inter)
	return string(s), nil
}

// func invokeMarshalJSON(inter interface{}) (string, error) {
// 	v := reflect.ValueOf(inter)
// 	switch al := inter.(type) {
// 	case *interface{}:
// 		v = reflect.ValueOf(*al)
// 	}
// 	ret := v.MethodByName("MarshalJSON").Call([]reflect.Value{})
// 	if err, ok := ret[1].Interface().(error); ok {
// 		Log.Error("reflect invoke MarshalJSON error")
// 		return "", err
// 	}
// 	b := ret[0].Interface().([]byte)
// 	return string(b), nil
// }

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

func (jsonObject *JSONObject) GetStringArray(key string) ([]string, error) {
	arr, err := jsonObject.GetArray(key)
	if err != nil {
		return nil, err
	}
	strSlice := make([]string, 0)
	for _, v := range arr {
		str, sok := v.(string)
		if !sok {
			Log.Error("convert {} to string error", v)
			return nil, err
		}
		strSlice = append(strSlice, str)
	}
	return strSlice, nil
}

func (jsonObject *JSONObject) GetJSONObject(key string) (*JSONObject, error) {
	value := jsonObject.ItemMap[key]
	// val, ok := value.(string)
	// if !ok {
	// 	Log.Fatal("convert to JSONObject error, value is {}", value)
	// }
	result, err := ParseJSONObject(value)
	return result, err
}

func (jsonObject *JSONObject) Size() int {
	return len(jsonObject.ItemMap)
}

func ParseJSONObject(inter interface{}) (*JSONObject, error) {
	if inter == nil {
		return nil, errors.New("ParseJSONObject param is nil")
	}

	kind := reflect.TypeOf(inter).Kind()
	if kind == reflect.Map {
		// 如果是map
		mapRes, ok := inter.(map[string]interface{})
		if ok {
			jsonObject := NewObject()
			jsonObject.ItemMap = mapRes
			return jsonObject, nil
		}
	} else if kind == reflect.Ptr {
		res, ok := inter.(*JSONObject)
		if ok {
			return res, nil
		}
	}

	jsonObject := NewObject()
	switch value := inter.(type) {
	case string:
		err := json.Unmarshal([]byte(value), &jsonObject.ItemMap)
		if err != nil {
			Log.Error("string ParseJSONObject error, string is {}, err is {}", value, err)
			return nil, err
		}
		return jsonObject, nil
	default:
		Log.Error("string ParseJSONObject error, type is {}", value)
		return nil, errors.New("ParseJSONObject error, type is not correct")
	}

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

func ParseArray(str string, inter interface{}) error {
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

func (jsonObject *JSONObject) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(jsonObject.ItemMap)
	return res, err
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

func IsEmpty(s string) bool {
	return len(s) == 0
}

func (jsonObject *JSONObject) Each(f func(key string, val interface{})) {
	for key, val := range jsonObject.ItemMap {
		f(key, val)
	}
}
