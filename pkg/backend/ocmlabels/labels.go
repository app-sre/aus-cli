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

	"github.com/app-sre/aus-cli/pkg/arguments"
	"github.com/app-sre/aus-cli/pkg/debug"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/utils"
	"github.com/openshift-online/ocm-cli/pkg/dump"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

type OCMLabelsContainer struct {
	supportedLabelKeys []string
	currentLabels      map[string]*amv1.Label
	desiredLabels      map[string]*amv1.Label
}

func LabelSupported(label *amv1.Label, supportedLabelKeys []string) bool {
	if supportedLabelKeys == nil {
		return true
	}
	return utils.StringInArray(supportedLabelKeys, label.Key())
}

func NewOCMLabelsContainer(currentLLabels []*amv1.Label) *OCMLabelsContainer {
	return &OCMLabelsContainer{
		supportedLabelKeys: nil,
		currentLabels:      newLabelMap(currentLLabels, nil),
		desiredLabels:      make(map[string]*amv1.Label),
	}
}

func NewRestrictingOCMLabelsContainer(currentLLabels []*amv1.Label, supportedLabelKeys []string) *OCMLabelsContainer {
	return &OCMLabelsContainer{
		supportedLabelKeys: supportedLabelKeys,
		currentLabels:      newLabelMap(currentLLabels, supportedLabelKeys),
		desiredLabels:      make(map[string]*amv1.Label),
	}
}

func (lc *OCMLabelsContainer) LabelSupported(label *amv1.Label) bool {
	return LabelSupported(label, lc.supportedLabelKeys)
}

func (lc *OCMLabelsContainer) AddLabel(label *amv1.Label) bool {
	if lc.LabelSupported(label) {
		lc.desiredLabels[label.Key()] = label
		return true
	}
	return false
}

func (lc *OCMLabelsContainer) AddLabels(label []*amv1.Label) bool {
	for _, l := range label {
		ok := lc.AddLabel(l)
		if !ok {
			return false
		}
	}
	return true
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
	org_labels, err := connection.AccountsMgmt().V1().Organizations().Organization(organizationId).Labels().List().Parameter("search", fmt.Sprintf("key like '%s%%'", keyPrefix)).Send()
	if err == nil {
		return org_labels.Items().Slice(), nil
	}
	labels, err := connection.AccountsMgmt().V1().Labels().List().Parameter("search", fmt.Sprintf("organization_id = '%s' and key like '%s%%'", organizationId, keyPrefix)).Send()
	if err != nil {
		return nil, err
	}
	return labels.Items().Slice(), nil

}

func getOrganizationLabel(organizationId string, key string, connection *sdk.Connection) (*amv1.Label, error) {
	labels, err := listOrganizationLabels(organizationId, key, connection)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return nil, nil
	}
	if len(labels) > 1 {
		return nil, fmt.Errorf("found more than one label with key %s", key)
	}
	return labels[0], nil
}

func listSubscriptionLabels(subscriptionId string, keyPrefix string, connection *sdk.Connection) ([]*amv1.Label, error) {
	labels, err := connection.AccountsMgmt().V1().Subscriptions().Subscription(subscriptionId).Labels().List().Parameter("search", fmt.Sprintf("key like '%s%%'", keyPrefix)).Send()
	if err != nil {
		return nil, err
	}
	return labels.Items().Slice(), nil
}

func deleteSubscriptionLabels(subscriptionId string, keyPrefix string, connection *sdk.Connection, dryRun bool) error {
	labels, err := listSubscriptionLabels(subscriptionId, keyPrefix, connection)
	if err != nil {
		return err
	}
	for _, label := range labels {
		err = deleteOCMLabel(label, dryRun, connection)
		if err != nil {
			return err
		}
	}
	return nil
}

func newLabelMap(labels []*amv1.Label, supportedLabelKeys []string) map[string]*amv1.Label {
	labelMap := make(map[string]*amv1.Label)
	for _, label := range labels {
		if LabelSupported(label, supportedLabelKeys) {
			labelMap[label.Key()] = label
		}
	}
	return labelMap
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
		return fmt.Errorf("label is missing subscription_id or organization_id")
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
