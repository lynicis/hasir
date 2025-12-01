package proto

import "buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

var VisibilityMap = map[shared.Visibility]Visibility{
	shared.Visibility_VISIBILITY_PUBLIC:  VisibilityPublic,
	shared.Visibility_VISIBILITY_PRIVATE: VisibilityPrivate,
}

var ReverseVisibilityMap = map[Visibility]shared.Visibility{
	VisibilityPublic:  shared.Visibility_VISIBILITY_PUBLIC,
	VisibilityPrivate: shared.Visibility_VISIBILITY_PRIVATE,
}
