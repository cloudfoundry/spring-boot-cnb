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

func TestLayersIndexSlicer(t *testing.T) {
	spec.Run(t, "LayersIndexSlicer", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("Slice", func() {

			var (
				slices layers.Slices
				s      springboot.LayersIndexSlicer
			)

			it.Before(func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Version: test-version
Spring-Boot-Layers-Index: test-inf/layers.idx`)

				e := springboot.NewLayersIndexSlicer(f.Build.Application.Root, "test-inf/layers.idx")

				s = e
			})

			it("returns error if Layers-Index not found", func() {
				_, err := s.Slice()
				g.Expect(err).To(gomega.HaveOccurred())
			})

			it("adds layers from index to slices with application default", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test-inf", "layers.idx"),
					`dependencies
application
resources`)
				test.TouchFile(t, f.Build.Application.Root, "org", "cloudfoundry", "Test.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "application", "org", "cloudfoundry", "TestApplication.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "unknown", "org", "cloudfoundry", "TestUnknown.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "resources", "static", "index.html")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "dependencies", "lib", "released-1.2.3.jar")

				slices = layers.Slices{
					{Paths: []string{"test-inf/layers/dependencies/lib/released-1.2.3.jar"}},
					{Paths: []string{
						"META-INF/MANIFEST.MF",
						"org/cloudfoundry/Test.class",
						"test-inf/layers/application/org/cloudfoundry/TestApplication.class",
						"test-inf/layers/unknown/org/cloudfoundry/TestUnknown.class",
						"test-inf/layers.idx",
					}},
					{Paths: []string{"test-inf/layers/resources/static/index.html"}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

			it("adds layers from index to slices without application default", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test-inf", "layers.idx"),
					`dependencies
classes`)
				test.TouchFile(t, f.Build.Application.Root, "org", "cloudfoundry", "Test.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "classes", "org", "cloudfoundry", "TestApplication.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "unknown", "org", "cloudfoundry", "TestUnknown.class")
				test.TouchFile(t, f.Build.Application.Root, "test-inf", "layers", "dependencies", "lib", "released-1.2.3.jar")

				slices = layers.Slices{
					{Paths: []string{"test-inf/layers/dependencies/lib/released-1.2.3.jar"}},
					{Paths: []string{
						"test-inf/layers/classes/org/cloudfoundry/TestApplication.class",
					}},
					{Paths: []string{
						"META-INF/MANIFEST.MF",
						"org/cloudfoundry/Test.class",
						"test-inf/layers/unknown/org/cloudfoundry/TestUnknown.class",
						"test-inf/layers.idx",
					}},
				}

				result, err := s.Slice()
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(slices))
			})

		})

	}, spec.Report(report.Terminal{}))
}
