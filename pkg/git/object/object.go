package object

import (
	"encoding/hex"
	"fmt"
)

type Mode int

const (
	UnknownMode Mode = iota
	RegularFileMode
	ExecutableFileMode
	SymLinkMode
	GitLinkMode
	DirectoryMode
)

func (m Mode) String() string {
	switch m {
	case RegularFileMode:
		return "RegularFile"
	case ExecutableFileMode:
		return "ExecutableFile"
	case SymLinkMode:
		return "SymLinkFile"
	case GitLinkMode:
		return "GitLinkFile"
	case DirectoryMode:
		return "Directory"
	default:
		return "UnknownMode"
	}
}

func ParseMode(str string) (Mode, error) {
	switch str {
	case "100644":
		return RegularFileMode, nil
	case "100755":
		return ExecutableFileMode, nil
	case "120000":
		return SymLinkMode, nil
	case "160000":
		return GitLinkMode, nil
	case "040000":
		return DirectoryMode, nil
	default:
		return UnknownMode, fmt.Errorf("unknown object mode %s", str)
	}
}

type Type int

const (
	UnknownType Type = iota
	Blob
	Tree
	Commit
	Tag
)

func (t Type) String() string {
	switch t {
	case Blob:
		return "blob"
	case Tree:
		return "tree"
	case Commit:
		return "commit"
	case Tag:
		return "tag"
	default:
		return "UnknownType"
	}
}

func ParseType(typeString string) (Type, error) {
	switch typeString {
	case "blob":
		return Blob, nil
	case "tree":
		return Tree, nil
	case "commit":
		return Commit, nil
	case "tag":
		return Tag, nil
	default:
		return UnknownType, fmt.Errorf("unknown object type %s", typeString)
	}
}

type ID [20]byte

const Sha1HexLength = 40
const Sha1Length = 20

func ParseID(sha1hex string) (ID, error) {
	if len(sha1hex) != Sha1HexLength {
		return ID{}, fmt.Errorf("invalid sha1 hex %s， length=%d", sha1hex, len(sha1hex))
	}

	sha1bytes, err := hex.DecodeString(sha1hex)
	if err != nil {
		return ID{}, nil
	}
	return ID(sha1bytes), err
}

func (id ID) String() string {
	return hex.EncodeToString(id[:Sha1Length])
}
