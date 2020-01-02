package rbac

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	namespaceParsers = map[string]namespaceParser{
		"project": projectNamespaceParser,
	}
)

type namespaceParser func(resource Resource) (Namespace, error)

func projectNamespaceParser(resource Resource) (Namespace, error) {
	parserRe := regexp.MustCompile("^/project/([^/]*)/?")

	matches := parserRe.FindStringSubmatch(resource.String())

	if len(matches) <= 1 {
		return nil, errors.New("not support resource")
	}

	projectID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &projectNamespace{projectID: projectID}, nil
}
