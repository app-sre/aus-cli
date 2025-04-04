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
	"github.com/app-sre/aus-cli/pkg/clusters"
	"github.com/app-sre/aus-cli/pkg/policy"
	"github.com/app-sre/aus-cli/pkg/sectors"
	"github.com/app-sre/aus-cli/pkg/versiondata"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

type PolicyBackend interface {
	ListPolicies(organizationId string, showClustersWithoutPolicy bool) (map[string]*clusters.ClusterInfo, error)

	ApplyPolicies(organizationId string, policies []policy.ClusterUpgradePolicy, dumpPolicy bool, dryRun bool) error

	DeletePolicy(organizationId string, clusterName string, dryRun bool) error

	ListBlockedVersionExpressions(organizationId string) ([]string, error)

	ApplyBlockedVersionExpressions(organizationId string, blockExpressions []string, dumpVersionBlocks bool, dryRun bool) error

	ListSectorConfiguration(organizationId string) ([]sectors.Sector, error)

	ApplySectorConfiguration(organizationId string, sectorDependencies []sectors.Sector, dumpSectorDeps bool, dryRun bool) error

	GetVersionDataInheritanceConfiguration(organizationId string) (versiondata.VersionDataInheritanceConfig, error)

	ApplyVersionDataInheritanceConfiguration(organizationId string, inheritance versiondata.VersionDataInheritanceConfig, dumpConfig bool, dryRun bool) error

	Status(organizationId string, showClustersWithoutPolicy bool) (organization *amv1.Organization, clusterInfos []*clusters.ClusterInfo, blockedVersions []string, sectors []sectors.Sector, inheritance versiondata.VersionDataInheritanceConfig, err error)
}

func NewPolicyBackend(backendType string) (PolicyBackend, error) {
	switch backendType {
	case "ocmlabels", "":
		return ocmlabels.NewOCMLabelsPolicyBackend(), nil
	default:
		return nil, fmt.Errorf("unknown backend type: %s", backendType)
	}
}
