// Copyright 2021 ByteDance Inc.
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

package skipmap

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anishathalye/porcupine"
)

const (
	opLoad = iota
	opLoadOrStore
	opStore
	opLoadAndDelete
	numOps
)

type skipmapInput struct {
	Op    int
	Key   string
	Value int
}

type skipmapOutput struct {
	Value int
	Ok    bool
}

type perKeyState struct {
	Exists bool
	Value  int
}

var skipmapModel = porcupine.Model{
	Init: func() interface{} { return perKeyState{} },
	Step: func(state, input, output interface{}) (bool, interface{}) {
		st := state.(perKeyState)
		inp := input.(skipmapInput)
		out := output.(skipmapOutput)

		switch inp.Op {
		case opLoad:
			if st.Exists {
				return out.Ok && out.Value == st.Value, st
			}
			return !out.Ok && out.Value == 0, st
		case opLoadOrStore:
			if st.Exists {
				return out.Ok && out.Value == st.Value, st
			}
			if !out.Ok && out.Value == inp.Value {
				return true, perKeyState{Exists: true, Value: inp.Value}
			}
			return false, st
		case opStore:
			return true, perKeyState{Exists: true, Value: inp.Value}
		case opLoadAndDelete:
			if st.Exists {
				if out.Ok && out.Value == st.Value {
					return true, perKeyState{}
				}
				return false, st
			}
			return !out.Ok && out.Value == 0, st
		default:
			return false, st
		}
	},
	Equal: func(a, b interface{}) bool {
		return a.(perKeyState) == b.(perKeyState)
	},
	DescribeOperation: func(input, output interface{}) string {
		inp := input.(skipmapInput)
		out := output.(skipmapOutput)
		switch inp.Op {
		case opLoad:
			if out.Ok {
				return fmt.Sprintf("Load(%s) -> %d", inp.Key, out.Value)
			}
			return fmt.Sprintf("Load(%s) -> nil", inp.Key)
		case opLoadOrStore:
			if out.Ok {
				return fmt.Sprintf("LoadOrStore(%s, %d) -> loaded %d", inp.Key, inp.Value, out.Value)
			}
			return fmt.Sprintf("LoadOrStore(%s, %d) -> stored", inp.Key, inp.Value)
		case opStore:
			return fmt.Sprintf("Store(%s, %d)", inp.Key, inp.Value)
		case opLoadAndDelete:
			if out.Ok {
				return fmt.Sprintf("LoadAndDelete(%s) -> %d", inp.Key, out.Value)
			}
			return fmt.Sprintf("LoadAndDelete(%s) -> nil", inp.Key)
		default:
			return "?"
		}
	},
	DescribeState: func(state interface{}) string {
		st := state.(perKeyState)
		if st.Exists {
			return fmt.Sprintf("value=%d", st.Value)
		}
		return "absent"
	},
}

func executeSkipmapOp(m *StringMap, input skipmapInput) skipmapOutput {
	switch input.Op {
	case opLoad:
		value, ok := m.Load(input.Key)
		if ok {
			return skipmapOutput{Value: value.(int), Ok: true}
		}
		return skipmapOutput{}
	case opLoadOrStore:
		actual, loaded := m.LoadOrStore(input.Key, input.Value)
		return skipmapOutput{Value: actual.(int), Ok: loaded}
	case opStore:
		m.Store(input.Key, input.Value)
		return skipmapOutput{}
	case opLoadAndDelete:
		value, loaded := m.LoadAndDelete(input.Key)
		if loaded {
			return skipmapOutput{Value: value.(int), Ok: true}
		}
		return skipmapOutput{}
	default:
		return skipmapOutput{}
	}
}

func checkPerKeyLinearizability(operations []porcupine.Operation, timeout time.Duration) (bool, []porcupine.LinearizationInfo) {
	byKey := make(map[string][]porcupine.Operation)
	for _, op := range operations {
		byKey[op.Input.(skipmapInput).Key] = append(byKey[op.Input.(skipmapInput).Key], op)
	}

	var mu sync.Mutex
	var failures []porcupine.LinearizationInfo
	var wg sync.WaitGroup

	for _, ops := range byKey {
		wg.Add(1)
		go func(ops []porcupine.Operation) {
			defer wg.Done()
			result, info := porcupine.CheckOperationsVerbose(skipmapModel, ops, timeout)
			if result == porcupine.Illegal {
				mu.Lock()
				failures = append(failures, info)
				mu.Unlock()
			}
		}(ops)
	}
	wg.Wait()

	return len(failures) == 0, failures
}

// TestSkipMapLinearizability runs a mixed linearizability test with concurrent
// clients exercising all skipmap operations on a shared key set.
func TestSkipMapLinearizability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping linearizability tests in short mode")
	}

	const (
		numClients      = 16
		numOpsPerClient = 200
		numKeys         = 10
	)

	m := NewString()
	localOps := make([][]porcupine.Operation, numClients)
	var wg sync.WaitGroup
	startTime := time.Now()

	for clientId := 0; clientId < numClients; clientId++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()
			ops := make([]porcupine.Operation, 0, numOpsPerClient)

			for opId := 0; opId < numOpsPerClient; opId++ {
				key := fmt.Sprintf("key_%d", rand.Intn(numKeys))
				value := cid*100000 + opId
				op := rand.Intn(numOps)
				input := skipmapInput{Op: op, Key: key, Value: value}

				var callTime int64
				atomic.StoreInt64(&callTime, time.Since(startTime).Nanoseconds())
				output := executeSkipmapOp(m, input)
				returnTime := time.Since(startTime).Nanoseconds()

				ops = append(ops, porcupine.Operation{
					ClientId: cid,
					Input:    input,
					Call:     callTime,
					Output:   output,
					Return:   returnTime,
				})
			}
			localOps[cid] = ops
		}(clientId)
	}
	wg.Wait()

	var operations []porcupine.Operation
	for _, ops := range localOps {
		operations = append(operations, ops...)
	}

	ok, failures := checkPerKeyLinearizability(operations, 30*time.Second)
	if !ok {
		for i, info := range failures {
			filename := fmt.Sprintf("violation_skipmap_%d_%s.html", i, time.Now().Format("20060102_150405"))
			if file, err := os.Create(filename); err == nil {
				porcupine.Visualize(skipmapModel, info, file)
				file.Close()
				t.Logf("visualization written to %s", filename)
			}
		}
		t.Fatalf("linearizability violation: %d key(s) failed", len(failures))
	}
}
