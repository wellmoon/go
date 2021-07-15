package lists

import (
	"sync"
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

// func (arrayList ArrayList) addAndGetIdx() int32 {
// 	return atomic.AddInt32(&arrayList.index, 1)
// }
