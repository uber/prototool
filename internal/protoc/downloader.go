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

package protoc

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/vars"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	fileLockRetryDelay = 250 * time.Millisecond
	fileLockTimeout    = 10 * time.Second
)

type downloader struct {
	lock sync.RWMutex

	logger    *zap.Logger
	cachePath string
	protocURL string
	config    settings.Config

	// the looked-up and verified to exist base path
	cachedBasePath string

	// If set, Prototool will invoke protoc and include
	// the well-known-types, from the configured binPath
	// and wktPath.
	protocBinPath string
	protocWKTPath string
}

func newDownloader(config settings.Config, options ...DownloaderOption) (*downloader, error) {
	downloader := &downloader{
		config: config,
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(downloader)
	}
	if downloader.config.Compile.ProtobufVersion == "" {
		downloader.config.Compile.ProtobufVersion = vars.DefaultProtocVersion
	}
	if downloader.protocBinPath != "" || downloader.protocWKTPath != "" {
		if downloader.protocURL != "" {
			return nil, fmt.Errorf("cannot use protoc-url in combination with either protoc-bin-path or protoc-wkt-path")
		}
		if downloader.protocBinPath == "" || downloader.protocWKTPath == "" {
			return nil, fmt.Errorf("both protoc-bin-path and protoc-wkt-path must be set")
		}
		cleanBinPath := filepath.Clean(downloader.protocBinPath)
		if _, err := os.Stat(cleanBinPath); os.IsNotExist(err) {
			return nil, err
		}
		cleanWKTPath := filepath.Clean(downloader.protocWKTPath)
		if _, err := os.Stat(cleanWKTPath); os.IsNotExist(err) {
			return nil, err
		}
		protobufPath := filepath.Join(cleanWKTPath, "google", "protobuf")
		info, err := os.Stat(protobufPath)
		if os.IsNotExist(err) {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%q is not a valid well-known types directory", protobufPath)
		}
		downloader.protocBinPath = cleanBinPath
		downloader.protocWKTPath = cleanWKTPath
	}
	return downloader, nil
}

func (d *downloader) Download() (string, error) {
	d.lock.RLock()
	cachedBasePath := d.cachedBasePath
	d.lock.RUnlock()
	if cachedBasePath != "" {
		return cachedBasePath, nil
	}
	return d.cache()
}

func (d *downloader) ProtocPath() (string, error) {
	if d.protocBinPath != "" {
		return d.protocBinPath, nil
	}
	basePath, err := d.Download()
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, "bin", "protoc"), nil
}

func (d *downloader) WellKnownTypesIncludePath() (string, error) {
	if d.protocWKTPath != "" {
		return d.protocWKTPath, nil
	}
	basePath, err := d.Download()
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, "include"), nil
}

func (d *downloader) Delete() error {
	basePath, err := d.getBasePathNoVersionOSARCH()
	if err != nil {
		return err
	}
	d.cachedBasePath = ""
	d.logger.Debug("deleting", zap.String("path", basePath))
	return os.RemoveAll(basePath)
}

func (d *downloader) cache() (_ string, retErr error) {
	if d.protocBinPath != "" {
		return d.protocBinPath, nil
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	basePath, err := d.getBasePath()
	if err != nil {
		return "", err
	}

	lock, err := newFlock(basePath)
	if err != nil {
		return "", err
	}
	if err := flockLock(lock); err != nil {
		return "", err
	}
	defer func() { retErr = multierr.Append(retErr, flockUnlock(lock)) }()

	if err := d.checkDownloaded(basePath); err != nil {
		if err := d.download(basePath); err != nil {
			return "", err
		}
		if err := d.checkDownloaded(basePath); err != nil {
			return "", err
		}
		d.logger.Debug("protobuf downloaded", zap.String("path", basePath))
	} else {
		d.logger.Debug("protobuf already downloaded", zap.String("path", basePath))
	}

	d.cachedBasePath = basePath
	return basePath, nil
}

func (d *downloader) checkDownloaded(basePath string) error {
	buffer := bytes.NewBuffer(nil)
	cmd := exec.Command(filepath.Join(basePath, "bin", "protoc"), "--version")
	cmd.Stdout = buffer
	if err := cmd.Run(); err != nil {
		return err
	}
	if d.protocURL != "" {
		// skip version check since we do not know the version
		return nil
	}
	output := strings.TrimSpace(buffer.String())
	d.logger.Debug("output from protoc --version", zap.String("output", output))
	expected := fmt.Sprintf("libprotoc %s", d.config.Compile.ProtobufVersion)
	if output != expected {
		return fmt.Errorf("expected %s from protoc --version, got %s", expected, output)
	}
	return nil
}

func (d *downloader) download(basePath string) (retErr error) {
	return d.downloadInternal(basePath, runtime.GOOS, runtime.GOARCH)
}

func (d *downloader) downloadInternal(basePath string, goos string, goarch string) (retErr error) {
	data, err := d.getDownloadData(goos, goarch)
	if err != nil {
		return err
	}
	// this is a working but hacky unzip
	// there must be a library for this
	// we don't properly copy directories, modification times, etc
	readerAt := bytes.NewReader(data)
	zipReader, err := zip.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return err
	}
	for _, file := range zipReader.File {
		fileMode := file.Mode()
		d.logger.Debug("found protobuf file in zip", zap.String("fileName", file.Name), zap.Any("fileMode", fileMode))
		if fileMode.IsDir() {
			continue
		}
		readCloser, err := file.Open()
		if err != nil {
			return err
		}
		defer func() {
			retErr = multierr.Append(retErr, readCloser.Close())
		}()
		fileData, err := ioutil.ReadAll(readCloser)
		if err != nil {
			return err
		}
		writeFilePath := filepath.Join(basePath, file.Name)
		if err := os.MkdirAll(filepath.Dir(writeFilePath), 0755); err != nil {
			return err
		}
		writeFile, err := os.OpenFile(writeFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}
		defer func() {
			retErr = multierr.Append(retErr, writeFile.Close())
		}()
		if _, err := writeFile.Write(fileData); err != nil {
			return err
		}
		d.logger.Debug("wrote protobuf file", zap.String("path", writeFilePath))
	}
	return nil
}

func (d *downloader) getDownloadData(goos string, goarch string) (_ []byte, retErr error) {
	url, err := d.getProtocURL(goos, goarch)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr == nil {
			d.logger.Debug("downloaded protobuf zip file", zap.String("url", url))
		}
	}()

	switch {
	case strings.HasPrefix(url, "file://"):
		return ioutil.ReadFile(strings.TrimPrefix(url, "file://"))
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		response, err := http.Get(url)
		if err != nil || response.StatusCode != http.StatusOK {
			// if there is not given protocURL, we tried to
			// download this from GitHub Releases, so add
			// extra context to the error message
			if d.protocURL == "" {
				return nil, fmt.Errorf("error downloading %s: %v\nMake sure GitHub Releases has a proper protoc zip file of the form protoc-VERSION-OS-ARCH.zip at https://github.com/protocolbuffers/protobuf/releases/v%s\nNote that many micro versions do not have this, and no version before 3.0.0-beta-2 has this", url, err, d.config.Compile.ProtobufVersion)
			}
			return nil, err
		}
		defer func() {
			if response.Body != nil {
				retErr = multierr.Append(retErr, response.Body.Close())
			}
		}()
		return ioutil.ReadAll(response.Body)
	default:
		return nil, fmt.Errorf("unknown url, can only handle http, https, file: %s", url)
	}

}

func (d *downloader) getProtocURL(goos string, goarch string) (string, error) {
	if d.protocURL != "" {
		return d.protocURL, nil
	}
	_, unameM, err := getUnameSUnameMPaths(goos, goarch)
	if err != nil {
		return "", err
	}
	protocS, err := getProtocSPath(goos)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s-%s.zip",
		d.config.Compile.ProtobufVersion,
		d.config.Compile.ProtobufVersion,
		protocS,
		unameM,
	), nil
}

func (d *downloader) getBasePath() (string, error) {
	basePathNoVersion, err := d.getBasePathNoVersion()
	if err != nil {
		return "", err
	}
	return filepath.Join(basePathNoVersion, d.getBasePathVersionPart()), nil
}

func (d *downloader) getBasePathNoVersionOSARCH() (string, error) {
	basePath := d.cachePath
	var err error
	if basePath == "" {
		basePath, err = getDefaultBasePathNoOSARCH()
		if err != nil {
			return "", err
		}
	} else {
		basePath, err = file.AbsClean(basePath)
		if err != nil {
			return "", err
		}
	}
	if err := file.CheckAbs(basePath); err != nil {
		return "", err
	}
	return basePath, nil
}

func (d *downloader) getBasePathNoVersion() (string, error) {
	basePath := d.cachePath
	var err error
	if basePath == "" {
		basePath, err = getDefaultBasePath()
		if err != nil {
			return "", err
		}
	} else {
		basePath, err = file.AbsClean(basePath)
		if err != nil {
			return "", err
		}
	}
	if err := file.CheckAbs(basePath); err != nil {
		return "", err
	}
	return filepath.Join(basePath, "protobuf"), nil
}

func (d *downloader) getBasePathVersionPart() string {
	if d.protocURL != "" {
		// we don't know the version or what is going on here
		hash := sha512.New()
		_, _ = hash.Write([]byte(d.protocURL))
		return base64.URLEncoding.EncodeToString(hash.Sum(nil))
	}
	return d.config.Compile.ProtobufVersion
}

func getDefaultBasePath() (string, error) {
	return getDefaultBasePathInternal(runtime.GOOS, runtime.GOARCH, os.Getenv)
}

func getDefaultBasePathInternal(goos string, goarch string, getenvFunc func(string) string) (string, error) {
	basePathNoOSARCH, err := getDefaultBasePathInternalNoOSARCH(goos, goarch, getenvFunc)
	if err != nil {
		return "", err
	}
	unameS, unameM, err := getUnameSUnameMPaths(goos, goarch)
	if err != nil {
		return "", err
	}
	return filepath.Join(basePathNoOSARCH, unameS, unameM), nil
}

func getDefaultBasePathNoOSARCH() (string, error) {
	return getDefaultBasePathInternalNoOSARCH(runtime.GOOS, runtime.GOARCH, os.Getenv)
}

func getDefaultBasePathInternalNoOSARCH(goos string, goarch string, getenvFunc func(string) string) (string, error) {
	unameS, _, err := getUnameSUnameMPaths(goos, goarch)
	if err != nil {
		return "", err
	}
	xdgCacheHome := getenvFunc("XDG_CACHE_HOME")
	if xdgCacheHome != "" {
		return filepath.Join(xdgCacheHome, "prototool"), nil
	}
	home := getenvFunc("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME is not set")
	}
	switch unameS {
	case "Darwin":
		return filepath.Join(home, "Library", "Caches", "prototool"), nil
	case "Linux":
		return filepath.Join(home, ".cache", "prototool"), nil
	default:
		return "", fmt.Errorf("invalid value for uname -s: %v", unameS)
	}
}

func getProtocSPath(goos string) (string, error) {
	switch goos {
	case "darwin":
		return "osx", nil
	case "linux":
		return "linux", nil
	default:
		return "", fmt.Errorf("unsupported value for runtime.GOOS: %v", goos)
	}
}

func getUnameSUnameMPaths(goos string, goarch string) (string, string, error) {
	var unameS string
	switch goos {
	case "darwin":
		unameS = "Darwin"
	case "linux":
		unameS = "Linux"
	default:
		return "", "", fmt.Errorf("unsupported value for runtime.GOOS: %v", goos)
	}
	var unameM string
	switch goarch {
	case "amd64":
		unameM = "x86_64"
	default:
		return "", "", fmt.Errorf("unsupported value for runtime.GOARCH: %v", goarch)
	}
	return unameS, unameM, nil
}

func newFlock(basePath string) (*flock.Flock, error) {
	fileLockPath := basePath + ".lock"
	// mkdir is atomic
	if err := os.MkdirAll(filepath.Dir(fileLockPath), 0755); err != nil {
		return nil, err
	}
	return flock.New(fileLockPath), nil
}

func flockLock(lock *flock.Flock) error {
	ctx, cancel := context.WithTimeout(context.Background(), fileLockTimeout)
	defer cancel()
	locked, err := lock.TryLockContext(ctx, fileLockRetryDelay)
	if err != nil {
		return fmt.Errorf("error acquiring file lock at %s - if you think this is in error, remove %s: %v", lock.Path(), lock.Path(), err)
	}
	if !locked {
		return fmt.Errorf("could not acquire file lock at %s after %v - if you think this is in error, remove %s", lock.Path(), fileLockTimeout, lock.Path())
	}
	return nil
}

func flockUnlock(lock *flock.Flock) error {
	if err := lock.Unlock(); err != nil {
		return fmt.Errorf("error unlocking file lock at %s: %v", lock.Path(), err)
	}
	return nil
}
