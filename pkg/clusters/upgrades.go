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
	"fmt"
	"regexp"

	semver "github.com/Masterminds/semver/v3"
	"github.com/app-sre/aus-cli/pkg/versions"
	csv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func (c *ClusterInfo) blockedVersionExpressions() ([]*regexp.Regexp, error) {
	if c.Policy == nil {
		return nil, fmt.Errorf("cluster %s has no policy", c.Cluster.ID())
	}
	return versions.ParsedBlockedVersionExpressions(c.Policy.Conditions.BlockedVersions)
}

func (c *ClusterInfo) AvailableUpgrades(considerBlockedVersions bool, additionalBlockedVersions []*regexp.Regexp) []string {
	if considerBlockedVersions {
		return c.Cluster.Version().AvailableUpgrades()
	}

	var upgrades []string
	clusterBlockedVersionExpressions, err := c.blockedVersionExpressions()
	if err != nil {
		return nil
	}
	for _, version := range c.Cluster.Version().AvailableUpgrades() {
		// check if the version is blocked on the cluster
		if versions.IsVersionBlocked(version, clusterBlockedVersionExpressions) {
			continue
		}
		// check if the version is blocked by other blockers (e.g. org level blockers)
		if versions.IsVersionBlocked(version, additionalBlockedVersions) {
			continue
		}
		upgrades = append(upgrades, version)
	}
	return upgrades
}

func (c *ClusterInfo) YStreamUpgrades(considerBlockedVersions bool, additionalBlockedVersions []*regexp.Regexp) []string {
	var ystreamUpgrades = make(map[string]struct{})
	currentVersion, _ := semver.NewVersion(c.Cluster.Version().RawID())
	yStreamUpgradeCondition, _ := semver.NewConstraint(fmt.Sprintf(">= %d.%d", currentVersion.Major(), currentVersion.Minor()+1))
	var upgrades = c.AvailableUpgrades(considerBlockedVersions, additionalBlockedVersions)
	for _, version := range upgrades {
		uv, _ := semver.NewVersion(version)
		if yStreamUpgradeCondition.Check(uv) {
			ystreamUpgrades[fmt.Sprintf("%d.%d", uv.Major(), uv.Minor())] = struct{}{}
		}
	}
	var keys []string
	for k := range ystreamUpgrades {
		keys = append(keys, k)
	}

	return keys
}

func (c *ClusterInfo) MissingGateAgreements(additionalBlockedVersions []*regexp.Regexp, gates map[string][]*csv1.VersionGate) ([]*csv1.VersionGate, error) {
	var missingGates []*csv1.VersionGate
	for _, yStreamUpgrade := range c.YStreamUpgrades(true, additionalBlockedVersions) {
		yStreamGates, ok := gates[yStreamUpgrade]
		if !ok {
			continue
		}
		for _, yStreamGate := range yStreamGates {
			if yStreamGate.STSOnly() && !c.STSEnabled() {
				continue
			}

			var missing = true
			if c.VersionGateAgreements != nil {
				for _, agreement := range *c.VersionGateAgreements {
					if agreement.VersionGate().ID() == yStreamGate.ID() {
						missing = false
						break
					}
				}
			}
			if missing {
				missingGates = append(missingGates, yStreamGate)
			}
		}
	}
	return missingGates, nil
}

func (c *ClusterInfo) YStreamsWithMissingSTSGateAgreements(additionalBlockedVersions []*regexp.Regexp, gates map[string][]*csv1.VersionGate) ([]string, error) {
	missingGates, err := c.MissingGateAgreements(additionalBlockedVersions, gates)
	if err != nil {
		return nil, err
	}

	ystreams := make([]string, 0)
	for _, missingGate := range missingGates {
		yStream, _ := semver.NewVersion(missingGate.VersionRawIDPrefix())
		ystreams = append(ystreams, fmt.Sprintf("%d.%d", yStream.Major(), yStream.Minor()))
	}
	return ystreams, nil
}
