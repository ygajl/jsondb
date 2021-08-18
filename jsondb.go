package jsondb

import (
	"github.com/ygajl/jsondb/action"
	"github.com/ygajl/jsondb/db"
)

type Parser interface {
	Init()
	ParseObj(mapInput map[string]interface{}) error
	GenerateRawSql() (string, error)
	ExecRawSql(s string) error
	GetMapRet() map[string]interface{}
}

func Connect(strDsn string) error {
	return db.Connect(strDsn)
}

func ParserJson(opt string, mapInput map[string]interface{}) (mapOut map[string]interface{}, err error) {
	var parser Parser
	defer func() {
		var mapRet map[string]interface{}
		if err != nil {
			mapRet = make(map[string]interface{})
			mapRet[action.KEY_CODE] = action.CODE_SERVER_ERROR
			mapRet[action.KEY_MSG] = err.Error()
		} else {
			mapRet = parser.GetMapRet()
			mapRet[action.KEY_CODE] = action.CODE_SUCCESS
			mapRet[action.KEY_MSG] = action.MSG_SUCCEED
		}
		mapOut = mapRet
	}()

	parser = getParser(opt, mapInput)
	if parser == nil {
		return map[string]interface{}{}, nil
	} else {
		parser.Init()
	}

	if err := parser.ParseObj(mapInput); err != nil {
		return map[string]interface{}{}, err
	}
	//
	if s, err := parser.GenerateRawSql(); err != nil {
		return map[string]interface{}{}, err
	} else {
		if err := parser.ExecRawSql(s); err != nil {
			return map[string]interface{}{}, err
		}
	}
	return map[string]interface{}{}, nil
}

func getParser(opt string, mapInput map[string]interface{}) Parser {
	switch opt {
	case "get":
		if action.IsArray(mapInput) {
			return new(action.TagArray)
		} else {
			return new(action.TagObjReieve)
		}
	case "post":
		return new(action.TagObjCreate)
	case "put":
		return new(action.TagObjUpdate)
	case "delete":
		return new(action.TagObjDelete)
	}
	return nil
}

func Register(strNameSql string, strNameJson string, inter interface{}) {
	action.Register(strNameSql, strNameJson, inter)
}
