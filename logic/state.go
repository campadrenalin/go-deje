package logic

import (
	"errors"
	"github.com/campadrenalin/go-deje/model"
	"github.com/campadrenalin/go-deje/util"
)

type DocumentState struct {
	Version string
	Content model.JSONObject
}

func NewDocumentState() DocumentState {
	return DocumentState{
		Version: "",
		Content: make(model.JSONObject),
	}
}

func (ds DocumentState) GetProperty(name string, s interface{}) error {
	data, ok := ds.Content[name]
	if !ok {
		return errors.New("Document does not have " + name + " property")
	}

	return util.CloneMarshal(data, s)
}
