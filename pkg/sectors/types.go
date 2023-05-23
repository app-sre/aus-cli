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

package sectors

import (
	"encoding/json"
	"io"
	"sort"
	"strings"
)

type SectorDepsUpdateMode int64

const (
	Modify SectorDepsUpdateMode = iota
	Replace
)

type SectorDependencies struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies,omitempty"`
}

func (s *SectorDependencies) DependsOn(sector string) bool {
	for _, dependency := range s.Dependencies {
		if dependency == sector {
			return true
		}
	}
	return false
}

func (s *SectorDependencies) AddDependencies(dependencies []string) {
	for _, sector := range dependencies {
		if !s.DependsOn(sector) {
			s.Dependencies = append(s.Dependencies, sector)
		}
	}
}

func (s *SectorDependencies) DeleteDependency(sector string) {
	for i, dependency := range s.Dependencies {
		if dependency == sector {
			s.Dependencies = append(s.Dependencies[:i], s.Dependencies[i+1:]...)
			return
		}
	}
}

func NewSectorDependencies(sectorRepr string) (SectorDependencies, error) {
	split := strings.Split(sectorRepr, "=")
	var dependencies []string = []string{}
	if len(split) == 2 {
		dependencies = strings.Split(split[1], ",")
	}
	return SectorDependencies{
		Name:         split[0],
		Dependencies: dependencies,
	}, nil
}

func NewSectorDependenciesList(sectorListRepr []string) ([]SectorDependencies, error) {
	sectorDependencies := []SectorDependencies{}
	for _, sectorRepr := range sectorListRepr {
		s, err := NewSectorDependencies(sectorRepr)
		if err != nil {
			return nil, err
		}
		sectorDependencies = append(sectorDependencies, s)
	}
	sortSectors(sectorDependencies)
	return sectorDependencies, nil
}

func AddMissingSectors(sectors []SectorDependencies) []SectorDependencies {
	sectorMap := make(map[string]SectorDependencies)
	for _, sector := range sectors {
		sectorMap[sector.Name] = sector
	}

	for _, sector := range sectors {
		for _, sectorName := range sector.Dependencies {
			if _, ok := sectorMap[sectorName]; !ok {
				sectorMap[sectorName] = SectorDependencies{
					Name:         sectorName,
					Dependencies: []string{},
				}
			}
		}
	}

	sectorDependencies := []SectorDependencies{}
	for _, sector := range sectorMap {
		sectorDependencies = append(sectorDependencies, sector)
	}
	sortSectors(sectorDependencies)
	return sectorDependencies
}

func sortSectors(sectors []SectorDependencies) {
	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].Name < sectors[j].Name
	})
}

func ConsolidateSectorDependencies(current []SectorDependencies, toAdd []SectorDependencies, toDelete []SectorDependencies) []SectorDependencies {
	sectorMap := make(map[string]*SectorDependencies)
	for i := range current {
		sectorMap[current[i].Name] = &current[i]
	}

	// adding
	for i := range toAdd {
		sector := toAdd[i]
		if currentSector, ok := sectorMap[sector.Name]; ok {
			currentSector.AddDependencies(sector.Dependencies)
		} else {
			sectorMap[sector.Name] = &sector
		}
	}

	// removing
	for i := range toDelete {
		sector := toDelete[i]
		if currentSector, ok := sectorMap[sector.Name]; ok {
			for _, dependency := range sector.Dependencies {
				currentSector.DeleteDependency(dependency)
			}
		}
	}

	// cleanup obsolete sectors
	for sectorName := range sectorMap {
		if sectorMap[sectorName].Dependencies == nil || len(sectorMap[sectorName].Dependencies) == 0 {
			delete(sectorMap, sectorName)
		}
	}

	sectorDependencies := []SectorDependencies{}
	for _, sector := range sectorMap {
		sectorDependencies = append(sectorDependencies, *sector)
	}
	sortSectors(sectorDependencies)
	return sectorDependencies
}

func ReadSectorDependenciesFromReader(reader io.Reader) ([]SectorDependencies, error) {
	var deps []SectorDependencies
	err := json.NewDecoder(reader).Decode(&deps)
	return deps, err
}
