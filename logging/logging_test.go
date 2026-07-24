// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging", func() {
	var (
		originalTFLog string
		hadTFLog      bool
	)

	BeforeEach(func() {
		originalTFLog, hadTFLog = os.LookupEnv("TF_LOG")
	})

	AfterEach(func() {
		if hadTFLog {
			Expect(os.Setenv("TF_LOG", originalTFLog)).To(Succeed())
		} else {
			Expect(os.Unsetenv("TF_LOG")).To(Succeed())
		}
	})

	setTFLog := func(value string, set bool) {
		if set {
			Expect(os.Setenv("TF_LOG", value)).To(Succeed())
		} else {
			Expect(os.Unsetenv("TF_LOG")).To(Succeed())
		}
	}

	DescribeTable("tfLogLevel normalizes the TF_LOG environment variable",
		func(value string, set bool, expected string) {
			setTFLog(value, set)
			Expect(tfLogLevel()).To(Equal(expected))
		},
		Entry("unset -> empty", "", false, ""),
		Entry("empty -> empty", "", true, ""),
		Entry("TRACE -> TRACE", "TRACE", true, "TRACE"),
		Entry("DEBUG -> DEBUG", "DEBUG", true, "DEBUG"),
		Entry("INFO -> INFO", "INFO", true, "INFO"),
		Entry("WARN -> WARN", "WARN", true, "WARN"),
		Entry("ERROR -> ERROR", "ERROR", true, "ERROR"),
		Entry("lowercase is upper-cased", "info", true, "INFO"),
		Entry("mixed case is upper-cased", "TrAcE", true, "TRACE"),
		Entry("invalid falls back to TRACE", "BOGUS", true, "TRACE"),
	)

	DescribeTable("New enables the expected levels for a given TF_LOG",
		func(value string, set bool, wantDebug, wantInfo, wantWarn, wantError bool) {
			setTFLog(value, set)
			logger := New().(*TfLogger)
			Expect(logger.DebugEnabled()).To(Equal(wantDebug))
			Expect(logger.InfoEnabled()).To(Equal(wantInfo))
			Expect(logger.WarnEnabled()).To(Equal(wantWarn))
			Expect(logger.ErrorEnabled()).To(Equal(wantError))
		},
		Entry("unset -> error only", "", false, false, false, false, true),
		Entry("TRACE -> all levels", "TRACE", true, true, true, true, true),
		Entry("DEBUG -> all levels", "DEBUG", true, true, true, true, true),
		Entry("INFO -> info, warn, error", "INFO", true, false, true, true, true),
		Entry("WARN -> warn, error", "WARN", true, false, false, true, true),
		Entry("ERROR -> error only", "ERROR", true, false, false, false, true),
		Entry("invalid -> all levels (TRACE fallback)", "BOGUS", true, true, true, true, true),
	)
})
