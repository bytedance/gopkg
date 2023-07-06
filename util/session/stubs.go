/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package session

import (
	"strconv"
	_ "unsafe"

	"github.com/modern-go/gls"
)

//go:nocheckptr
func goID() uint64 {
	return uint64(gls.GoID())
}

type labelMap map[string]string

//go:linkname setPprofLabel runtime/pprof.runtime_setProfLabel
func setPprofLabel(m *labelMap)

//go:linkname getPproLabel runtime/pprof.runtime_getProfLabel
func getPproLabel() *labelMap

const Pprof_Label_Session_ID = "go_session_id"

func transmitSessionID(id SessionID) {
	m := getPproLabel()

	var n labelMap
	if m == nil {
		n = make(labelMap)
	} else {
		n = make(labelMap, len(*m))
		for k, v := range *m {
			if k != Pprof_Label_Session_ID {
				n[k] = v
			}
		}
	}
	
	n[Pprof_Label_Session_ID] = strconv.FormatInt(int64(id), 10)
	setPprofLabel(&n)
}

func getSessionID() (SessionID, bool) {
	m := getPproLabel()
	if m == nil {
		return 0, false
	}
	if v, ok := (*m)[Pprof_Label_Session_ID]; !ok {
		return 0, false
	} else {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return SessionID(id), true
	}
}

func clearSessionID() {
	m := getPproLabel()
	if m == nil {
		return 
	}
	if _, ok := (*m)[Pprof_Label_Session_ID]; !ok {
		return
	}
	n := make(labelMap, len(*m))
	for k, v := range *m {
		if k != Pprof_Label_Session_ID {
			n[k] = v
		}
	}
	setPprofLabel(&n)
}
