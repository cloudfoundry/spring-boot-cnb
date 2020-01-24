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

	"github.com/cloudfoundry/libcfbuildpack/v2/test"
	"github.com/cloudfoundry/spring-boot-cnb/springboot"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMetadata(t *testing.T) {
	spec.Run(t, "Metadata", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.DetectFactory

		it.Before(func() {
			f = test.NewDetectFactory(t)
		})

		it("returns false if no Spring-Boot-Version", func() {
			_, ok, err := springboot.NewMetadata(f.Detect.Application, f.Detect.Logger)
			g.Expect(ok).To(gomega.BeFalse())
			g.Expect(err).NotTo(gomega.HaveOccurred())
		})

		it("parses Main-Class", func() {
			test.TouchFile(t, filepath.Join(f.Detect.Application.Root, "test-lib", "test.jar"))
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "META-INF", "MANIFEST.MF"),
				`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

			md, ok, err := springboot.NewMetadata(f.Detect.Application, f.Detect.Logger)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(md).To(gomega.Equal(springboot.Metadata{
				Classes: "test-classes",
				ClassPath: []string{
					filepath.Join(f.Detect.Application.Root, "test-classes"),
					filepath.Join(f.Detect.Application.Root, "test-lib", "test.jar"),
				},
				Lib:        "test-lib",
				StartClass: "test-start-class",
				Version:    "test-version",
			}))
		})
	}, spec.Report(report.Terminal{}))
}
