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

package versiondata

import (
	"encoding/json"
	"io"
)

type VersionDataInheritanceUpdateMode int64

const (
	Modify VersionDataInheritanceUpdateMode = iota
	Replace
)

type VersionDataInheritanceConfig struct {
	InheritingFromOrgs []string `json:"inherit,omitempty"`
	PublishingToOrgs   []string `json:"publish,omitempty"`
}

func NewVersionDataInheritanceConfigFromReader(reader io.Reader) (VersionDataInheritanceConfig, error) {
	var config = VersionDataInheritanceConfig{}
	err := json.NewDecoder(reader).Decode(&config)
	return config, err
}

func ConsolidateVersionDataInheritanceConfig(current VersionDataInheritanceConfig, desired VersionDataInheritanceConfig) VersionDataInheritanceConfig {
	// merge the inheriting orgs lists without duplicates
	inheritingFromOrgsMap := make(map[string]bool)
	for _, s := range current.InheritingFromOrgs {
		inheritingFromOrgsMap[s] = true
	}
	for _, s := range desired.InheritingFromOrgs {
		inheritingFromOrgsMap[s] = true
	}
	inheritingFromOrgs := []string{}
	for s := range inheritingFromOrgsMap {
		inheritingFromOrgs = append(inheritingFromOrgs, s)
	}

	// merge the publishing orgs lists without duplicates
	publishingToOrgsMap := make(map[string]bool)
	for _, s := range current.PublishingToOrgs {
		publishingToOrgsMap[s] = true
	}
	for _, s := range desired.PublishingToOrgs {
		publishingToOrgsMap[s] = true
	}
	publishingToOrgs := []string{}
	for s := range publishingToOrgsMap {
		publishingToOrgs = append(publishingToOrgs, s)
	}

	return VersionDataInheritanceConfig{
		InheritingFromOrgs: inheritingFromOrgs,
		PublishingToOrgs:   publishingToOrgs,
	}
}
