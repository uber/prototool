// Copyright (c) 2018 Uber Technologies, Inc.
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

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/uber/prototool/internal/settings"
	"go.uber.org/zap"
)

type protoSetProvider struct {
	logger         *zap.Logger
	walkTimeout    time.Duration
	configProvider settings.ConfigProvider
}

func newProtoSetProvider(options ...ProtoSetProviderOption) *protoSetProvider {
	protoSetProvider := &protoSetProvider{
		logger:      zap.NewNop(),
		walkTimeout: DefaultWalkTimeout,
	}
	for _, option := range options {
		option(protoSetProvider)
	}
	protoSetProvider.configProvider = settings.NewConfigProvider(
		settings.ConfigProviderWithLogger(protoSetProvider.logger),
	)
	return protoSetProvider
}

func (c *protoSetProvider) GetForDir(workDirPath string, dirPath string) ([]*ProtoSet, error) {
	workDirPath, err := absClean(workDirPath)
	if err != nil {
		return nil, err
	}
	absDirPath, err := absClean(dirPath)
	if err != nil {
		return nil, err
	}
	configFilePath, err := c.configProvider.GetFilePathForDir(absDirPath)
	if err != nil {
		return nil, err
	}
	// we need everything for generation, not just the files in the given directory
	// so we go back to the config file if it is shallower
	// display path will be unaffected as this is based on workDirPath
	configDirPath := absDirPath
	if configFilePath != "" {
		configDirPath = filepath.Dir(configFilePath)
	}

	protoFiles, err := c.walkAndGetAllProtoFiles(workDirPath, configDirPath)
	if err != nil {
		return nil, err
	}
	dirPathToProtoFiles := getDirPathToProtoFiles(protoFiles)
	protoSets, err := c.getBaseProtoSets(dirPathToProtoFiles)
	if err != nil {
		return nil, err
	}
	for _, protoSet := range protoSets {
		protoSet.WorkDirPath = workDirPath
		protoSet.DirPath = absDirPath
	}
	c.logger.Debug("returning ProtoSets", zap.String("workDirPath", workDirPath), zap.String("dirPath", dirPath), zap.Any("protoSets", protoSets))
	return protoSets, nil
}

func (c *protoSetProvider) GetForFiles(workDirPath string, filePaths ...string) ([]*ProtoSet, error) {
	workDirPath, err := absClean(workDirPath)
	if err != nil {
		return nil, err
	}
	protoFiles, err := getProtoFiles(filePaths)
	if err != nil {
		return nil, err
	}
	dirPathToProtoFiles := getDirPathToProtoFiles(protoFiles)
	protoSets, err := c.getBaseProtoSets(dirPathToProtoFiles)
	if err != nil {
		return nil, err
	}
	for _, protoSet := range protoSets {
		protoSet.WorkDirPath = workDirPath
		protoSet.DirPath = workDirPath
	}
	c.logger.Debug("returning ProtoSets", zap.String("workDirPath", workDirPath), zap.Strings("filePaths", filePaths), zap.Any("protoSets", protoSets))
	return protoSets, nil
}

func (c *protoSetProvider) getBaseProtoSets(dirPathToProtoFiles map[string][]*ProtoFile) ([]*ProtoSet, error) {
	// TODO: this should be handled elsewhere
	if len(dirPathToProtoFiles) == 0 {
		return nil, nil
	}
	filePathToProtoSet := make(map[string]*ProtoSet)
	for dirPath, protoFiles := range dirPathToProtoFiles {
		configFilePath, err := c.configProvider.GetFilePathForDir(dirPath)
		if err != nil {
			return nil, err
		}
		protoSet, ok := filePathToProtoSet[configFilePath]
		if !ok {
			protoSet = &ProtoSet{
				DirPathToFiles: make(map[string][]*ProtoFile),
			}
			filePathToProtoSet[configFilePath] = protoSet
		}
		protoSet.DirPathToFiles[dirPath] = append(protoSet.DirPathToFiles[dirPath], protoFiles...)
		var config settings.Config
		if configFilePath != "" {
			config, err = c.configProvider.Get(configFilePath)
			if err != nil {
				return nil, err
			}
		}
		protoSet.Config = config
	}
	protoSets := make([]*ProtoSet, 0, len(filePathToProtoSet))
	for _, protoSet := range filePathToProtoSet {
		protoSets = append(protoSets, protoSet)
	}
	sort.Slice(protoSets, func(i int, j int) bool {
		return protoSets[i].Config.DirPath < protoSets[j].Config.DirPath
	})
	return protoSets, nil
}

func (c *protoSetProvider) walkAndGetAllProtoFiles(workDirPath string, dirPath string) ([]*ProtoFile, error) {
	var protoFiles []*ProtoFile
	absWorkDirPath, err := absClean(workDirPath)
	if err != nil {
		return nil, err
	}
	absDirPath, err := absClean(dirPath)
	if err != nil {
		return nil, err
	}
	allExcludePrefixes := make(map[string]struct{})
	numWalkedFiles := 0
	timedOut := false
	walkErrC := make(chan error)
	go func() {
		walkErrC <- filepath.Walk(
			absDirPath,
			func(filePath string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				numWalkedFiles++
				if timedOut {
					return fmt.Errorf("walking the diectory structure looking for proto files timed out after %v and having seen %d files, are you sure you are operating in the right context?", c.walkTimeout, numWalkedFiles)
				}
				absFilePath, err := absClean(filePath)
				if err != nil {
					return err
				}
				if fileInfo.IsDir() {
					excludePrefixes, err := c.configProvider.GetExcludePrefixesForDir(absFilePath)
					if err != nil {
						return err
					}
					for _, excludePrefix := range excludePrefixes {
						allExcludePrefixes[excludePrefix] = struct{}{}
					}
					for excludePrefix := range allExcludePrefixes {
						if strings.HasPrefix(absFilePath, excludePrefix) {
							return filepath.SkipDir
						}
					}
					return nil
				}
				if filepath.Ext(filePath) != ".proto" {
					return nil
				}
				for excludePrefix := range allExcludePrefixes {
					if strings.HasPrefix(absFilePath, excludePrefix) {
						return nil
					}
				}
				//displayPath := filePath
				//if !filepath.IsAbs(dirPath) {
				displayPath, err := filepath.Rel(absWorkDirPath, filePath)
				if err != nil {
					//return err
					displayPath = filePath
				}
				//}
				displayPath = filepath.Clean(displayPath)
				protoFiles = append(protoFiles, &ProtoFile{
					Path:        absFilePath,
					DisplayPath: displayPath,
				})
				return nil
			},
		)
	}()
	if c.walkTimeout == 0 {
		if walkErr := <-walkErrC; walkErr != nil {
			return nil, walkErr
		}
		return protoFiles, nil
	}
	select {
	case walkErr := <-walkErrC:
		if walkErr != nil {
			return nil, walkErr
		}
		return protoFiles, nil
	case <-time.After(c.walkTimeout):
		timedOut = true
		if walkErr := <-walkErrC; walkErr != nil {
			return nil, walkErr
		}
		// TODO(pedge): nice job with your code, I should learn how to code
		return nil, fmt.Errorf("should never get here")
	}
}

func getDirPathToProtoFiles(protoFiles []*ProtoFile) map[string][]*ProtoFile {
	dirPathToProtoFiles := make(map[string][]*ProtoFile)
	for _, protoFile := range protoFiles {
		dir := filepath.Dir(protoFile.Path)
		dirPathToProtoFiles[dir] = append(dirPathToProtoFiles[dir], protoFile)
	}
	return dirPathToProtoFiles
}

func getProtoFiles(filePaths []string) ([]*ProtoFile, error) {
	protoFiles := make([]*ProtoFile, 0, len(filePaths))
	for _, filePath := range filePaths {
		absFilePath, err := absClean(filePath)
		if err != nil {
			return nil, err
		}
		protoFiles = append(protoFiles, &ProtoFile{
			Path:        absFilePath,
			DisplayPath: filePath,
		})
	}
	return protoFiles, nil
}

func absClean(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if !filepath.IsAbs(path) {
		return filepath.Abs(path)
	}
	return filepath.Clean(path), nil
}
