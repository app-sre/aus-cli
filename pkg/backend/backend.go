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

package backend

import (
	"fmt"

	"github.com/app-sre/aus-cli/pkg/backend/ocmlabels"
	"github.com/app-sre/aus-cli/pkg/policy"
	"github.com/app-sre/aus-cli/pkg/sectors"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

type PolicyBackend interface {
	ListPolicies(organizationId string, showClustersWithoutPolicy bool) (map[string]*policy.ClusterInfo, error)

	ApplyPolicies(organizationId string, policies []policy.ClusterUpgradePolicy, dumpPolicy bool, dryRun bool) error

	DeletePolicy(organizationId string, clusterName string, dryRun bool) error

	ListBlockedVersionExpressions(organizationId string) ([]string, error)

	ApplyBlockedVersionExpressions(organizationId string, blockExpressions []string, dumpVersionBlocks bool, dryRun bool) error

	ListSectorConfiguration(organizationId string) ([]sectors.SectorDependencies, error)

	ApplySectorConfiguration(organizationId string, sectorDependencies []sectors.SectorDependencies, dumpSectorDeps bool, dryRun bool) error

	Status(organizationId string, showClustersWithoutPolicy bool) (organization *amv1.Organization, clusters []policy.ClusterInfo, blockedVersions []string, sectors []sectors.SectorDependencies, err error)
}

func NewPolicyBackend(backendType string) (PolicyBackend, error) {
	switch backendType {
	case "ocmlabels", "":
		return ocmlabels.NewOCMLabelsPolicyBackend(), nil
	default:
		return nil, fmt.Errorf("unknown backend type: %s", backendType)
	}
}
