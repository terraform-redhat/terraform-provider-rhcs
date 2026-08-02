/*
Copyright (c) 2021 Red Hat, Inc.

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

package machine_types

import (
	"testing"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func buildMachineType(id, name, cloudProvider string, cpuValue float64, cpuUnit string, ramValue float64, ramUnit string) *cmv1.MachineType {
	obj, _ := cmv1.NewMachineType().
		ID(id).
		Name(name).
		CloudProvider(cmv1.NewCloudProvider().ID(cloudProvider)).
		CPU(cmv1.NewValue().Value(cpuValue).Unit(cpuUnit)).
		Memory(cmv1.NewValue().Value(ramValue).Unit(ramUnit)).
		Build()
	return obj
}

func TestMachineTypeState_CPUUnits(t *testing.T) {
	ds := &MachineTypesDataSource{}

	t.Run("vCPU is accepted", func(t *testing.T) {
		mt := buildMachineType("t3.medium", "T3 Medium", "aws", 4, "vCPU", 1, "B")
		state, err := ds.machineTypeState(mt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.CPU != 4 {
			t.Errorf("expected CPU 4, got %d", state.CPU)
		}
	})

	t.Run("unknown CPU unit returns error", func(t *testing.T) {
		mt := buildMachineType("t3.medium", "T3 Medium", "aws", 4, "mCPU", 1, "B")
		_, err := ds.machineTypeState(mt)
		if err == nil {
			t.Fatal("expected error for unknown CPU unit, got nil")
		}
	})
}

func TestMachineTypeState_RAMUnits(t *testing.T) {
	ds := &MachineTypesDataSource{}

	cases := []struct {
		unit    string
		input   float64
		wantRAM int64
	}{
		{"B", 1024, 1024},
		{"KB", 1, 1_000},
		{"MB", 1, 1_000_000},
		{"GB", 1, 1_000_000_000},
		{"TB", 1, 1_000_000_000_000},
		{"PB", 1, 1_000_000_000_000_000},
		{"KiB", 1, 1024},
		{"MiB", 1, 1024 * 1024},
		{"GiB", 1, 1024 * 1024 * 1024},
		{"TiB", 1, 1024 * 1024 * 1024 * 1024},
		{"PiB", 1, 1024 * 1024 * 1024 * 1024 * 1024},
	}

	for _, tc := range cases {
		t.Run(tc.unit, func(t *testing.T) {
			mt := buildMachineType("x", "x", "aws", 4, "vCPU", tc.input, tc.unit)
			state, err := ds.machineTypeState(mt)
			if err != nil {
				t.Fatalf("unexpected error for unit %q: %v", tc.unit, err)
			}
			if state.RAM != tc.wantRAM {
				t.Errorf("unit %q: expected RAM %d, got %d", tc.unit, tc.wantRAM, state.RAM)
			}
		})
	}

	t.Run("unknown RAM unit returns error", func(t *testing.T) {
		mt := buildMachineType("x", "x", "aws", 4, "vCPU", 1, "XB")
		_, err := ds.machineTypeState(mt)
		if err == nil {
			t.Fatal("expected error for unknown RAM unit, got nil")
		}
	})
}

func TestMachineTypeState_Fields(t *testing.T) {
	ds := &MachineTypesDataSource{}
	mt := buildMachineType("m5.xlarge", "M5 XLarge", "aws", 8, "vCPU", 16*1024*1024*1024, "B")
	state, err := ds.machineTypeState(mt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.ID != "m5.xlarge" {
		t.Errorf("expected ID m5.xlarge, got %q", state.ID)
	}
	if state.Name != "M5 XLarge" {
		t.Errorf("expected Name 'M5 XLarge', got %q", state.Name)
	}
	if state.CloudProvider != "aws" {
		t.Errorf("expected CloudProvider aws, got %q", state.CloudProvider)
	}
	if state.CPU != 8 {
		t.Errorf("expected CPU 8, got %d", state.CPU)
	}
	if state.RAM != 16*1024*1024*1024 {
		t.Errorf("expected RAM %d, got %d", 16*1024*1024*1024, state.RAM)
	}
}
