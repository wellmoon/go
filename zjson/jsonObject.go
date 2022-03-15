package zjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
)

type JSONObject struct {
	ItemMap map[string]interface{}
	lock    sync.Mutex
}

func NewObject() *JSONObject {
	// itemMap := make(map[string]interface{})
	newObj := &JSONObject{}
	newObj.ItemMap = make(map[string]interface{})
	return newObj
}

func (jsonObject *JSONObject) Put(key string, val interface{}) {
	// kind := reflect.TypeOf(val).Kind()
	// if kind == reflect.Ptr {
	// 	jsonObject.ItemMap[key] = val
	// } else {
	// 	jsonObject.ItemMap[key] = &val
	// }
	jsonObject.lock.Lock()
	defer jsonObject.lock.Unlock()
	jsonObject.ItemMap[key] = val
}

func (jsonObject *JSONObject) Remove(key string) {
	jsonObject.lock.Lock()
	defer jsonObject.lock.Unlock()
	delete(jsonObject.ItemMap, key)
}

func (jsonObject *JSONObject) Contains(key string) bool {
	jsonObject.lock.Lock()
	defer jsonObject.lock.Unlock()
	var _, ok = jsonObject.ItemMap[key]
	return ok
}

func (jsonObject *JSONObject) GetInt(key string) int {
	value := jsonObject.Get(key)
	res, err := ToInt(value)
	if err != nil {
		panic(err)
	}
	return res
}

func (jsonObject *JSONObject) GetInt64(key string) int64 {
	value := jsonObject.Get(key)
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
		return r, err
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case int32:
		return int(value), nil
	default:
		r, err := strconv.Atoi(ToStr(value))
		return r, err
	}
}

func ToInt32(value interface{}) (int32, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return int32(r), err
	case int:
		return int32(value), nil
	case int32:
		return value, nil
	case int64:
		return int32(value), nil
	default:
		r, err := strconv.Atoi(ToStr(value))
		if err != nil {
			return 0, err
		}
		return int32(r), err
	}
}

func ToInt64(value interface{}) (int64, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
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
			return 0, err
		}
		return int64(r), err
	}
}

func ToFloat64(value interface{}) (float64, error) {
	switch value := value.(type) {
	case string:
		r, err := strconv.ParseFloat(value, 64)
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
		return r, err
	}
}

func (jsonObject *JSONObject) GetString(key string) string {
	value := jsonObject.Get(key)
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
	value := jsonObject.Get(key)
	v := value
	val, ok := v.(float64)
	if !ok {
		panic("convert to float error")
	}
	return val
}

func (jsonObject *JSONObject) GetBool(key string) bool {
	value := jsonObject.Get(key)
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
	jsonObject.lock.Lock()
	defer jsonObject.lock.Unlock()
	return jsonObject.ItemMap[key]
}

func (jsonObject *JSONObject) GetArray(key string) ([]interface{}, error) {
	value := jsonObject.Get(key)
	v := value
	val, ok := v.([]interface{})
	if !ok {
		// 尝试把字符串转为slice
		str, sok := v.(string)
		if !sok {
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
			return nil, errors.New("convert to string array error")
		}
		strSlice = append(strSlice, str)
	}
	return strSlice, nil
}

func (jsonObject *JSONObject) GetJSONObject(key string) (*JSONObject, error) {
	value := jsonObject.Get(key)
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
	if kind == reflect.Map || kind == reflect.Ptr {
		mapRes, ok := inter.(map[string]interface{})
		if ok {
			jsonObject := NewObject()
			jsonObject.ItemMap = mapRes
			return jsonObject, nil
		} else {
			// cast to map fail, convert by bytes. but value object is a new object, memory address will be changed
			b, err := json.Marshal(inter)
			if err != nil {
				return nil, errors.New("can't convert map to JSONObject")
			}
			jsonObject := NewObject()
			err = json.Unmarshal(b, &jsonObject.ItemMap)
			if err != nil {
				return nil, errors.New("can't convert map to JSONObject")
			}
			return jsonObject, nil
		}

	} else if kind == reflect.Ptr {
		b, err := json.Marshal(inter)
		if err != nil {
			return nil, errors.New("can't convert to JSONObject")
		}
		jsonObject := NewObject()
		err = json.Unmarshal(b, &jsonObject.ItemMap)
		if err != nil {
			return nil, errors.New("can't convert to JSONObject")
		}
		return jsonObject, nil
	}

	jsonObject := NewObject()
	switch value := inter.(type) {
	case []byte:
		err := json.Unmarshal(value, &jsonObject.ItemMap)
		if err != nil {
			return nil, err
		}
		return jsonObject, nil
	case string:
		err := json.Unmarshal([]byte(value), &jsonObject.ItemMap)
		if err != nil {
			return nil, err
		}
		return jsonObject, nil
	default:
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
		return nil, err
	}
	return mapRes, nil
}

func Parse(str string, inter interface{}) error {
	err := json.Unmarshal([]byte(str), inter)
	if err != nil {
		return err
	}
	return nil
}

func ParseBytes(bytes []byte) *JSONObject {
	jsonObject := NewObject()
	err := json.Unmarshal(bytes, &jsonObject.ItemMap)
	if err != nil {
		return nil
	}
	return jsonObject
}

func ParseArray(str string, inter interface{}) error {
	err := json.Unmarshal([]byte(str), inter)
	if err != nil {
		return err
	}
	return nil
}

func (jsonObject *JSONObject) ToJSONString() string {
	res, _ := json.Marshal(jsonObject.ItemMap)
	return string(res)
}

func (jsonObject *JSONObject) ToBytes() []byte {
	res, _ := json.Marshal(jsonObject.ItemMap)
	return res
}

func (jsonObject *JSONObject) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(jsonObject.ItemMap)
	return res, err
}

func ToJSONString(obj interface{}) string {
	// kind := reflect.TypeOf(obj).Kind()
	// if kind == reflect.Slice {
	// 	l := ArrayList{}
	// 	l.innerList = al.([]interface{})
	// 	return &l
	// }
	res, _ := json.Marshal(obj)
	return string(res)
}

func (jsonObject *JSONObject) String() string {
	// res, _ := json.Marshal(jsonObject)
	// return string(res)
	return jsonObject.ToJSONString()
}

func (jsonObject *JSONObject) IsNull() bool {
	return jsonObject.ItemMap == nil
}

func IsEmpty(s string) bool {
	return len(s) == 0
}

func (jsonObject *JSONObject) Each(f func(key string, val interface{})) {
	jsonObject.lock.Lock()
	defer jsonObject.lock.Unlock()
	for key, val := range jsonObject.ItemMap {
		f(key, val)
	}
}
