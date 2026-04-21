// Copyright 2024 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudrun

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecideRevisionName_Digest(t *testing.T) {
	testcases := []struct {
		name        string
		serviceName string
		image       string
		commit      string
		wantLen     int
		wantPrefix  string
	}{
		{
			name:        "digest image",
			serviceName: "helloworld",
			image:       "gcr.io/pipecd/helloworld@sha256:1bd4d708f94fcfb58b51b3a0818edc52bcc079104fc0fa3347ff0f309f9ed8b2",
			commit:      "ca1937a",
			wantLen:     63, // or less
		},
		{
			name:        "very long service name",
			serviceName: "a-very-long-service-name-that-is-already-near-the-limit-of-sixty-three-characters",
			image:       "gcr.io/pipecd/helloworld:v1.0.0",
			commit:      "1234567",
			wantLen:     63,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := `
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: ` + tc.serviceName + `
spec:
  template:
    spec:
      containers:
      - image: ` + tc.image + `
`
			sm, err := ParseServiceManifest([]byte(manifest))
			require.NoError(t, err)

			name, err := DecideRevisionName(sm, tc.commit)
			require.NoError(t, err)

			t.Logf("Generated revision name: %s (length: %d)", name, len(name))
			assert.True(t, len(name) < 64, "Revision name must be less than 64 characters, but got %d: %s", len(name), name)
		})
	}
}
