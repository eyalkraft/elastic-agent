// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package composed

import (
	"errors"
	"testing"

	"github.com/elastic/elastic-agent/internal/pkg/agent/application/upgrade/artifact"
	"github.com/elastic/elastic-agent/internal/pkg/agent/application/upgrade/artifact/download"

	"github.com/stretchr/testify/assert"
)

type ErrorVerifier struct {
	called bool
}

func (d *ErrorVerifier) Verify(a artifact.Artifact, version string, _ bool, _ ...string) error {
	d.called = true
	return errors.New("failing")
}

func (d *ErrorVerifier) Called() bool { return d.called }

type FailVerifier struct {
	called bool
}

func (d *FailVerifier) Verify(a artifact.Artifact, version string, _ bool, _ ...string) error {
	d.called = true
	return &download.InvalidSignatureError{}
}

func (d *FailVerifier) Called() bool { return d.called }

type SuccVerifier struct {
	called bool
}

func (d *SuccVerifier) Verify(a artifact.Artifact, version string, _ bool, _ ...string) error {
	d.called = true
	return nil
}

func (d *SuccVerifier) Called() bool { return d.called }

func TestVerifier(t *testing.T) {
	testCases := []verifyTestCase{
		{
			verifiers:      []CheckableVerifier{&ErrorVerifier{}, &SuccVerifier{}, &FailVerifier{}},
			checkFunc:      func(d []CheckableVerifier) bool { return d[0].Called() && d[1].Called() && !d[2].Called() },
			expectedResult: true,
		}, {
			verifiers:      []CheckableVerifier{&SuccVerifier{}, &ErrorVerifier{}, &FailVerifier{}},
			checkFunc:      func(d []CheckableVerifier) bool { return d[0].Called() && !d[1].Called() && !d[2].Called() },
			expectedResult: true,
		}, {
			verifiers:      []CheckableVerifier{&FailVerifier{}, &ErrorVerifier{}, &SuccVerifier{}},
			checkFunc:      func(d []CheckableVerifier) bool { return d[0].Called() && !d[1].Called() && !d[2].Called() },
			expectedResult: false,
		}, {
			verifiers:      []CheckableVerifier{&ErrorVerifier{}, &FailVerifier{}, &SuccVerifier{}},
			checkFunc:      func(d []CheckableVerifier) bool { return d[0].Called() && d[1].Called() && !d[2].Called() },
			expectedResult: false,
		}, {
			verifiers:      []CheckableVerifier{&ErrorVerifier{}, &ErrorVerifier{}, &SuccVerifier{}},
			checkFunc:      func(d []CheckableVerifier) bool { return d[0].Called() && d[1].Called() && d[2].Called() },
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		d := NewVerifier(tc.verifiers[0], tc.verifiers[1], tc.verifiers[2])
		err := d.Verify(artifact.Artifact{Name: "a", Cmd: "a", Artifact: "a/a"}, "b", false)

		assert.Equal(t, tc.expectedResult, err == nil)

		assert.True(t, tc.checkFunc(tc.verifiers))
	}
}

type CheckableVerifier interface {
	download.Verifier
	Called() bool
}

type verifyTestCase struct {
	verifiers      []CheckableVerifier
	checkFunc      func(verifiers []CheckableVerifier) bool
	expectedResult bool
}
