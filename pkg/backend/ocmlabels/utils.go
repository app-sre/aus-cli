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
	"fmt"

	"github.com/app-sre/aus-cli/pkg/clusters"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

func newAusLabelKey(suffix string) string {
	return fmt.Sprintf("sre-capabilities.aus.%s", suffix)
}

func getClusterInfos(organizationId string, subscriptionSearchQuery string, connection *sdk.Connection) ([]*clusters.ClusterInfo, error) {
	clusterInfos, err := clusters.ClusterInfosForOrganization(organizationId, subscriptionSearchQuery, false, connection)
	if err != nil {
		return nil, err
	}
	for _, clusterInfo := range clusterInfos {
		policy, err := getPolicyForSubscription(clusterInfo.Subscription, clusterInfo.Cluster)
		if err != nil {
			return nil, err
		}
		clusterInfo.Policy = policy
	}
	return clusterInfos, nil
}
