// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"os"
	"path"
	"testing"
)

func TestCleanupWorkspace_RemovesStateAndVarsDirs(t *testing.T) {
	manifestsDir := t.TempDir()
	const ws = "hf-e2e-1786714361"

	stateDir := path.Join(manifestsDir, "terraform.tfstate.d", ws)
	varsDir := path.Join(manifestsDir, "terraform.tfvars.d", ws)
	for _, d := range []string{stateDir, varsDir} {
		if err := os.MkdirAll(d, 0o777); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := os.WriteFile(path.Join(stateDir, "terraform.tfstate"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(path.Join(varsDir, "terraform.tfvars"), []byte(""), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	ctx := &terraformExecutorContext{manifestsDir: manifestsDir, tfWorkspace: ws}
	if err := ctx.cleanupWorkspace(); err != nil {
		t.Fatalf("cleanupWorkspace() unexpected error: %v", err)
	}

	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Errorf("state dir should be removed, stat err = %v", err)
	}
	if _, err := os.Stat(varsDir); !os.IsNotExist(err) {
		t.Errorf("vars dir should be removed, stat err = %v", err)
	}
}

func TestCleanupWorkspace_DefaultWorkspaceIsNoOp(t *testing.T) {
	manifestsDir := t.TempDir()
	// The default (empty) workspace stores state directly in the manifest dir;
	// cleanup must not touch it.
	stateFile := path.Join(manifestsDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte("{}"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	ctx := &terraformExecutorContext{manifestsDir: manifestsDir, tfWorkspace: ""}
	if err := ctx.cleanupWorkspace(); err != nil {
		t.Fatalf("cleanupWorkspace() unexpected error: %v", err)
	}

	if _, err := os.Stat(stateFile); err != nil {
		t.Errorf("default workspace state file must be left untouched, stat err = %v", err)
	}
}

func TestCleanupWorkspace_MissingDirsIsNoError(t *testing.T) {
	ctx := &terraformExecutorContext{manifestsDir: t.TempDir(), tfWorkspace: "never-created"}
	if err := ctx.cleanupWorkspace(); err != nil {
		t.Fatalf("cleanupWorkspace() on missing dirs should be a no-op, got: %v", err)
	}
}
