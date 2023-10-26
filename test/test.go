package test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/emed-appts/emed-mailer/internal/config"
)

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}

// PrepareTestEnvironment is a reusable TestMain(...) function for unit tests that need a complete
// application environment.
// Therefore it loads a special config file for testing purposes.
func PrepareTestEnvironment(m *testing.M, pathToRoot string) {
	// set AppWorkPath
	config.AppWorkPath = path.Join(pathToRoot, "test")

	// setup test config
	config.General.Root = path.Join(config.AppWorkPath, "data")
	if err := os.MkdirAll(config.General.Root, os.ModePerm); err != nil {
		fatalTestError("could not create folders of root path: %v", err)
	}

	exitCode := m.Run()

	// cleanup generated data
	if err := os.RemoveAll(config.General.Root); err != nil {
		fatalTestError("cleanup root folder failed: %v", err)
	}

	os.Exit(exitCode)
}
