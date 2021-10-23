package pool

import (
	"reflect"
	"runtime"
)

type zfunc struct {
	f interface{}
	p []interface{}
}

func (zf *zfunc) invoke() {
	kind := reflect.TypeOf(zf.f).Kind()
	if kind == reflect.Func {
		funcValue := reflect.ValueOf(zf.f)
		paramList := make([]reflect.Value, 0)
		for _, inter := range zf.p {
			paramList = append(paramList, reflect.ValueOf(inter))
		}
		funcValue.Call(paramList)
	}
}

type Zpool struct {
	waitCh chan *zfunc
	curCh  chan int8
}

func New() *Zpool {
	p := &Zpool{}
	p.curCh = make(chan int8, runtime.NumCPU()*2)
	p.waitCh = make(chan *zfunc, runtime.NumCPU()*8)
	p.start()
	return p
}

func NewWithNum(maxGoroutineNum int, waitQueueNum int) *Zpool {
	p := &Zpool{}
	p.curCh = make(chan int8, maxGoroutineNum)
	p.waitCh = make(chan *zfunc, waitQueueNum)
	p.start()
	return p
}
func (p *Zpool) GetCurNum() int {
	return len(p.curCh)
}

func (p *Zpool) GetWaitNum() int {
	return len(p.waitCh)
}

func (p *Zpool) Run(f interface{}, params ...interface{}) {
	zf := &zfunc{}
	zf.f = f
	zf.p = params
	p.waitCh <- zf
}

func (p *Zpool) start() {
	go func() {
		for {
			zf := <-p.waitCh
			p.curCh <- 1 // occupy a goroutine
			go func() {
				zf.invoke()
				_, ok := <-p.curCh // release a goroutine
				if !ok {
					panic("pick from curCh error")
				}
			}()
		}
	}()
}
