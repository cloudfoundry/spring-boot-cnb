/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/spring-boot-cnb/cli"
	"github.com/cloudfoundry/spring-boot-cnb/springboot"
)

func main() {
	build, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build: %s\n", err)
		os.Exit(101)
	}

	if code, err := b(build); err != nil {
		build.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func b(build build.Build) (int, error) {
	var ps []buildpackplan.Plan

	if s, ok, err := springboot.NewSpringBoot(build); err != nil {
		return build.Failure(102), err
	} else if ok {
		build.Logger.Title(build.Buildpack)

		if err = s.Contribute(); err != nil {
			return build.Failure(103), err
		}

		p, err := s.Plan()
		if err != nil {
			return build.Failure(103), err
		}

		ps = append(ps, p)
	}

	if c, ok, err := cli.NewCommand(build); err != nil {
		return build.Failure(102), err
	} else if ok {
		build.Logger.Title(build.Buildpack)

		if l, err := cli.NewCLI(build); err != nil {
			return build.Failure(102), err
		} else {
			if err := l.Contribute(); err != nil {
				return build.Failure(103), err
			}
		}

		if err = c.Contribute(); err != nil {
			return build.Failure(103), err
		}
	}

	return build.Success(ps...)
}
