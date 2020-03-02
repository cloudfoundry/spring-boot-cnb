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
	"github.com/cloudfoundry/libcfbuildpack/v2/layers"
	"os"
	"path/filepath"
	"strings"
)

type DefaultSlicer struct {
	applicationRoot string
	metadata        Metadata
}

func (s DefaultSlicer) Slice() (layers.Slices, error) {
	var app, dep, launch, snap, rem layers.Slice

	if err := filepath.Walk(s.applicationRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(s.applicationRoot, path)
		if err != nil {
			return err
		}

		if s.isApplicationSlice(rel) {
			app.Paths = append(app.Paths, rel)
		} else if s.isDependencySlice(rel) {
			dep.Paths = append(dep.Paths, rel)
		} else if s.isLaunchSlice(rel) {
			launch.Paths = append(launch.Paths, rel)
		} else if s.isSnapshotSlice(rel) {
			snap.Paths = append(snap.Paths, rel)
		} else {
			rem.Paths = append(rem.Paths, rel)
		}

		return nil
	}); err != nil {
		return layers.Slices{}, err
	}

	return layers.Slices{launch, dep, snap, app, rem}, nil // intentionally ordered
}

func (s DefaultSlicer) isApplicationSlice(path string) bool {
	return strings.HasPrefix(path, s.metadata.Classes)
}

func (s DefaultSlicer) isDependencySlice(path string) bool {
	return strings.HasPrefix(path, s.metadata.Lib) && filepath.Ext(path) == ".jar" && !strings.Contains(path, "SNAPSHOT")
}

func (s DefaultSlicer) isLaunchSlice(path string) bool {
	return !strings.HasPrefix(path, s.metadata.Classes) && !strings.HasPrefix(path, s.metadata.Lib) && !strings.HasPrefix(path, "META-INF/")
}

func (s DefaultSlicer) isSnapshotSlice(path string) bool {
	return strings.HasPrefix(path, s.metadata.Lib) && filepath.Ext(path) == ".jar" && strings.Contains(path, "SNAPSHOT")
}

func NewDefaultSlicer(applicationRoot string, metadata Metadata) DefaultSlicer {
	return DefaultSlicer{
		applicationRoot,
		metadata,
	}
}
