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
	"github.com/app-sre/aus-cli/pkg/output"
	sdk "github.com/openshift-online/ocm-sdk-go"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GetVersionGates(connection *sdk.Connection) (map[string][]*csv1.VersionGate, error) {
	gates, err := connection.ClustersMgmt().V1().VersionGates().List().Send()
	if err != nil {
		return nil, err
	}
	gates_map := make(map[string][]*csv1.VersionGate)
	for _, gate := range gates.Items().Slice() {
		prefix := gate.VersionRawIDPrefix()
		if _, ok := gates_map[prefix]; !ok {
			gates_map[prefix] = []*csv1.VersionGate{}
		}
		gates_map[prefix] = append(gates_map[prefix], gate)
	}
	return gates_map, nil
}

func GetVersionGateAgreements(clusterId string, connection *sdk.Connection) (map[string]*csv1.VersionGateAgreement, error) {
	agreements, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterId).GateAgreements().List().Send()
	if err != nil {
		return nil, err
	}
	agreements_map := make(map[string]*csv1.VersionGateAgreement)
	for _, agreement := range agreements.Items().Slice() {
		agreements_map[agreement.ID()] = agreement
	}
	return agreements_map, nil
}

func AckVersionGate(cluster *csv1.Cluster, gate *csv1.VersionGate, connection *sdk.Connection, dryRun bool) (*csv1.VersionGateAgreement, error) {
	agreement, err := csv1.NewVersionGateAgreement().
		VersionGate(csv1.NewVersionGate().Copy(gate)).
		Build()
	if err != nil {
		return nil, err
	}

	output.Log(dryRun, "Apply agreement for gate %s to cluster %s\n", gate.ID(), cluster.Name())
	if !dryRun {
		response, err := connection.ClustersMgmt().V1().
			Clusters().Cluster(cluster.ID()).
			GateAgreements().
			Add().
			Body(agreement).
			Send()
		if err != nil {
			return nil, err
		}

		return response.Body(), nil
	} else {
		return agreement, nil
	}
}
