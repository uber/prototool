// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package create

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/uber/prototool/internal/protostrs"
	"github.com/uber/prototool/internal/settings"
	"go.uber.org/zap"
)

var (
	tmplV1 = template.Must(template.New("tmplV1").Parse(`syntax = "proto3";

package {{.Pkg}};

option go_package = "{{.GoPkg}}";
option java_multiple_files = true;
option java_outer_classname = "{{.JavaOuterClassname}}";
option java_package = "{{.JavaPkg}}";`))

	tmplV2 = template.Must(template.New("tmplV2").Parse(`syntax = "proto3";

package {{.Pkg}};

option csharp_namespace = "{{.CSharpNamespace}}";
option go_package = "{{.GoPkg}}";
option java_multiple_files = true;
option java_outer_classname = "{{.JavaOuterClassname}}";
option java_package = "{{.JavaPkg}}";
option objc_class_prefix = "{{.OBJCClassPrefix}}";
option php_namespace = "{{.PHPNamespace}}";`))
)

type tmplData struct {
	Pkg                string
	CSharpNamespace    string
	GoPkg              string
	JavaOuterClassname string
	JavaPkg            string
	OBJCClassPrefix    string
	PHPNamespace       string
}

type handler struct {
	logger         *zap.Logger
	configProvider settings.ConfigProvider
	pkg            string
}

func newHandler(options ...HandlerOption) *handler {
	handler := &handler{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(handler)
	}
	handler.configProvider = settings.NewConfigProvider(
		settings.ConfigProviderWithLogger(handler.logger),
	)
	return handler
}

func (h *handler) Create(filePaths ...string) error {
	for _, filePath := range filePaths {
		if err := h.checkFilePath(filePath); err != nil {
			return err
		}
	}
	for _, filePath := range filePaths {
		if err := h.create(filePath); err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) checkFilePath(filePath string) error {
	if filePath == "" {
		return errors.New("filePath empty")
	}
	dirPath := filepath.Dir(filePath)

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%q is not a directory somehow", dirPath)
	}
	if _, err := os.Stat(filePath); err == nil {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			return fmt.Errorf("%q already exists", filePath)
		}
	}
	return nil
}

func (h *handler) create(filePath string) error {
	pkg, err := h.getPkg(filePath)
	if err != nil {
		return err
	}
	isV2, err := h.isV2(filePath)
	if err != nil {
		return err
	}
	var data []byte
	if isV2 {
		data, err = getDataV2(
			&tmplData{
				Pkg:                pkg,
				CSharpNamespace:    protostrs.CSharpNamespace(pkg),
				GoPkg:              protostrs.GoPackageLastTwo(pkg),
				JavaOuterClassname: protostrs.JavaOuterClassname(filePath),
				JavaPkg:            protostrs.JavaPackage(pkg),
				OBJCClassPrefix:    protostrs.OBJCClassPrefix(pkg),
				PHPNamespace:       protostrs.PHPNamespace(pkg),
			},
		)
		if err != nil {
			return err
		}
	} else {
		data, err = getDataV1(
			&tmplData{
				Pkg:                pkg,
				GoPkg:              protostrs.GoPackage(pkg),
				JavaOuterClassname: protostrs.JavaOuterClassname(filePath),
				JavaPkg:            protostrs.JavaPackage(pkg),
			},
		)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

func (h *handler) isV2(filePath string) (bool, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}
	absDirPath := filepath.Dir(absFilePath)
	config, err := h.configProvider.GetForDir(absDirPath)
	if err != nil {
		return false, err
	}
	// no config file found
	if config.DirPath == "" {
		return false, nil
	}
	return strings.ToLower(config.Lint.Group) == "uber2", nil
}

func (h *handler) getPkg(filePath string) (string, error) {
	if h.pkg != "" {
		return h.pkg, nil
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	absDirPath := filepath.Dir(absFilePath)
	config, err := h.configProvider.GetForDir(absDirPath)
	if err != nil {
		return "", err
	}
	// no config file found, can't compute package
	if config.DirPath == "" {
		return DefaultPackage, nil
	}
	// we need to get all the matching directories and then choose the longest
	// ie if you have a, a/b, we choose a/b
	var longestCreateDirPath string
	var longestBasePkg string
	// note that createDirPath will always be absolute per the spec in
	// the settings package
	for createDirPath, basePkg := range config.Create.DirPathToBasePackage {
		// TODO: cannot do rel right away because it will do ../.. if necessary
		// strings.HasPrefix is not OS independent however
		if !strings.HasPrefix(absDirPath, createDirPath) {
			continue
		}
		if len(createDirPath) > len(longestCreateDirPath) {
			longestCreateDirPath = createDirPath
			longestBasePkg = basePkg
		}
	}
	if longestCreateDirPath != "" {
		rel, err := filepath.Rel(longestCreateDirPath, absDirPath)
		if err != nil {
			return "", err
		}
		return getPkgFromRel(rel, longestBasePkg), nil
	}

	// no package mapping found, do default logic

	// TODO: cannot do rel right away because it will do ../.. if necessary
	// strings.HasPrefix is not OS independent however
	if !strings.HasPrefix(absDirPath, config.DirPath) {
		return DefaultPackage, nil
	}
	rel, err := filepath.Rel(config.DirPath, absDirPath)
	if err != nil {
		return "", err
	}
	return getPkgFromRel(rel, ""), nil
}

func getPkgFromRel(rel string, basePkg string) string {
	if rel == "." {
		if basePkg == "" {
			return DefaultPackage
		}
		return basePkg
	}
	relPkg := strings.Join(strings.Split(rel, string(os.PathSeparator)), ".")
	if basePkg == "" {
		return relPkg
	}
	return basePkg + "." + relPkg
}

func getDataV1(tmplData *tmplData) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := tmplV1.Execute(buffer, tmplData); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func getDataV2(tmplData *tmplData) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := tmplV2.Execute(buffer, tmplData); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
