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

type Sector struct {
	Name                string   `json:"name"`
	Dependencies        []string `json:"dependencies,omitempty"`
	MaxParallelUpgrades string   `json:"maxParallelUpgrades,omitempty"`
}

func (s *Sector) DependsOn(sector string) bool {
	for _, dependency := range s.Dependencies {
		if dependency == sector {
			return true
		}
	}
	return false
}

func (s *Sector) AddDependencies(dependencies []string) {
	for _, sector := range dependencies {
		if !s.DependsOn(sector) {
			s.Dependencies = append(s.Dependencies, sector)
		}
	}
}

func (s *Sector) DeleteDependency(sector string) {
	for i, dependency := range s.Dependencies {
		if dependency == sector {
			s.Dependencies = append(s.Dependencies[:i], s.Dependencies[i+1:]...)
			return
		}
	}
}

func newSectorDependencies(sectorRepr string) (Sector, error) {
	split := strings.Split(sectorRepr, "=")
	var dependencies = []string{}
	if len(split) == 2 {
		dependencies = strings.Split(split[1], ",")
	}
	return Sector{
		Name:         split[0],
		Dependencies: dependencies,
	}, nil
}

func NewSectorDependenciesList(sectorListRepr []string) ([]Sector, error) {
	sectorDependencies := []Sector{}
	for _, sectorRepr := range sectorListRepr {
		s, err := newSectorDependencies(sectorRepr)
		if err != nil {
			return nil, err
		}
		sectorDependencies = append(sectorDependencies, s)
	}
	sortSectors(sectorDependencies)
	return sectorDependencies, nil
}

func newSectorMaxParallelUpgrades(sectorRepr string) (Sector, error) {
	split := strings.Split(sectorRepr, "=")
	var maxParallelUpgrades string
	if len(split) == 2 {
		maxParallelUpgrades = split[1]
	}
	return Sector{
		Name:                split[0],
		MaxParallelUpgrades: maxParallelUpgrades,
	}, nil
}

func NewSectorMaxParallelUpgradesList(sectorListRepr []string) ([]Sector, error) {
	sectorList := []Sector{}
	for _, sectorRepr := range sectorListRepr {
		s, err := newSectorMaxParallelUpgrades(sectorRepr)
		if err != nil {
			return nil, err
		}
		sectorList = append(sectorList, s)
	}
	sortSectors(sectorList)
	return sectorList, nil
}

func AddMissingSectors(sectors []Sector) []Sector {
	sectorMap := make(map[string]Sector)
	for _, sector := range sectors {
		sectorMap[sector.Name] = sector
	}

	for _, sector := range sectors {
		for _, sectorName := range sector.Dependencies {
			if _, ok := sectorMap[sectorName]; !ok {
				sectorMap[sectorName] = Sector{
					Name:         sectorName,
					Dependencies: []string{},
				}
			}
		}
	}

	sectorDependencies := []Sector{}
	for _, sector := range sectorMap {
		sectorDependencies = append(sectorDependencies, sector)
	}
	sortSectors(sectorDependencies)
	return sectorDependencies
}

func sortSectors(sectors []Sector) {
	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].Name < sectors[j].Name
	})
}

func ConsolidateSectorList(current []Sector, toAdd []Sector, toDelete []Sector, maxParallelUpgrades []Sector) []Sector {
	sectorMap := make(map[string]*Sector)
	for i := range current {
		sectorMap[current[i].Name] = &current[i]
	}

	// maxParallelUpgrades
	for i := range maxParallelUpgrades {
		sector := maxParallelUpgrades[i]
		if currentSector, ok := sectorMap[sector.Name]; ok {
			currentSector.MaxParallelUpgrades = sector.MaxParallelUpgrades
		} else {
			sectorMap[sector.Name] = &sector
		}
	}

	// adding
	for i := range toAdd {
		sector := toAdd[i]
		if currentSector, ok := sectorMap[sector.Name]; ok {
			currentSector.AddDependencies(sector.Dependencies)
			if sector.MaxParallelUpgrades != "" {
				currentSector.MaxParallelUpgrades = sector.MaxParallelUpgrades
			}
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
		if len(sectorMap[sectorName].Dependencies) == 0 && sectorMap[sectorName].MaxParallelUpgrades == "" {
			delete(sectorMap, sectorName)
		}
	}

	sectorList := []Sector{}
	for _, sector := range sectorMap {
		sectorList = append(sectorList, *sector)
	}
	sortSectors(sectorList)
	return sectorList
}

func ReadSectorsFromReader(reader io.Reader) ([]Sector, error) {
	var sectorList []Sector
	err := json.NewDecoder(reader).Decode(&sectorList)
	return sectorList, err
}
