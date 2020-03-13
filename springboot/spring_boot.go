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

package springboot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/buildpacks/libbuildpack/v2/application"
	"github.com/cloudfoundry/libcfbuildpack/v2/build"
	"github.com/cloudfoundry/libcfbuildpack/v2/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/v2/helper"
	"github.com/cloudfoundry/libcfbuildpack/v2/layers"
	"github.com/cloudfoundry/libcfbuildpack/v2/logger"
	"github.com/mitchellh/mapstructure"
)

// Dependency indicates that an application is a Spring Boot application.
const Dependency = "spring-boot"

// SpringBoot represents a Spring Boot JVM application.
type SpringBoot struct {
	// Metadata is metadata about the Spring Boot application.
	Metadata Metadata

	application application.Application
	layer       layers.Layer
	layers      layers.Layers
	logger      logger.Logger

	slicer Slicer
}

// Contribute makes the contribution to build, cache, and launch.
func (s SpringBoot) Contribute() error {
	if err := s.layer.Contribute(s.Metadata, func(layer layers.Layer) error {
		return layer.PrependPathSharedEnv("CLASSPATH", strings.Join(s.Metadata.ClassPath, string(filepath.ListSeparator)))
	}, layers.Build, layers.Cache, layers.Launch); err != nil {
		return err
	}

	slices, err := s.slicer.Slice()
	if err != nil {
		return err
	}

	command := fmt.Sprintf("java -cp $CLASSPATH $JAVA_OPTS %s", s.Metadata.StartClass)

	return s.layers.WriteApplicationMetadata(layers.Metadata{
		Slices: slices,
		Processes: layers.Processes{
			{Type: "spring-boot", Command: command},
			{Type: "task", Command: command},
			{Type: "web", Command: command},
		},
	})
}

// Plan returns the dependency information for this application.
func (s SpringBoot) Plan() (buildpackplan.Plan, error) {
	p := buildpackplan.Plan{
		Name:     Dependency,
		Metadata: buildpackplan.Metadata{},
	}

	if err := mapstructure.Decode(s.Metadata, &p.Metadata); err != nil {
		return buildpackplan.Plan{}, err
	}

	if d, err := s.dependencies(); err != nil {
		return buildpackplan.Plan{}, err
	} else {
		p.Metadata["dependencies"] = d
	}

	return p, nil
}

type result struct {
	err   error
	value JARDependency
}

func (s SpringBoot) dependencies() (JARDependencies, error) {
	ch := make(chan result)
	var wg sync.WaitGroup

	l := filepath.Join(s.application.Root, s.Metadata.Lib)
	if exists, err := helper.FileExists(l); err != nil {
		return JARDependencies{}, err
	} else if !exists {
		return JARDependencies{}, nil
	}

	if err := filepath.Walk(l, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			d, ok, err := NewJARDependency(path, s.logger)
			if err != nil {
				ch <- result{err: err}
				return
			}

			if ok {
				ch <- result{value: d}
			}
		}()

		return nil
	}); err != nil {
		return nil, err
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var d JARDependencies
	for r := range ch {
		if r.err != nil {
			return JARDependencies{}, r.err
		}

		d = append(d, r.value)
	}
	sort.Sort(d)

	return d, nil
}

// NewSpringBoot creates a new SpringBoot instance.  OK is true if the build plan contains a "jvm-application"
// dependency and a "Spring-Boot-Version" manifest key.
func NewSpringBoot(build build.Build) (SpringBoot, bool, error) {
	md, ok, err := NewMetadata(build.Application, build.Logger)
	if err != nil {
		return SpringBoot{}, false, err
	}

	if !ok {
		return SpringBoot{}, false, nil
	}

	var slicer Slicer
	if md.LayersIndex == "" {
		slicer = NewDefaultSlicer(build.Application.Root, md)
	} else {
		slicer = NewLayersIndexSlicer(build.Application.Root, md.LayersIndex)
	}

	return SpringBoot{
		md,
		build.Application,
		build.Layers.Layer(Dependency),
		build.Layers,
		build.Logger,
		slicer,
	}, true, nil
}

type Slicer interface {
	Slice() (layers.Slices, error)
}
