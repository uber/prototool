package breaking

import (
	"fmt"

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

func checkMessagesNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachPackagePair(addFailure, from, to, checkMessagesNotDeletedPackage)
}

func checkMessagesNotDeletedPackage(addFailure func(*text.Failure), from *extract.Package, to *extract.Package) error {
	return nil
}

func checkMessageFieldsNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return nil
}

func checkMessageFieldsHaveSameType(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return nil
}

func forEachPackagePair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.Package,
		*extract.Package,
	) error,
) error {
	fromPackageNameToPackage := from.PackageNameToPackage()
	toPackageNameToPackage := to.PackageNameToPackage()
	for fromPackageName, fromPackage := range fromPackageNameToPackage {
		if toPackage, ok := toPackageNameToPackage[fromPackageName]; ok {
			if err := f(addFailure, fromPackage, toPackage); err != nil {
				return err
			}
		}
	}
	return nil
}

func newPackagesNotDeletedFailure(packageName string) *text.Failure {
	return newTextFailuref(`Package %q was deleted.`, packageName)
}

func newTextFailuref(format string, args ...interface{}) *text.Failure {
	return &text.Failure{
		Message: fmt.Sprintf(format, args...),
	}
}
