// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
)

// envGitCeilingDirectories bounds how far up the directory tree git will search
// for a repository. Git excludes the ceiling directory itself from the search.
const envGitCeilingDirectories = "GIT_CEILING_DIRECTORIES"

// gitDiscoveryEnvVars are the environment variables git consults to locate a
// repository instead of discovering one from the working directory. Git exports
// GIT_DIR to every hook it runs, so these leak into `make pre-push-checks` and
// would otherwise make the discovery specs below assert against the real repo.
var gitDiscoveryEnvVars = []string{
	"GIT_DIR",
	"GIT_WORK_TREE",
	"GIT_COMMON_DIR",
	"GIT_INDEX_FILE",
	"GIT_OBJECT_DIRECTORY",
	envGitCeilingDirectories,
	"GIT_DISCOVERY_ACROSS_FILESYSTEM",
}

var _ = Describe("Test config", func() {
	Context("GetRootDir works well", func() {
		var oldWorkDir string
		var tmpWorkDir string
		var testSubDir string
		var err error
		var oldWorkspace string
		var oldGitEnv map[string]string

		BeforeEach(func() {
			// Clear any inherited git environment so `git` discovers the repo
			// from the working directory rather than from GIT_DIR & friends
			oldGitEnv = map[string]string{}
			for _, key := range gitDiscoveryEnvVars {
				if value, ok := os.LookupEnv(key); ok {
					oldGitEnv[key] = value
					Expect(os.Unsetenv(key)).To(Succeed())
				}
			}

			// Create a temporary working directory. Normalize it to an absolute,
			// symlink-free path so it matches what `git rev-parse --show-toplevel`
			// and os.Getwd report (e.g. macOS resolves TMPDIR through /private)
			tmpWorkDir, err = os.MkdirTemp("", "GetRootDir-*")
			Expect(err).ToNot(HaveOccurred())
			tmpWorkDir, err = filepath.Abs(tmpWorkDir)
			Expect(err).ToNot(HaveOccurred())
			tmpWorkDir, err = filepath.EvalSymlinks(tmpWorkDir)
			Expect(err).ToNot(HaveOccurred())
			oldWorkDir, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			err = os.Chdir(tmpWorkDir)
			Expect(err).ToNot(HaveOccurred())

			// Create test sub-directory
			testSubDir = tmpWorkDir + "/subdir"
			err = os.Mkdir(testSubDir, 0755)
			Expect(err).ToNot(HaveOccurred())

			// Store and clear the WORKSPACE environment variable
			oldWorkspace = os.Getenv(EnvWorkspace)
			err = os.Unsetenv(EnvWorkspace)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// Restore the old working directory
			err = os.Chdir(oldWorkDir)
			Expect(err).ToNot(HaveOccurred())
			// Delete the temporary working directory
			err = os.RemoveAll(tmpWorkDir)
			Expect(err).ToNot(HaveOccurred())

			// Restore the old WORKSPACE environment variable if it was set
			if oldWorkspace != "" {
				err = os.Setenv(EnvWorkspace, oldWorkspace)
				Expect(err).ToNot(HaveOccurred())
			}

			// Restore any git environment variables that were cleared
			for key, value := range oldGitEnv {
				Expect(os.Setenv(key, value)).To(Succeed())
			}
		})

		It("GetRootDir works when WORKSPACE env is set", func() {
			workspace := "TEST VALUE"
			err = os.Setenv(EnvWorkspace, workspace)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Unsetenv(EnvWorkspace)
				Expect(err).ToNot(HaveOccurred())
			}()

			rootDir := GetRootDir()
			Expect(rootDir).To(Equal(workspace))
		})
		It("GetRootDir works when WORKSPACE env is unset and in git workspace", func() {
			cmd := exec.Command("git", "init")
			err = cmd.Run()
			Expect(err).ToNot(HaveOccurred())

			// Guard against a leaked git environment redirecting `git init`
			// somewhere other than the temporary working directory
			_, err = os.Stat(tmpWorkDir + "/.git")
			Expect(err).ToNot(HaveOccurred())

			err = os.Chdir(testSubDir)
			Expect(err).ToNot(HaveOccurred())

			rootDir := GetRootDir()
			Expect(rootDir).To(Equal(tmpWorkDir))
			Expect(rootDir).ToNot(Equal(testSubDir))
		})
		It("GetRootDir works when WORKSPACE env is unset and not in git folder", func() {
			// Stop git from discovering a repository that encloses the temporary
			// directory, which happens when TMPDIR itself lives inside a checkout
			err = os.Setenv(envGitCeilingDirectories, tmpWorkDir)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Unsetenv(envGitCeilingDirectories)
				Expect(err).ToNot(HaveOccurred())
			}()

			err = os.Chdir(testSubDir)
			Expect(err).ToNot(HaveOccurred())

			rootDir := GetRootDir()
			Expect(rootDir).ToNot(Equal(tmpWorkDir))
			Expect(rootDir).To(Equal(testSubDir))
		})
	})
})
