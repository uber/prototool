package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkPackagesNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	fromPackageNameToPackage := from.PackageNameToPackage()
	toPackageNameToPackage := to.PackageNameToPackage()
	for fromPackageName := range fromPackageNameToPackage {
		if _, ok := toPackageNameToPackage[fromPackageName]; !ok {
			addFailure(newPackagesNotDeletedFailure(fromPackageName))
		}
	}
	return nil
}

func newPackagesNotDeletedFailure(packageName string) *text.Failure {
	return newTextFailuref(`Package %q was deleted.`, packageName)
}
