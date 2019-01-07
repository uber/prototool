package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkMessagesNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return nil
}

func checkMessageFieldsNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return nil
}

func checkMessageFieldsHaveSameType(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return nil
}
