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
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func SubscriptionsForOrganization(organizationId string, searchQuery string, connection *sdk.Connection) (map[string]*amv1.Subscription, error) {
	if searchQuery == "" {
		searchQuery = fmt.Sprintf("organization_id = '%s' and managed = true and status in ('Active', 'Reserved')", organizationId)
	} else {
		searchQuery = fmt.Sprintf("organization_id = '%s' and %s", organizationId, searchQuery)
	}
	subscriptions, err := connection.AccountsMgmt().V1().Subscriptions().List().Parameter("fetchLabels", "true").Size(100).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	subscriptionMap := make(map[string]*amv1.Subscription)
	for _, subscription := range subscriptions.Items().Slice() {
		subscriptionMap[subscription.ID()] = subscription
	}
	return subscriptionMap, nil
}

func SubscriptionForDisplayName(organizationId string, displayName string, connection *sdk.Connection) (*amv1.Subscription, error) {
	searchQuery := fmt.Sprintf("managed = true and status in ('Active', 'Reserved') and display_name = '%s'", displayName)
	subscriptions, err := SubscriptionsForOrganization(organizationId, searchQuery, connection)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) > 1 {
		return nil, fmt.Errorf("more than one subscription found for display name '%s' in organization '%s'", displayName, organizationId)
	}
	for _, subscription := range subscriptions {
		return subscription, nil
	}
	return nil, fmt.Errorf("no subscription found for display name '%s' in organization '%s'", displayName, organizationId)
}
