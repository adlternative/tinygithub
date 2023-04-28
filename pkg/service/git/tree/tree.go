package tree

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/service/git/object"
	"strings"
)

type Entry struct {
	Mode object.Mode
	Type object.Type
	Oid  object.ID
	Path string
}

func Parse(treeLine string) (*Entry, error) {
	parts := strings.Split(treeLine, "\t")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tree line")
	}

	firstPart := parts[0]
	firstParts := strings.Fields(firstPart)
	if len(firstParts) != 3 {
		return nil, fmt.Errorf("invalid tree line")
	}

	mode, err := object.ParseMode(firstParts[0])
	if err != nil {
		return nil, err
	}

	parseType, err := object.ParseType(firstParts[1])
	if err != nil {
		return nil, err
	}

	oid, err := object.ParseID(firstParts[2])
	if err != nil {
		return nil, err
	}

	return &Entry{
		Mode: mode,
		Type: parseType,
		Oid:  oid,
		Path: parts[1],
	}, nil
}
