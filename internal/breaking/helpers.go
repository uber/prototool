package breaking

import (
	"fmt"

	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

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

func joinFullyQualifiedName(fullyQualifiedName string, name string) string {
	if fullyQualifiedName == "" {
		return name
	}
	return fullyQualifiedName + "." + name
}

func newTextFailuref(format string, args ...interface{}) *text.Failure {
	return &text.Failure{
		Message: fmt.Sprintf(format, args...),
	}
}
