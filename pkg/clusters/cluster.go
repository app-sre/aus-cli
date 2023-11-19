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

package clusters

import (
	"sort"

	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/policy"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterInfo struct {
	Subscription          *amv1.Subscription
	Cluster               *csv1.Cluster
	VersionGateAgreements *map[string]*csv1.VersionGateAgreement
	Policy                *policy.ClusterUpgradePolicy
}

func (c *ClusterInfo) STSEnabled() bool {
	aws, ok := c.Cluster.GetAWS()
	if !ok {
		return false
	}
	sts, ok := aws.GetSTS()
	if !ok {
		return false
	}
	return sts.Enabled()
}

func SortClusters(clusters []*ClusterInfo) {
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Cluster.Name() < clusters[j].Cluster.Name()
	})
}

func GetClusterInfoByName(organizationID string, clusterName string, connection *sdk.Connection) (*ClusterInfo, error) {
	cluster, err := ocm.GetClusterByName(organizationID, clusterName, connection)
	if err != nil {
		return nil, err
	}
	agreements, err := ocm.GetVersionGateAgreements(cluster.ID(), connection)
	if err != nil {
		return nil, err
	}
	return &ClusterInfo{
		Cluster:               cluster,
		VersionGateAgreements: &agreements,
	}, nil
}

func ClusterInfosForOrganization(organizationId string, subscriptionSearchQuery string, withAgreements bool, connection *sdk.Connection) ([]*ClusterInfo, error) {
	subscriptions, err := ocm.SubscriptionsForOrganization(organizationId, subscriptionSearchQuery, connection)
	if err != nil {
		return nil, err
	}
	clusterMap, err := ocm.ClustersForOrganization(organizationId, connection)
	if err != nil {
		return nil, err
	}

	clusterInfos := make([]*ClusterInfo, 0)
	for _, subscription := range subscriptions {
		clusterInfo := clusterMap[subscription.ClusterID()]
		if clusterInfo == nil {
			continue
		}
		var agreements *map[string]*csv1.VersionGateAgreement
		if withAgreements {
			agreements_map, err := ocm.GetVersionGateAgreements(subscription.ClusterID(), connection)
			if err != nil {
				return nil, err
			}
			agreements = &agreements_map
		}
		clusterInfos = append(
			clusterInfos,
			&ClusterInfo{
				Subscription:          subscription,
				Cluster:               clusterInfo,
				VersionGateAgreements: agreements,
			},
		)
	}
	return clusterInfos, nil
}
