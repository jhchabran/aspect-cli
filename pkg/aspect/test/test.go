/*
 * Copyright 2022 Aspect Build Systems, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"context"
	"fmt"

	"aspect.build/cli/pkg/aspecterrors"
	"aspect.build/cli/pkg/bazel"
	"aspect.build/cli/pkg/ioutils"
	"aspect.build/cli/pkg/plugin/system/bep"
	"github.com/spf13/cobra"
)

type Test struct {
	ioutils.Streams
	bzl bazel.Bazel
}

func New(streams ioutils.Streams, bzl bazel.Bazel) *Test {
	return &Test{
		Streams: streams,
		bzl:     bzl,
	}
}

func (runner *Test) Run(ctx context.Context, _ *cobra.Command, args []string) (exitErr error) {
	bazelCmd := []string{"test"}

	if bep.HasBESBackend(ctx) {
		besBackend := bep.BESBackendFromContext(ctx)
		besBackendFlag := fmt.Sprintf("--bes_backend=%s", besBackend.Addr())
		bazelCmd = append(bazelCmd, besBackendFlag)
	}

	bazelCmd = append(bazelCmd, args...)
	exitCode, bazelErr := runner.bzl.RunCommand(runner.Streams, nil, bazelCmd...)

	// Process the subscribers errors before the Bazel one.
	subscriberErrors := bep.BESErrors(ctx)
	if len(subscriberErrors) > 0 {
		for _, err := range subscriberErrors {
			fmt.Fprintf(runner.Streams.Stderr, "Error: failed to run test command: %v\n", err)
		}
		exitCode = 1
	}

	if exitCode != 0 {
		err := &aspecterrors.ExitError{ExitCode: exitCode}
		if bazelErr != nil {
			err.Err = bazelErr
		}
		return err
	}

	return nil
}
