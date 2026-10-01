// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"errors"
	"testing"
)

type fakeClusterReadinessLifecycle struct {
	created   bool
	waited    bool
	clusterID string
	createErr error
	waitErr   error
}

func (f *fakeClusterReadinessLifecycle) Create() (string, error) {
	f.created = true
	return f.clusterID, f.createErr
}

func (f *fakeClusterReadinessLifecycle) WaitReady(_ context.Context, _ string) error {
	f.waited = true
	return f.waitErr
}

func TestCreateAndWaitReady(t *testing.T) {
	fake := &fakeClusterReadinessLifecycle{clusterID: "cluster-123"}
	got, err := CreateAndWaitReady(context.Background(), fake)
	if err != nil {
		t.Fatal(err)
	}
	if got != "cluster-123" || !fake.created || !fake.waited {
		t.Fatalf("workflow result=%q created=%t waited=%t", got, fake.created, fake.waited)
	}
}

func TestCreateAndWaitReadyStopsOnCreateError(t *testing.T) {
	wantErr := errors.New("create failed")
	fake := &fakeClusterReadinessLifecycle{createErr: wantErr}
	if _, err := CreateAndWaitReady(context.Background(), fake); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if fake.waited {
		t.Fatal("WaitReady should not run after create failure")
	}
}

func TestCreateAndWaitReadyStopsOnEmptyID(t *testing.T) {
	fake := &fakeClusterReadinessLifecycle{}
	if _, err := CreateAndWaitReady(context.Background(), fake); err == nil {
		t.Fatal("expected an empty cluster ID error")
	}
	if fake.waited {
		t.Fatal("WaitReady should not run with an empty cluster ID")
	}
}

func TestCreateAndWaitReadyReturnsIDOnReadinessError(t *testing.T) {
	wantErr := errors.New("not ready")
	fake := &fakeClusterReadinessLifecycle{clusterID: "cluster-123", waitErr: wantErr}
	got, err := CreateAndWaitReady(context.Background(), fake)
	if got != "cluster-123" || !errors.Is(err, wantErr) {
		t.Fatalf("result=(%q, %v), want cluster ID and %v", got, err, wantErr)
	}
}
