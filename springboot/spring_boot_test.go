/*
 * Copyright 2018-2020 the original author or authors.
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

package springboot_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/spring-boot-cnb/springboot"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSpringBoot(t *testing.T) {
	spec.Run(t, "SpringBoot", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("NewSpringBoot", func() {

			it("returns false when no Spring-Boot-Version", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"), "")

				_, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(gomega.BeFalse())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("returns true when Spring-Boot-Version exists", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

				_, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		when("Slices", func() {

			var (
				metadata layers.Metadata
				s        springboot.SpringBoot
			)

			it.Before(func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

				e, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())

				s = e

				command := "java -cp $CLASSPATH $JAVA_OPTS test-start-class"
				metadata = layers.Metadata{
					Processes: []layers.Process{
						{Type: "spring-boot", Command: command},
						{Type: "task", Command: command},
						{Type: "web", Command: command},
					},
				}
			})

			it("adds application files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-classes", "org", "cloudfoundry", "Test.class")

				metadata.Slices = layers.Slices{
					{},
					{},
					{},
					{Paths: []string{"test-classes/org/cloudfoundry/Test.class"}},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				g.Expect(s.Contribute()).To(gomega.Succeed())
				g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(metadata))
			})

			it("adds dependency files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-lib", "test-1.2.3.jar")

				metadata.Slices = layers.Slices{
					{},
					{Paths: []string{"test-lib/test-1.2.3.jar"}},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				g.Expect(s.Contribute()).To(gomega.Succeed())
				g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(metadata))
			})

			it("adds launch files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "org", "cloudfoundry", "Test.class")

				metadata.Slices = layers.Slices{
					{Paths: []string{"org/cloudfoundry/Test.class"}},
					{},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				g.Expect(s.Contribute()).To(gomega.Succeed())
				g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(metadata))
			})

			it("adds snapshot files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-lib", "test-1.2.3-SNAPSHOT.jar")

				metadata.Slices = layers.Slices{
					{},
					{},
					{Paths: []string{"test-lib/test-1.2.3-SNAPSHOT.jar"}},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				g.Expect(s.Contribute()).To(gomega.Succeed())
				g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(metadata))
			})

			it("adds remainder files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "META-INF", "test-file")

				metadata.Slices = layers.Slices{
					{},
					{},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF", "META-INF/test-file"}},
				}

				g.Expect(s.Contribute()).To(gomega.Succeed())
				g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(metadata))
			})
		})

		it("contributes dependencies to BOM", func() {
			test.CopyFile(t, filepath.Join("testdata", "test-artifact-1-1.2.3.jar"),
				filepath.Join(f.Build.Application.Root, "test-lib", "test-artifact-1-1.2.3.jar"))
			test.CopyFile(t, filepath.Join("testdata", "test-artifact-2-4.5.6-SNAPSHOT.jar"),
				filepath.Join(f.Build.Application.Root, "test-lib", "test-artifact-2-4.5.6-SNAPSHOT.jar"))

			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
				`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

			e, ok, err := springboot.NewSpringBoot(f.Build)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(e.Plan()).To(gomega.Equal(buildpackplan.Plan{
				Name:    springboot.Dependency,
				Version: "",
				Metadata: buildpackplan.Metadata{
					"lib":         "test-lib",
					"start-class": "test-start-class",
					"version":     "test-version",
					"classes":     "test-classes",
					"classpath": []string{
						filepath.Join(f.Build.Application.Root, "test-classes"),
						filepath.Join(f.Build.Application.Root, "test-lib", "test-artifact-1-1.2.3.jar"),
						filepath.Join(f.Build.Application.Root, "test-lib", "test-artifact-2-4.5.6-SNAPSHOT.jar"),
					},
					"dependencies": springboot.JARDependencies{
						{
							Name:    "test-artifact-1",
							Version: "1.2.3",
							SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
						},
						{
							Name:    "test-artifact-2",
							Version: "4.5.6-SNAPSHOT",
							SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
						},
					},
				},
			}))
		})

		it("contributes command", func() {
			test.TouchFile(t, filepath.Join(f.Build.Application.Root, "test-lib", "test.jar"))
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
				`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

			e, ok, err := springboot.NewSpringBoot(f.Build)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(e.Contribute()).To(gomega.Succeed())

			layer := f.Build.Layers.Layer("spring-boot")
			g.Expect(layer).To(test.HaveLayerMetadata(true, true, true))
			g.Expect(layer).To(test.HavePrependPathSharedEnvironment("CLASSPATH", strings.Join([]string{
				filepath.Join(f.Build.Application.Root, "test-classes"),
				filepath.Join(f.Build.Application.Root, "test-lib", "test.jar"),
			}, string(filepath.ListSeparator))))

			command := "java -cp $CLASSPATH $JAVA_OPTS test-start-class"
			g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Slices: layers.Slices{
					{},
					{Paths: []string{"test-lib/test.jar"}},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				},
				Processes: layers.Processes{
					{Type: "spring-boot", Command: command},
					{Type: "task", Command: command},
					{Type: "web", Command: command},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
