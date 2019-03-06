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

// Package git provides git helper functionality.
package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/uber/prototool/internal/file"
	"go.uber.org/zap"
)

// TemporaryClone clones the git directory at the given dirPath into a temporary directory
//
// It is safe to os.RemoveAll this directory.
//
// If branchOrTag is not empty, this specific branch or tag will be cloned.
func TemporaryClone(logger *zap.Logger, dirPath string, branchOrTag string) (string, error) {
	absDirPath, err := file.AbsClean(dirPath)
	if err != nil {
		return "", err
	}
	fileInfo, err := os.Stat(filepath.Join(absDirPath, ".git"))
	if err != nil {
		return "", err
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("%q does not contain a .git directory", absDirPath)
	}
	cloneDirPath, err := ioutil.TempDir("", "prototool")
	if err != nil {
		return "", err
	}
	args := []string{"clone", "--depth", "1"}
	if branchOrTag != "" {
		args = append(args, "--branch", branchOrTag)
	}
	args = append(args, fmt.Sprintf("file://%s", absDirPath), cloneDirPath)
	logger.Sugar().Debugf("git %s", strings.Join(args, " "))
	output, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s had error: %s", strings.Join(args, " "), string(output))
	}
	return cloneDirPath, nil
}
