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
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/policy"
	"github.com/app-sre/aus-cli/pkg/utils"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func (f *OCMLabelsPolicyBackend) ListPolicies(organizationId string, showClustersWithoutPolicy bool) (map[string]*policy.ClusterInfo, error) {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return nil, err
	}

	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return nil, err
		}
	}

	return listPoliciesInOrganization(organizationId, showClustersWithoutPolicy, connection)
}

func (f *OCMLabelsPolicyBackend) DeletePolicy(organizationId string, clusterName string, dryRun bool) error {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return err
	}
	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return err
		}
	}

	subscription, err := getSubscriptionForDisplayName(organizationId, clusterName, connection)
	if err != nil {
		return err
	}

	output.Log(dryRun, "Delete cluster upgrade policy from %s\n", clusterName)
	return deleteSubscriptionLabels(subscription.ID(), newAusLabelKey(""), connection, dryRun)
}

func (f *OCMLabelsPolicyBackend) ApplyPolicies(organizationId string, policies []policy.ClusterUpgradePolicy, dumpPolicy bool, dryRun bool) error {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return err
	}
	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return err
		}
	}

	if dumpPolicy {
		body, err := json.Marshal(policies)
		if err != nil {
			return err
		}
		err = output.PrettyList(os.Stdout, body)
		if err != nil {
			return err
		}
		return nil
	}

	for _, policy := range policies {
		err := f.applyPolicy(organizationId, policy, connection, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *OCMLabelsPolicyBackend) applyPolicy(organizationId string, policy policy.ClusterUpgradePolicy, connection *sdk.Connection, dryRun bool) error {
	subscription, err := getSubscriptionForDisplayName(organizationId, policy.ClusterName, connection)
	if err != nil {
		return err
	}

	// get current labels and build a container out of them
	policyLabels, err := listSubscriptionLabels(subscription.ID(), newAusLabelKey(""), connection)
	if err != nil {
		return err
	}
	labelsContainer := NewOCMLabelsContainer(policyLabels)

	// build labels for policy and add them to the container
	desiredLabels, err := newClusterUpgradePolicyFromOCMLabels(policy, subscription.ID())
	if err != nil {
		return err
	}
	labelsContainer.AddLabels(desiredLabels)

	// reconcile
	output.Log(dryRun, "Apply cluster upgrade policy to %s\n", policy.ClusterName)
	return labelsContainer.Reconcile(dryRun, connection)
}

func listPoliciesInOrganization(organizationId string, showClustersWithoutPolicy bool, connection *sdk.Connection) (map[string]*policy.ClusterInfo, error) {
	cluster_map := make(map[string]*policy.ClusterInfo)
	clusterInfos, err := clusterInfos(organizationId, "", connection)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusterInfos {
		if (cluster.Policy != nil && cluster.Policy.Validate() == nil) || showClustersWithoutPolicy {
			cluster_map[cluster.Cluster.Name()] = cluster
		}
	}
	return cluster_map, nil
}

func newClusterUpgradePolicyFromOCMLabels(policy policy.ClusterUpgradePolicy, subscriptionID string) ([]*amv1.Label, error) {
	labels := []*amv1.Label{}
	soakDayLabel, _ := buildOCMLabel(newAusLabelKey("soak-days"), strconv.Itoa(policy.Conditions.SoakDays), subscriptionID, "")
	labels = append(labels, soakDayLabel)
	workloadsLabel, _ := buildOCMLabel(newAusLabelKey("workloads"), utils.StringArrayToCSV(policy.Workloads), subscriptionID, "")
	labels = append(labels, workloadsLabel)
	if policy.Conditions.Sector != "" {
		sectorLabel, _ := buildOCMLabel(newAusLabelKey("sector"), policy.Conditions.Sector, subscriptionID, "")
		labels = append(labels, sectorLabel)
	}
	scheduleLabel, _ := buildOCMLabel(newAusLabelKey("schedule"), policy.Schedule, subscriptionID, "")
	labels = append(labels, scheduleLabel)
	if len(policy.Conditions.Mutexes) > 0 {
		mutexesLabel, _ := buildOCMLabel(newAusLabelKey("mutexes"), utils.StringArrayToCSV(policy.Conditions.Mutexes), subscriptionID, "")
		labels = append(labels, mutexesLabel)
	}
	return labels, nil
}

func getPolicyForSubscription(subscription *amv1.Subscription, cluster *csv1.Cluster) (*policy.ClusterUpgradePolicy, error) {
	labelsMap := newLabelMap(subscription.Labels())
	policy := policy.ClusterUpgradePolicy{
		ClusterName: subscription.DisplayName(),
	}

	scheduleLabel, ok := labelsMap[newAusLabelKey("schedule")]
	if ok {
		policy.Schedule = scheduleLabel.Value()
	}
	workloadsLabel, ok := labelsMap[newAusLabelKey("workloads")]
	if ok {
		policy.Workloads = strings.Split(workloadsLabel.Value(), ",")
	} else {
		policy.Workloads = []string{}
	}
	mutexesLabel, ok := labelsMap[newAusLabelKey("mutexes")]
	if ok {
		policy.Conditions.Mutexes = strings.Split(mutexesLabel.Value(), ",")
	} else {
		policy.Conditions.Mutexes = []string{}
	}
	soakDaysLabel, ok := labelsMap[newAusLabelKey("soak-days")]
	if ok {
		soakDays, err := strconv.Atoi(soakDaysLabel.Value())
		if err != nil {
			return nil, err
		}
		policy.Conditions.SoakDays = soakDays
	}
	sectorLabel, ok := labelsMap[newAusLabelKey("sector")]
	if ok {
		policy.Conditions.Sector = sectorLabel.Value()
	}
	return &policy, nil
}
