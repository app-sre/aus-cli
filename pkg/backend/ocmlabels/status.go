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

package ocmlabels

import (
	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/policy"
	"github.com/app-sre/aus-cli/pkg/sectors"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func (f *OCMLabelsPolicyBackend) Status(organizationId string, showClustersWithoutPolicy bool) (organization *amv1.Organization, clusters []policy.ClusterInfo, blockedVersions []string, sectors []sectors.SectorDependencies, err error) {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return
	}
	organization, err = ocm.GetOrganization(organizationId, connection)
	if err != nil {
		return
	}

	blockedVersions, err = getBlockedVersionsForOrganization(organization.ID(), connection)
	if err != nil {
		return
	}
	clustersMap, err := listPoliciesInOrganization(organization.ID(), showClustersWithoutPolicy, connection)
	if err != nil {
		return
	}
	clusters = []policy.ClusterInfo{}
	for _, c := range clustersMap {
		clusters = append(clusters, *c)
	}
	policy.SortClusters(clusters)
	sectors, err = listSectorConfigurationFromOrganizationLabels(organization.ID(), connection)
	return
}
