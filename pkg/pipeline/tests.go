package pipeline

import (
	"os"

	"dagger.io/dagger"
)

const (
	testsPath = "./test/"
)

// Tests returns containers for all tests.
func Tests(client *dagger.Client, brokers map[string]*dagger.Service) map[string]*dagger.Container {
	containers := make(map[string]*dagger.Container, 0)

	// Set examples
	for _, p := range testsPaths() {
		t := client.Container().
			// Add base image
			From(GolangImage).
			// Add source code as work directory
			With(sourceAsWorkdir(client)).
			// Set brokers as dependencies of app and user
			With(BindBrokers(brokers)).
			// Execute command
			WithExec([]string{"go", "test", testsPath + p})

		// Add user containers to containers
		containers[p] = t
	}

	return containers
}

func testsPaths() []string {
	paths := make([]string, 0)

	tests, err := os.ReadDir("./test")
	if err != nil {
		panic(err)
	}

	for _, t := range tests {
		if !t.Type().IsDir() {
			continue
		}

		subtests, err := os.ReadDir(testsPath + t.Name())
		if err != nil {
			panic(err)
		}

		for _, st := range subtests {
			if !st.Type().IsDir() {
				continue
			}

			paths = append(paths, t.Name()+"/"+st.Name())
		}
	}

	return paths
}
