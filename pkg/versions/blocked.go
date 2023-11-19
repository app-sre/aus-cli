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

package versions

import (
	"encoding/json"
	"io"
	"regexp"
	"sort"
	"strings"
)

type BlockedVersionUpdateMode int64

const (
	Modify BlockedVersionUpdateMode = iota
	Replace
)

func SortVersionExpressions(versionExpressions []string) []string {
	sort.Strings(versionExpressions)
	return versionExpressions
}

func ReadVersionExpressionsFromReader(reader io.Reader) ([]string, error) {
	var expressions []string
	err := json.NewDecoder(reader).Decode(&expressions)
	return expressions, err
}

func ConsolidateVersionBlocks(currentVersionBlocks []string, toBlock []string, toUnblock []string) []string {
	stringMap := make(map[string]bool)
	for _, version := range currentVersionBlocks {
		if normalizedVersion, ok := normalizeVersionBlock(version); ok {
			stringMap[normalizedVersion] = true
		}
	}
	for _, version := range toBlock {
		if normalizedVersion, ok := normalizeVersionBlock(version); ok {
			stringMap[normalizedVersion] = true
		}
	}
	for _, version := range toUnblock {
		if normalizedVersion, ok := normalizeVersionBlock(version); ok {
			delete(stringMap, normalizedVersion)
		}
	}

	result := []string{}
	for s := range stringMap {
		result = append(result, s)
	}

	return result
}

func ParsedBlockedVersionExpressions(blockedVersions []string) ([]*regexp.Regexp, error) {
	var blockedVersionExpressions []*regexp.Regexp
	for _, blockedVersion := range blockedVersions {
		blockedVersionExpression, err := regexp.Compile(blockedVersion)
		if err != nil {
			return nil, err
		}
		blockedVersionExpressions = append(blockedVersionExpressions, blockedVersionExpression)
	}
	return blockedVersionExpressions, nil
}

func IsVersionBlocked(version string, blockedVersions []*regexp.Regexp) bool {
	for _, blockedVersion := range blockedVersions {
		if blockedVersion.MatchString(version) {
			return true
		}
	}
	return false
}

func normalizeVersionBlock(version string) (string, bool) {
	normalized := strings.Trim(version, " ,")
	if normalized == "" {
		return "", false
	}
	return normalized, true
}
