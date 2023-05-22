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
	"bytes"
	"fmt"
	"os"

	"github.com/openshift-online/ocm-cli/pkg/dump"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/arguments"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/debug"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/output"
)

type OCMLabelsContainer struct {
	currentLabels map[string]*amv1.Label
	desiredLabels map[string]*amv1.Label
}

func NewOCMLabelsContainer(currentLLabels []*amv1.Label) *OCMLabelsContainer {
	return &OCMLabelsContainer{
		currentLabels: newLabelMap(currentLLabels),
		desiredLabels: make(map[string]*amv1.Label),
	}
}

func (lc *OCMLabelsContainer) AddLabel(label *amv1.Label) {
	lc.desiredLabels[label.Key()] = label
}

func (lc *OCMLabelsContainer) AddLabels(label []*amv1.Label) {
	for _, l := range label {
		lc.AddLabel(l)
	}
}

func (lc *OCMLabelsContainer) Reconcile(dryRun bool, connection *sdk.Connection) error {
	currentLabelsCopy := make(map[string]*amv1.Label)
	for k, v := range lc.currentLabels {
		currentLabelsCopy[k] = v
	}

	// apply labels
	for _, label := range lc.desiredLabels {
		delete(currentLabelsCopy, label.Key())
		err := applyOCMLabel(label, dryRun, connection)
		if err != nil {
			return err
		}
	}

	// remove obsolete labels
	for _, label := range currentLabelsCopy {
		err := deleteOCMLabel(label, dryRun, connection)
		// maybe ignore 404
		if err != nil {
			return err
		}
	}

	return nil
}

func listOrganizationLabels(organizationId string, keyPrefix string, connection *sdk.Connection) ([]*amv1.Label, error) {
	labels, err := connection.AccountsMgmt().V1().Organizations().Organization(organizationId).Labels().List().Parameter("search", fmt.Sprintf("key like '%s%%'", keyPrefix)).Send()
	if err != nil {
		return nil, err
	}
	return labels.Items().Slice(), nil
}

func getOrganizationLabel(organizationId string, key string, connection *sdk.Connection) (*amv1.Label, error) {
	response, err := connection.AccountsMgmt().V1().Organizations().Organization(organizationId).Labels().Label(key).Get().Send()
	if response.Status() == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	body, ok := response.GetBody()
	if !ok {
		return nil, nil
	}
	return body, nil
}

func listSubscriptionLabels(subscriptionId string, keyPrefix string, connection *sdk.Connection) ([]*amv1.Label, error) {
	labels, err := connection.AccountsMgmt().V1().Subscriptions().Subscription(subscriptionId).Labels().List().Parameter("search", fmt.Sprintf("key like '%s%%'", keyPrefix)).Send()
	if err != nil {
		return nil, err
	}
	return labels.Items().Slice(), nil
}

func newLabelMap(labels []*amv1.Label) map[string]*amv1.Label {
	labelMap := make(map[string]*amv1.Label)
	for _, label := range labels {
		labelMap[label.Key()] = label
	}
	return labelMap
}

func listSubscriptions(organizationId string, displayName string, connection *sdk.Connection) ([]*amv1.Subscription, error) {
	searchQuery := ""
	if displayName != "" {
		searchQuery = fmt.Sprintf("organization_id = '%s' and managed = true and status in ('Active', 'Reserved') and display_name = '%s'", organizationId, displayName)
	} else {
		searchQuery = fmt.Sprintf("organization_id = '%s' and managed = true and status in ('Active', 'Reserved')", organizationId)
	}
	subscriptions, err := connection.AccountsMgmt().V1().Subscriptions().List().Parameter("fetchLabels", "true").Size(100).Search(searchQuery).Send()
	if err != nil {
		return nil, err
	}
	return subscriptions.Items().Slice(), nil
}

func buildOCMLabel(key string, value string, subscriptionId string, organizationId string) (*amv1.Label, error) {
	label := amv1.NewLabel().Key(key).Value(value)
	if subscriptionId != "" {
		label = label.SubscriptionID(subscriptionId)
	}
	if organizationId != "" {
		label = label.OrganizationID(organizationId)
	}
	return label.Build()
}

func applyOCMLabel(label *amv1.Label, dryRun bool, connection *sdk.Connection) error {
	var request *amv1.GenericLabelsAddRequest
	if label.SubscriptionID() != "" {
		request = connection.AccountsMgmt().V1().Subscriptions().Subscription(label.SubscriptionID()).Labels().Add().Body(label)
		output.Debug(dryRun, "Add label %s to subscription %s\n", label.Key(), label.SubscriptionID())
	} else if label.OrganizationID() != "" {
		request = connection.AccountsMgmt().V1().Organizations().Organization(label.OrganizationID()).Labels().Add().Body(label)
		output.Debug(dryRun, "Add label %s to organization %s\n", label.Key(), label.OrganizationID())
	} else {
		return fmt.Errorf("Label is missing subscription_id or organization_id")
	}
	if debug.Enabled() {
		buf := new(bytes.Buffer)
		err := amv1.MarshalLabel(label, buf)
		if err != nil {
			return err
		}
		err = dump.Pretty(os.Stdout, buf.Bytes())
		if err != nil {
			return err
		}
	}
	if !dryRun {
		_, err := request.Send()
		return err
	}
	return nil
}

func deleteOCMLabel(label *amv1.Label, dryRun bool, connection *sdk.Connection) error {
	request := connection.Delete()
	err := arguments.ApplyPathArg(request, label.HREF())
	if err != nil {
		return err
	}
	output.Debug(dryRun, "Delete label %s\n", label.HREF())
	if !dryRun {
		_, err := request.Send()
		return err
	}
	return nil
}
