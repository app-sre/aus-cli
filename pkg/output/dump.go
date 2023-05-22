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
	"encoding/json"
	"io"

	"github.com/nwidger/jsoncolor"
	"github.com/openshift-online/ocm-cli/pkg/output"
)

// Pretty dumps the given data to the given stream so that it looks pretty. If the data is a valid
// JSON document then it will be indented before printing it. If the stream is a terminal then the
// output will also use colors.
func PrettyList(stream io.Writer, body []byte) error {
	if len(body) == 0 {
		return nil
	}
	data := make([]interface{}, 0)
	err := json.Unmarshal(body, &data)
	if err != nil {
		return dumpBytes(stream, body)
	}
	if output.IsTerminal(stream) {
		return dumpColor(stream, data)
	}
	return dumpMonochrome(stream, data)
}

func dumpColor(stream io.Writer, data interface{}) error {
	encoder := jsoncolor.NewEncoder(stream)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func dumpMonochrome(stream io.Writer, data interface{}) error {
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func dumpBytes(stream io.Writer, data []byte) error {
	_, err := stream.Write(data)
	if err != nil {
		return err
	}
	_, err = stream.Write([]byte("\n"))
	return err
}
