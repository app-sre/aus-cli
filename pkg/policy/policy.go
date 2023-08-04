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

package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.com/app-sre/aus-cli/pkg/blockedversions"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterInfo struct {
	Subscription *amv1.Subscription
	Cluster      *csv1.Cluster
	Policy       *ClusterUpgradePolicy
}

func (c ClusterInfo) blockedVersionExpressions() ([]*regexp.Regexp, error) {
	return blockedversions.ParsedBlockedVersionExpressions(c.Policy.Conditions.BlockedVersions)
}

func (c ClusterInfo) AvailableUpgrades(additionalBlockedVersions []*regexp.Regexp) []string {
	var upgrades []string
	clusterBlockedVersionExpressions, error := c.blockedVersionExpressions()
	if error != nil {
		return nil
	}
	for _, version := range c.Cluster.Version().AvailableUpgrades() {
		// check if the version is blocked on the cluster
		if isVersionBlocked(version, clusterBlockedVersionExpressions) {
			continue
		}
		// check if the version is blocked by other blockers (e.g. org level blockers)
		if isVersionBlocked(version, additionalBlockedVersions) {
			continue
		}
		upgrades = append(upgrades, version)
	}
	return upgrades
}

func isVersionBlocked(version string, blockedVersions []*regexp.Regexp) bool {
	for _, blockedVersion := range blockedVersions {
		if blockedVersion.MatchString(version) {
			return true
		}
	}
	return false
}

type ClusterUpgradePolicy struct {
	ClusterName string                         `json:"name"`
	Schedule    string                         `json:"schedule"`
	Workloads   []string                       `json:"workloads"`
	Conditions  ClusterUpgradePolicyConditions `json:"conditions"`
}

type ClusterUpgradePolicyConditions struct {
	SoakDays        int      `json:"soak_days"`
	Sector          string   `json:"sector,omitempty"`
	Mutexes         []string `json:"mutexes,omitempty"`
	BlockedVersions []string `json:"blocked_versions,omitempty"`
}

func (p ClusterUpgradePolicy) Validate() error {
	if p.ClusterName == "" {
		return fmt.Errorf("cluster name is required")
	}
	if p.Conditions.SoakDays < 0 {
		return fmt.Errorf("soak-days must be >= 0")
	}
	if len(p.Workloads) == 0 {
		return fmt.Errorf("workloads are required")
	}
	if p.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}
	return nil
}

func NewClusterUpgradePolicy(clusterName string, schedule string, workloads []string, soakDays int, sector string, mutexes []string, blockedVersions []string) ClusterUpgradePolicy {
	return ClusterUpgradePolicy{
		ClusterName: clusterName,
		Schedule:    schedule,
		Workloads:   workloads,
		Conditions: ClusterUpgradePolicyConditions{
			SoakDays:        soakDays,
			Sector:          sector,
			Mutexes:         mutexes,
			BlockedVersions: blockedVersions,
		},
	}
}

func NewClusterUpgradePolicyFromReader(reader io.Reader) ([]ClusterUpgradePolicy, error) {
	var policies []ClusterUpgradePolicy
	err := json.NewDecoder(reader).Decode(&policies)
	return policies, err
}

func SortPolicies(policies []ClusterUpgradePolicy) {
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].ClusterName < policies[j].ClusterName
	})
}

func SortClusters(clusters []ClusterInfo) {
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Cluster.Name() < clusters[j].Cluster.Name()
	})
}
