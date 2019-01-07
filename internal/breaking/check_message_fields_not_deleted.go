package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/text"
)

func checkMessageFieldsNotDeleted(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	return forEachMessagePair(addFailure, from, to, checkMessageFieldsNotDeletedMessage)
}

func checkMessageFieldsNotDeletedMessage(addFailure func(*text.Failure), from *extract.Message, to *extract.Message) error {
	fromFieldNameToField := from.FieldNameToField()
	toFieldNameToField := to.FieldNameToField()
	for fieldName := range fromFieldNameToField {
		if _, ok := toFieldNameToField[fieldName]; !ok {
			addFailure(newMessageFieldsNotDeletedFailure(from.FullyQualifiedName(), fieldName))
		}
	}
	return nil
}

func newMessageFieldsNotDeletedFailure(messageName string, fieldName string) *text.Failure {
	return newTextFailuref(`Message field %q on message %q was deleted.`, fieldName, messageName)
}
