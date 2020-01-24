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
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"regexp"

	"github.com/cloudfoundry/libcfbuildpack/logger"
)

var pattern = regexp.MustCompile(".+/(.*)-([\\d].*)\\.jar")

// JARDependency represents a JAR dependency within an application
type JARDependency struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	SHA256  string `toml:"sha256"`
}

// NewJARDependency creates a new instance of JAR dependency, returning true if it matches the standard Maven naming
// scheme.
func NewJARDependency(path string, logger logger.Logger) (JARDependency, bool, error) {
	m := pattern.FindStringSubmatch(path)
	if m == nil {
		return JARDependency{}, false, nil
	}

	h, err := hash(path)
	if err != nil {
		return JARDependency{}, false, err
	}

	return JARDependency{
		Name:    m[1],
		Version: m[2],
		SHA256:  h,
	}, true, nil
}

func hash(file string) (string, error) {
	s := sha256.New()

	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(s, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(s.Sum(nil)), nil
}
