package lists

import (
	"encoding/json"
	"reflect"
	"sync"

	Log "github.com/wellmoon/go/logger"
)

type ArrayList struct {
	innerList []interface{}
	mutex     sync.Mutex
}

func NewArrayList() *ArrayList {
	return &ArrayList{}
}

func (arrayList *ArrayList) Add(val interface{}) {
	arrayList.mutex.Lock()
	arrayList.innerList = append(arrayList.innerList, val)
	arrayList.mutex.Unlock()
}

func (arrayList *ArrayList) AddAll(val *ArrayList) {
	arrayList.mutex.Lock()
	defer arrayList.mutex.Unlock()
	arrayList.innerList = append(arrayList.innerList, val.innerList...)
}

func (arrayList *ArrayList) Remove(val interface{}) interface{} {
	return arrayList.RemoveAt(arrayList.IndexOf(val))
}

func (arrayList *ArrayList) Contains(val interface{}) bool {
	return arrayList.IndexOf(val) != -1
}

func (arrayList *ArrayList) IndexOf(val interface{}) int {
	arrayList.mutex.Lock()
	defer arrayList.mutex.Unlock()
	for idx, e := range arrayList.innerList {
		if e == val {
			return idx
		}
	}
	return -1
}

func (arrayList *ArrayList) Size() int {
	return len(arrayList.innerList)
}

func (arrayList *ArrayList) RemoveAt(idx int) interface{} {
	if idx < 0 {
		return nil
	}
	arrayList.mutex.Lock()
	defer arrayList.mutex.Unlock()
	e := arrayList.innerList[idx]
	arrayList.innerList = append(arrayList.innerList[:idx], arrayList.innerList[idx+1:]...)
	return e
}

func (arrayList *ArrayList) Get(idx int) interface{} {
	return arrayList.innerList[idx]
}

func (arrayList *ArrayList) GetArray() []interface{} {
	return arrayList.innerList
}

func (arrayList *ArrayList) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(arrayList.innerList)
	return res, err
}

func StringSliceContain(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func ToArrayList(al interface{}) *ArrayList {

	kind := reflect.TypeOf(al).Kind()
	Log.Debug("kind is {}", kind)
	if kind == reflect.Ptr {
		return al.(*ArrayList)
	} else if kind == reflect.Slice {
		l := ArrayList{}
		l.innerList = al.([]interface{})
		return &l
	}

	switch value := al.(type) {
	case string:
		l := ArrayList{}
		json.Unmarshal([]byte(value), &l.innerList)
		return &l
	default:
		Log.Error("ToArrayList error for {}", value)
		return nil
	}
}

// sample
// delList.Each(func(idx int, v interface{}) {
//     do something
// })
func (arrayList *ArrayList) Each(f func(key int, val interface{})) {
	for idx, val := range arrayList.innerList {
		f(idx, val)
	}
}

// func (arrayList ArrayList) addAndGetIdx() int32 {
// 	return atomic.AddInt32(&arrayList.index, 1)
// }
