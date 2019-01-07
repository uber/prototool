package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkMessagesNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachPackagePair(addFailure, from, to, checkMessagesNotDeletedPackage)
}

func checkMessagesNotDeletedPackage(addFailure func(*text.Failure), from *extract.Package, to *extract.Package) error {
	return checkMessagesNotDeletedMap(addFailure, from.FullyQualifiedName(), from.MessageNameToMessage(), to.MessageNameToMessage())
}

func checkMessagesNotDeletedMap(addFailure func(*text.Failure), fullyQualifiedName string, from map[string]*extract.Message, to map[string]*extract.Message) error {
	for fromMessageName, fromMessage := range from {
		toMessage, ok := to[fromMessageName]
		if !ok {
			addFailure(newMessagesNotDeletedFailure(joinFullyQualifiedName(fullyQualifiedName, fromMessageName)))
		} else if err := checkMessagesNotDeletedMap(addFailure, fromMessage.FullyQualifiedName(), fromMessage.NestedMessageNameToMessage(), toMessage.NestedMessageNameToMessage()); err != nil {
			return err
		}
	}
	return nil
}

func newMessagesNotDeletedFailure(messageName string) *text.Failure {
	return newTextFailuref(`Message %q was deleted.`, messageName)
}
