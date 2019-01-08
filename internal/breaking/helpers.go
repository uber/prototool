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

func forEachMessagePair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.Message,
		*extract.Message,
	) error,
) error {
	return forEachPackagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromPackage *extract.Package, toPackage *extract.Package) error {
			return forEachMessagePairRec(addFailure, fromPackage.MessageNameToMessage(), toPackage.MessageNameToMessage(), f)
		},
	)
}

func forEachMessagePairRec(
	addFailure func(*text.Failure),
	fromMessageNameToMessage map[string]*extract.Message,
	toMessageNameToMessage map[string]*extract.Message,
	f func(
		func(*text.Failure),
		*extract.Message,
		*extract.Message,
	) error,
) error {
	for fromMessageName, fromMessage := range fromMessageNameToMessage {
		if toMessage, ok := toMessageNameToMessage[fromMessageName]; ok {
			if err := f(addFailure, fromMessage, toMessage); err != nil {
				return err
			}
			if err := forEachMessagePairRec(addFailure, fromMessage.NestedMessageNameToMessage(), toMessage.NestedMessageNameToMessage(), f); err != nil {
				return err
			}
		}
	}
	return nil
}

func forEachMessageFieldPair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.MessageField,
		*extract.MessageField,
	) error,
) error {
	return forEachMessagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromMessage *extract.Message, toMessage *extract.Message) error {
			fromFieldNumberToField := fromMessage.FieldNumberToField()
			toFieldNumberToField := toMessage.FieldNumberToField()
			for fromFieldNumber, fromField := range fromFieldNumberToField {
				if toField, ok := toFieldNumberToField[fromFieldNumber]; ok {
					if err := f(addFailure, fromField, toField); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func forEachMessageOneofPair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.MessageOneof,
		*extract.MessageOneof,
	) error,
) error {
	return forEachMessagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromMessage *extract.Message, toMessage *extract.Message) error {
			fromOneofNameToOneof := fromMessage.OneofNameToOneof()
			toOneofNameToOneof := toMessage.OneofNameToOneof()
			for fromOneofName, fromOneof := range fromOneofNameToOneof {
				if toOneof, ok := toOneofNameToOneof[fromOneofName]; ok {
					if err := f(addFailure, fromOneof, toOneof); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func forEachEnumPair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.Enum,
		*extract.Enum,
	) error,
) error {
	if err := forEachPackagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromPackage *extract.Package, toPackage *extract.Package) error {
			return forEachEnumPairInternal(addFailure, fromPackage.EnumNameToEnum(), toPackage.EnumNameToEnum(), f)
		},
	); err != nil {
		return err
	}
	return forEachMessagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromMessage *extract.Message, toMessage *extract.Message) error {
			return forEachEnumPairInternal(addFailure, fromMessage.NestedEnumNameToEnum(), toMessage.NestedEnumNameToEnum(), f)
		},
	)
}

func forEachEnumPairInternal(
	addFailure func(*text.Failure),
	fromEnumNameToEnum map[string]*extract.Enum,
	toEnumNameToEnum map[string]*extract.Enum,
	f func(
		func(*text.Failure),
		*extract.Enum,
		*extract.Enum,
	) error,
) error {
	for fromEnumName, fromEnum := range fromEnumNameToEnum {
		if toEnum, ok := toEnumNameToEnum[fromEnumName]; ok {
			if err := f(addFailure, fromEnum, toEnum); err != nil {
				return err
			}
		}
	}
	return nil
}

func forEachEnumValuePair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.EnumValue,
		*extract.EnumValue,
	) error,
) error {
	return forEachEnumPair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromEnum *extract.Enum, toEnum *extract.Enum) error {
			fromValueNumberToValue := fromEnum.ValueNumberToValue()
			toValueNumberToValue := toEnum.ValueNumberToValue()
			for fromValueNumber, fromValue := range fromValueNumberToValue {
				if toValue, ok := toValueNumberToValue[fromValueNumber]; ok {
					if err := f(addFailure, fromValue, toValue); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func forEachServicePair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.Service,
		*extract.Service,
	) error,
) error {
	return forEachPackagePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromPackage *extract.Package, toPackage *extract.Package) error {
			fromServiceNameToService := fromPackage.ServiceNameToService()
			toServiceNameToService := toPackage.ServiceNameToService()
			for fromServiceName, fromService := range fromServiceNameToService {
				if toService, ok := toServiceNameToService[fromServiceName]; ok {
					if err := f(addFailure, fromService, toService); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func forEachServiceMethodPair(
	addFailure func(*text.Failure),
	from *extract.PackageSet,
	to *extract.PackageSet,
	f func(
		func(*text.Failure),
		*extract.ServiceMethod,
		*extract.ServiceMethod,
	) error,
) error {
	return forEachServicePair(
		addFailure,
		from,
		to,
		func(addFailure func(*text.Failure), fromService *extract.Service, toService *extract.Service) error {
			fromMethodNameToMethod := fromService.MethodNameToMethod()
			toMethodNameToMethod := toService.MethodNameToMethod()
			for fromMethodName, fromMethod := range fromMethodNameToMethod {
				if toMethod, ok := toMethodNameToMethod[fromMethodName]; ok {
					if err := f(addFailure, fromMethod, toMethod); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func newTextFailuref(format string, args ...interface{}) *text.Failure {
	return &text.Failure{
		Message: fmt.Sprintf(format, args...),
	}
}
