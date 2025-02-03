package types

import "github.com/lovoo/goka/codec"

type StopkwordInput struct {
	Key      string
	Replacer string
}
type StopkwordRemoveInput struct {
	Key string
}
type StopwordsValueCodec struct {
	codec.String
}
