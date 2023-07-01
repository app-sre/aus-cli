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

	"github.com/app-sre/aus-cli/pkg/policy"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func newAusLabelKey(suffix string) string {
	return fmt.Sprintf("sre-capabilities.aus.%s", suffix)
}

func getSubscriptionForDisplayName(organizationId string, displayName string, connection *sdk.Connection) (*amv1.Subscription, error) {
	searchQuery := fmt.Sprintf("organization_id = '%s' and managed = true and status in ('Active', 'Reserved') and display_name = '%s'", organizationId, displayName)
	subscriptions, err := listSubscriptions(organizationId, searchQuery, connection)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	if len(subscriptions) > 1 {
		return nil, fmt.Errorf("more than one subscription found for display name '%s'", displayName)
	}
	return subscriptions[0], nil
}

func listSubscriptions(organizationId string, searchQuery string, connection *sdk.Connection) ([]*amv1.Subscription, error) {
	if searchQuery == "" {
		searchQuery = fmt.Sprintf("organization_id = '%s' and managed = true and status in ('Active', 'Reserved')", organizationId)
	} else {
		searchQuery = fmt.Sprintf("organization_id = '%s' and %s", organizationId, searchQuery)
	}
	subscriptions, err := connection.AccountsMgmt().V1().Subscriptions().List().Parameter("fetchLabels", "true").Size(100).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	return subscriptions.Items().Slice(), nil
}

func listClusters(organizationId string, connection *sdk.Connection) ([]*csv1.Cluster, error) {
	searchQuery := fmt.Sprintf("organization.id = '%s' and managed = 'true'", organizationId)
	clusters, err := connection.ClustersMgmt().V1().Clusters().List().Size(100).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	return clusters.Items().Slice(), nil
}

func clusterInfos(organizationId string, subscriptionSearchQuery string, connection *sdk.Connection) ([]*policy.ClusterInfo, error) {
	subscriptions, err := listSubscriptions(organizationId, subscriptionSearchQuery, connection)
	if err != nil {
		return nil, err
	}
	clusters, err := listClusters(organizationId, connection)
	if err != nil {
		return nil, err
	}
	clusterMap := make(map[string]*csv1.Cluster)
	for _, cluster := range clusters {
		clusterMap[cluster.ID()] = cluster
	}

	clusterInfos := make([]*policy.ClusterInfo, 0)
	for _, subscription := range subscriptions {
		cluster := clusterMap[subscription.ClusterID()]
		if cluster == nil {
			continue
		}
		pol, err := getPolicyForSubscription(subscription, cluster)
		if err != nil {
			fmt.Println("some error")
		}
		clusterInfos = append(
			clusterInfos,
			&policy.ClusterInfo{
				Subscription: subscription,
				Cluster:      cluster,
				Policy:       pol,
			},
		)
	}
	return clusterInfos, nil
}
