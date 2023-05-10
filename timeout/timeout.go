// Copyright 2023 The acquirecloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package timeout

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type (
	// Future object allows to cancel a future execution request made by Call()
	Future interface {
		// Cancel allows to cancel a future execution.
		Cancel()
	}

	callControl struct {
		lock     sync.Mutex
		wakeCh   chan bool
		futures  *futures
		watchers int
	}

	future struct {
		f     func()
		fireT time.Time
		idx   int
	}

	futures []*future

	dummyFuture struct{}
)

func init() {
	cc = new(callControl)
	cc.futures = &futures{}
	cc.wakeCh = make(chan bool, 100)
	heap.Init(cc.futures)
}

var cc *callControl

// VoidFuture maybe used to initialize a Future variable, without checking whether it is nil or not
var VoidFuture Future = dummyFuture{}

// Call allows scheduling future execution of the function f in timeout provided.
// The function returns the Future object, which may be used for cancelling the execution
// request if needed.
func Call(f func(), timeout time.Duration) Future {
	fu := new(future)
	fu.f = f
	fu.fireT = time.Now().Add(timeout)
	fu.idx = -1
	if f != nil {
		cc.add(fu)
	}
	return fu
}

// Cancel cancels the future execution if not called yet
func (fu *future) Cancel() {
	cc.cancel(fu)
}

// String implements fmt.Stringify
func (fu *future) String() string {
	return cc.futureAsString(fu)
}

func (cc *callControl) add(fu *future) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	heap.Push(cc.futures, fu)
	if cc.watchers == 0 {
		cc.watchers++
		go cc.watcher(nil)
	} else {
		cc.notifyWatcher()
	}
}

func (cc *callControl) futureAsString(fu *future) string {
	cc.lock.Lock()
	f := "<not assigned>"
	if fu.f != nil {
		f = "<assigned>"
	}
	charged := true
	if fu.idx < 0 {
		charged = false
	}
	res := fmt.Sprintf("{?f: %s, fireT: %v, chanrged: %t}", f, fu.fireT, charged)
	cc.lock.Unlock()
	return res
}

func (cc *callControl) cancel(fu *future) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if fu.idx < 0 {
		return
	}
	fu.f = nil
	heap.Remove(cc.futures, fu.idx)
	if cc.watchers > 0 {
		cc.notifyWatcher()
	}
}

func (cc *callControl) notifyWatcher() {
	select {
	case cc.wakeCh <- true:
	default:
	}
}

func (cc *callControl) watcher(f func()) {
	misCount := false
	for {
		if f != nil {
			f()
		}
		f = nil

		var tmt time.Duration
		cc.lock.Lock()
		if cc.futures.Len() == 0 {
			if cc.watchers > 1 || misCount {
				cc.watchers--
				cc.lock.Unlock()
				return
			}
			misCount = true
			tmt = time.Second * 30
		} else {
			misCount = false
			fireT := (*cc.futures)[0].fireT
			now := time.Now()
			if now.After(fireT) {
				fu := heap.Pop(cc.futures).(*future)
				f = fu.f
				if cc.watchers < 10 {
					cc.watchers++
					go cc.watcher(f)
					f = nil
				}
				cc.lock.Unlock()
				continue
			}
			if cc.watchers > 1 {
				cc.watchers--
				cc.lock.Unlock()
				return
			}
			tmt = fireT.Sub(now)
		}
		cc.lock.Unlock()

		tmr := time.NewTimer(tmt)
		select {
		case <-tmr.C:
		case <-cc.wakeCh:
			if !tmr.Stop() {
				<-tmr.C
			}
			misCount = false
		}
	}
}

func (fs *futures) Len() int {
	return len(*fs)
}

func (fs *futures) Less(i, j int) bool {
	fi := (*fs)[i]
	fj := (*fs)[j]
	return fi.fireT.Before(fj.fireT)
}

func (fs *futures) Swap(i, j int) {
	(*fs)[i], (*fs)[j] = (*fs)[j], (*fs)[i]
	(*fs)[i].idx, (*fs)[j].idx = i, j
}

func (fs *futures) Push(x any) {
	fu := x.(*future)
	fu.idx = fs.Len()
	(*fs) = append(*fs, fu)
}

func (fs *futures) Pop() any {
	last := fs.Len() - 1
	res := (*fs)[last]
	(*fs)[last] = nil
	(*fs) = (*fs)[:last]
	res.idx = -1
	return res
}

func (d dummyFuture) Cancel() {
	// Do nothing
}
