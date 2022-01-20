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

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"go/format"
	"io"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("skipset.go")
	if err != nil {
		panic(err)
	}
	filedata, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	w := new(bytes.Buffer)
	w.WriteString(`// Copyright 2021 ByteDance Inc.
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

`)
	// Step 1. Add file header
	w.WriteString(`// Code generated by go run types_gen.go; DO NOT EDIT.` + "\r\n")
	// Step 2. Add imports and package statement
	w.WriteString(string(filedata)[strings.Index(string(filedata), "package skipset") : strings.Index(string(filedata), ")\n")+1])
	// Step 3. Generate code for all basic types
	ts := []string{"Float32", "Float64", "Int32", "Int16", "Int", "Uint64", "Uint32", "Uint16", "Uint"} // all types need to be converted
	for _, upper := range ts {
		data := string(filedata)
		// Step 4-1. Remove all string before import
		data = data[strings.Index(data, ")\n")+1:]
		// Step 4-2. Replace all cases
		dataDesc := replace(data, upper, true)
		dataAsc := replace(data, upper, false)
		w.WriteString(dataAsc)
		w.WriteString("\r\n")
		w.WriteString(dataDesc)
		w.WriteString("\r\n")
	}
	// Step 5. Generate string set
	data := string(filedata)
	data = data[strings.Index(data, ")\n")+1:]
	datastring := replaceString(data)
	w.WriteString(datastring)
	w.WriteString("\r\n")

	out, err := format.Source(w.Bytes())
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("types.go", out, 0660); err != nil {
		panic(err)
	}
}

func replace(data string, upper string, desc bool) string {
	lower := strings.ToLower(upper)

	var descstr string
	if desc {
		descstr = "Desc"
	}
	data = strings.Replace(data, "NewInt64", "New"+upper+descstr, -1)
	data = strings.Replace(data, "newInt64Node", "new"+upper+"Node"+descstr, -1)
	data = strings.Replace(data, "unlockInt64", "unlock"+upper+descstr, -1)
	data = strings.Replace(data, "Int64Set", upper+"Set"+descstr, -1)
	data = strings.Replace(data, "int64Node", lower+"Node"+descstr, -1)
	data = strings.Replace(data, "value int64", "value "+lower, -1)
	data = strings.Replace(data, "int64 skip set", lower+" skip set", -1) // comment

	if desc {
		// Special cases for DESC.
		data = strings.Replace(data, "ascending", "descending", -1)
		data = strings.Replace(data, "return n.value < value", "return n.value > value", -1)
	}
	return data
}

func replaceString(data string) string {
	const (
		upper = "String"
		lower = "string"
	)

	// Add `score uint64` field.
	data = strings.Replace(data,
		`type int64Node struct {
	value int64`,
		`type int64Node struct {
	value int64
	score uint64`, -1)

	data = strings.Replace(data,
		`&int64Node{`,
		`&int64Node{
		score: hash(value),`, -1)

	// Refactor comparsion.
	data = data + "\n"
	data += `// Return 1 if n is bigger, 0 if equal, else -1.
func (n *stringNode) cmp(score uint64, value string) int {
	if n.score > score {
		return 1
	} else if n.score == score {
		return cmpstring(n.value, value)
	}
	return -1
}`

	data = strings.Replace(data,
		`.lessthan(value)`,
		`.cmp(score, value) < 0`, -1)
	data = strings.Replace(data,
		`.equal(value)`,
		`.cmp(score, value) == 0`, -1)

	// Remove `lessthan` and `equal`
	data = strings.Replace(data,
		`func (n *int64Node) lessthan(value int64) bool {
	return n.value < value
}`, "", -1)
	data = strings.Replace(data,
		`func (n *int64Node) equal(value int64) bool {
	return n.value == value
}`, "", -1)

	// Add "score := hash(value)"
	data = addLineAfter(data, "func (s *Int64Set) findNodeRemove", "score := hash(value)")
	data = addLineAfter(data, "func (s *Int64Set) findNodeAdd", "score := hash(value)")
	data = addLineAfter(data, "func (s *Int64Set) Contains", "score := hash(value)")

	// Update new value "newInt64Node(0"
	data = strings.Replace(data,
		"newInt64Node(0", `newInt64Node(""`, -1)

	data = strings.Replace(data, "NewInt64", "New"+upper, -1)
	data = strings.Replace(data, "newInt64Node", "new"+upper+"Node", -1)
	data = strings.Replace(data, "unlockInt64", "unlock"+upper, -1)
	data = strings.Replace(data, "Int64Set", upper+"Set", -1)
	data = strings.Replace(data, "int64Node", lower+"Node", -1)
	data = strings.Replace(data, "value int64", "value "+lower, -1)
	data = strings.Replace(data, "int64 skip set", lower+" skip set", -1) // comment
	data = strings.Replace(data, " in ascending order", "", -1)           // comment

	return data
}

func lowerSlice(s []string) []string {
	n := make([]string, len(s))
	for i, v := range s {
		n[i] = strings.ToLower(v)
	}
	return n
}

func inSlice(s []string, val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}

func addLineAfter(src string, after string, added string) string {
	all := strings.Split(string(src), "\n")
	for i, v := range all {
		if strings.Index(v, after) != -1 {
			res := make([]string, len(all)+1)
			for j := 0; j <= i; j++ {
				res[j] = all[j]
			}
			res[i+1] = added
			for j := i + 1; j < len(all); j++ {
				res[j+1] = all[j]
			}
			return strings.Join(res, "\n")
		}
	}
	panic("can not find:" + after)
}
