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

package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GetClusterByName(organizationId string, clusterName string, connection *sdk.Connection) (*csv1.Cluster, error) {
	searchQuery := fmt.Sprintf("organization.id = '%s' and name = '%s' and state = 'ready'", organizationId, clusterName)
	clustersResponse, err := connection.ClustersMgmt().V1().Clusters().List().Size(1).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	clusters := clustersResponse.Items().Slice()
	if len(clusters) == 0 {
		return nil, nil
	}
	return clusters[0], nil
}

func ClustersForOrganization(organizationId string, connection *sdk.Connection) (map[string]*csv1.Cluster, error) {
	searchQuery := fmt.Sprintf("organization.id = '%s' and managed = 'true' and state = 'ready'", organizationId)
	clusters, err := connection.ClustersMgmt().V1().Clusters().List().Size(100).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	clusterMap := make(map[string]*csv1.Cluster)
	for _, cluster := range clusters.Items().Slice() {
		clusterMap[cluster.ID()] = cluster
	}
	return clusterMap, nil
}
