/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type PrefixWriter struct {
	out    io.Writer
	prefix string
}

type flusher interface {
	Flush()
}

func NewPrefixWriter(out io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{out: out, prefix: prefix}
}

func (pw *PrefixWriter) Write(p []byte) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if len(line) > 0 {
			fmt.Fprintf(pw.out, "%s%s", pw.prefix, line)
			if i > 0 {
				fmt.Fprintln(pw.out)
			}
		} else {
			fmt.Fprintln(pw.out)
		}
	}
}

func (pw *PrefixWriter) Flush() {
	if f, ok := pw.out.(flusher); ok {
		f.Flush()
	}
}

func (pw *PrefixWriter) WriteString(format string, a ...interface{}) {
	pw.Write([]byte(fmt.Sprintf(format, a...)))
}

func PrintListMultiline(w *PrefixWriter, title string, values []string) {
	w.WriteString("%s:\t", title)
	if len(values) == 0 {
		w.WriteString("<none>\n")
		return
	}
	indent := "\t"
	for i, value := range values {
		if i != 0 {
			w.WriteString(indent)
		}
		w.WriteString("%s\n", value)
	}
}

func TabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	if err := f(out); err != nil {
		return "", err
	}

	if err := out.Flush(); err != nil {
		return "", err
	}

	str := buf.String()
	return str, nil
}
