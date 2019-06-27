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

package cli_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/spring-boot-cnb/cli"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCommand(t *testing.T) {
	spec.Run(t, "Spring Boot CLI Command", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("NewCommand", func() {
			it("returns false when no jvm-application", func() {
				test.TouchFile(t, f.Build.Application.Root, "test.groovy")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns false when no groovy files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns true when jvm-application and groovy files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.CopyDirectory(t, filepath.Join("testdata", "valid_app"), f.Build.Application.Root)

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("ignores .groovy directories", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.TouchFile(t, f.Build.Application.Root, "test.groovy", "test")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("rejects non-POGO, non-config files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "x")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("ignores logback files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "ch", "qos", "logback", "test.groovy"), "class X {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("detects POGO files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "class X {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("detects config files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "beans {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("detects invalid .groovy files", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.CopyFile(t, filepath.Join("testdata", "valid_app", "invalid.groovy"), filepath.Join(f.Build.Application.Root, "test.groovy"))

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

		})

		it("contributes command", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
			test.CopyDirectory(t, filepath.Join("testdata", "valid_app"), f.Build.Application.Root)

			c, ok, err := cli.NewCommand(f.Build)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(c.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("command")
			g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			g.Expect(layer).To(test.HaveAppendLaunchEnvironment("GROOVY_FILES", strings.Join([]string{
				"",
				filepath.Join(f.Build.Application.Root, "directory", "pogo_4.groovy"),
				filepath.Join(f.Build.Application.Root, "invalid.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_1.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_2.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_3.groovy"),
			}, " ")))

			command := "spring run -cp $CLASSPATH $GROOVY_FILES"
			g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Processes: []layers.Process{
					{"spring-boot-cli", command},
					{"task", command},
					{"web", command},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
