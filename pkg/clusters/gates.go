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
	"github.com/app-sre/aus-cli/pkg/ocm"
	sdk "github.com/openshift-online/ocm-sdk-go"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func AckAllGatesForYStream(cluster *ClusterInfo, yStream string, connection *sdk.Connection, dryRun bool) ([]*csv1.VersionGateAgreement, error) {
	agreements := []*csv1.VersionGateAgreement{}
	gates, err := ocm.GetVersionGates(connection)
	if err != nil {
		return nil, err
	}
	unackedGates, err := cluster.MissingGateAgreements(nil, gates)
	if err != nil {
		return nil, err
	}
	for _, gate := range unackedGates {
		if gate.VersionRawIDPrefix() == yStream {
			agreement, err := ocm.AckVersionGate(cluster.Cluster, gate, connection, dryRun)
			if err != nil {
				return nil, err
			}
			agreements = append(agreements, agreement)
		}
	}
	return agreements, nil
}
