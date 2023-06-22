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
	"sort"

	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterInfo struct {
	Subscription *amv1.Subscription
	Cluster      *csv1.Cluster
	Policy       *ClusterUpgradePolicy
}

type ClusterUpgradePolicy struct {
	ClusterName string                         `json:"name"`
	Schedule    string                         `json:"schedule"`
	Workloads   []string                       `json:"workloads"`
	Conditions  ClusterUpgradePolicyConditions `json:"conditions"`
}

type ClusterUpgradePolicyConditions struct {
	SoakDays int      `json:"soak_days"`
	Sector   string   `json:"sector,omitempty"`
	Mutexes  []string `json:"mutexes,omitempty"`
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

func NewClusterUpgradePolicy(clusterName string, schedule string, workloads []string, soakDays int, sector string, mutexes []string) ClusterUpgradePolicy {
	return ClusterUpgradePolicy{
		ClusterName: clusterName,
		Schedule:    schedule,
		Workloads:   workloads,
		Conditions: ClusterUpgradePolicyConditions{
			SoakDays: soakDays,
			Sector:   sector,
			Mutexes:  mutexes,
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
