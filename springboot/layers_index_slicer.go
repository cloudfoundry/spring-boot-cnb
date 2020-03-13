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
)

type LayersIndexSlicer struct {
	applicationRoot string
	layersIndex     LayersIndex
}

func (s LayersIndexSlicer) Slice() (layers.Slices, error) {
	layerNames, err := s.layersIndex.layerNames()
	if err != nil {
		return layers.Slices{}, err
	}

	slicesByLayerName := make(map[string]layers.Slice)

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

		layerName := s.layersIndex.layerNameForPath(rel, layerNames)
		layer, exists := slicesByLayerName[layerName]
		if !exists {
			layer = layers.Slice{}
		}
		layer.Paths = append(layer.Paths, rel)
		slicesByLayerName[layerName] = layer

		return nil
	}); err != nil {
		return layers.Slices{}, err
	}

	var result []layers.Slice
	for _, layerName := range layerNames {
		result = append(result, slicesByLayerName[layerName])
	}
	return result, nil
}

func NewLayersIndexSlicer(applicationRoot string, layersIndex string) LayersIndexSlicer {
	return LayersIndexSlicer{
		applicationRoot,
		NewLayersIndex(applicationRoot, layersIndex),
	}
}
