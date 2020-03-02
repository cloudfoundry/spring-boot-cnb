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
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

type LayersIndex struct {
	applicationRoot string
	layersIndexFile string

	layersIndexDir   string
	defaultLayerName string

	layerPathRegex *regexp.Regexp
}

func (l LayersIndex) layerNames() ([]string, error) {
	if l.layersIndexFile == "" {
		return nil, nil
	}

	layerNames, err := readFileLines(filepath.Join(l.applicationRoot, l.layersIndexFile))
	if err != nil {
		return nil, err
	}

	if !sliceContains(layerNames, l.defaultLayerName) {
		layerNames = append(layerNames, l.defaultLayerName)
	}

	return layerNames, nil
}

func (l LayersIndex) layerClassPaths() ([]string, error) {
	if l.layersIndexFile == "" {
		return nil, nil
	}

	layerNames, err := l.layerNames()
	if err != nil {
		return nil, err
	}

	var layerPaths []string
	for _, layerName := range layerNames {
		layerPaths = append(layerPaths, filepath.Join(l.applicationRoot, l.layersIndexDir, "layers", layerName, "classes"))
	}

	return layerPaths, nil
}

func (l LayersIndex) layerNameForPath(path string, layerNames []string) string {
	matches := l.layerPathRegex.FindStringSubmatch(path)
	if len(matches) > 1 {
		layerName := matches[1]
		if sliceContains(layerNames, layerName) {
			return layerName
		}
	}
	return l.defaultLayerName
}

func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var layerNames []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			layerNames = append(layerNames, text)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return layerNames, nil
}

func sliceContains(s []string, val string) (ok bool) {
	for i := range s {
		if ok = s[i] == val; ok {
			return
		}
	}
	return
}

func NewLayersIndex(applicationRoot string, layersIndexFile string) LayersIndex {
	layersIndexDir, _ := filepath.Split(layersIndexFile)

	return LayersIndex{
		applicationRoot,
		layersIndexFile,
		layersIndexDir,
		"application",
		regexp.MustCompile(`^` + layersIndexDir + `layers/([a-zA-Z0-9-]+)/.*$`),
	}
}
