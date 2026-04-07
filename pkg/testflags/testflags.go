package testflags

import (
	"os"
	"strings"
	"testing"

	"github.com/khalilonline/gokart/pkg/utils"
)

type TestType string

const (
	testTypeEnvVar          = "TEST_TYPE"
	Performance    TestType = "performance"
	Security       TestType = "security"
	Unit           TestType = "unit"
	Integration    TestType = "integration"
)

func skip(t testing.TB, expectedTestTypes ...TestType) {
	t.Helper()
	switch len(expectedTestTypes) {
	case 0:
		t.Skipf("skipping test")
	case 1:
		t.Skipf("skipping test, set environment variable %s to %s", testTypeEnvVar, string(expectedTestTypes[0]))
	default:
		t.Skipf("skipping test, set environment variable %s to one of (%s)", testTypeEnvVar, strings.Join(utils.ToStrings(expectedTestTypes), ", "))
	}
}

// PerformanceTest is a helper function that skips the test if TEST_TYPE is not set to "performance".
func PerformanceTest(t testing.TB) {
	t.Helper()
	Evaluate(t, Performance)
}

// SecurityTest is a helper function that skips the test if TEST_TYPE is not set to "security".
func SecurityTest(t testing.TB) {
	t.Helper()
	Evaluate(t, Security)
}

// UnitTest is a helper function that skips the test if TEST_TYPE is not set to "unit".
func UnitTest(t testing.TB) {
	t.Helper()
	Evaluate(t, Unit)
}

// IntegrationTest is a helper function that skips the test if TEST_TYPE is not set to "integration".
func IntegrationTest(t testing.TB) {
	t.Helper()
	Evaluate(t, Integration)
}

// Evaluate evaluates the parsed test type against the TestTypes. If there is no match, the test t is skipped.
//
// You can combine your own TestTypes with those defined in this package like so:
//
//	var readerIntegration flags.TestType = "reader_integration"
//
//	func TestReader(t *testing.T) {
//	  flags.Evaluate(t, readerIntegration, flags.Integration)
//	  ...
//	  Test will only run if TEST_TYPE is set to "reader_integration" or "integration"
//	  ...
//	}
func Evaluate(t testing.TB, testTypes ...TestType) {
	t.Helper()
	parsedVal := strings.ToLower(os.Getenv(testTypeEnvVar))
	for _, tt := range testTypes {
		if strings.ToLower(string(tt)) == parsedVal {
			return
		}
	}
	skip(t, testTypes...)
}
