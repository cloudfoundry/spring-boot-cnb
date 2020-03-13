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
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/v2/layers"
	"github.com/cloudfoundry/libcfbuildpack/v2/test"
	"github.com/cloudfoundry/spring-boot-cnb/springboot"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDefaultSlicer(t *testing.T) {
	spec.Run(t, "DefaultSlicer", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("Slice", func() {

			var (
				slices layers.Slices
				s      springboot.DefaultSlicer
			)

			it.Before(func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

				md, ok, err := springboot.NewMetadata(f.Build.Application, f.Build.Logger)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())

				e := springboot.NewDefaultSlicer(f.Build.Application.Root, md)

				s = e
			})

			it("adds application files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-classes", "org", "cloudfoundry", "Test.class")

				slices = layers.Slices{
					{},
					{},
					{},
					{Paths: []string{"test-classes/org/cloudfoundry/Test.class"}},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

			it("adds dependency files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-lib", "test-1.2.3.jar")

				slices = layers.Slices{
					{},
					{Paths: []string{"test-lib/test-1.2.3.jar"}},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

			it("adds launch files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "org", "cloudfoundry", "Test.class")

				slices = layers.Slices{
					{Paths: []string{"org/cloudfoundry/Test.class"}},
					{},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

			it("adds snapshot files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "test-lib", "test-1.2.3-SNAPSHOT.jar")

				slices = layers.Slices{
					{},
					{},
					{Paths: []string{"test-lib/test-1.2.3-SNAPSHOT.jar"}},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

			it("adds remainder files to slice", func() {
				test.TouchFile(t, f.Build.Application.Root, "META-INF", "test-file")

				slices = layers.Slices{
					{},
					{},
					{},
					{},
					{Paths: []string{"META-INF/MANIFEST.MF", "META-INF/test-file"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})
		})

	}, spec.Report(report.Terminal{}))
}
