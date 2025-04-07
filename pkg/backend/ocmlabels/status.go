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
	"github.com/app-sre/aus-cli/pkg/clusters"
	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/sectors"
	"github.com/app-sre/aus-cli/pkg/versiondata"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func (f *OCMLabelsPolicyBackend) Status(organizationId string, showClustersWithoutPolicy bool) (organization *amv1.Organization, clusterInfos []*clusters.ClusterInfo, blockedVersions []string, sectors []sectors.Sector, inheritance versiondata.VersionDataInheritanceConfig, err error) {
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
	clusterInfos = []*clusters.ClusterInfo{}
	for _, c := range clustersMap {
		clusterInfos = append(clusterInfos, c)
	}
	clusters.SortClusters(clusterInfos)

	sectors, err = listSectorsFromOrganizationLabels(organization.ID(), connection)
	if err != nil {
		return
	}
	inheritance, err = listVersionDataInheritanceConfiguration(organization.ID(), connection)
	return
}
