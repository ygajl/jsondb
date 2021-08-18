package action

import (
	"errors"

	"github.com/ygajl/jsondb/db"
)

type TagObjDelete struct {
	tagObjBase
}

func NewObjDelete() TagObjDelete {
	var tObj TagObjDelete
	tObj.mapWhere = make(map[string]string)
	tObj.mapRet = make(map[string]interface{})
	return tObj
}

func (t *TagObjDelete) Init() {
	t.mapWhere = make(map[string]string)
	t.mapRet = make(map[string]interface{})
}

func (t *TagObjDelete) GenerateRawSql() (string, error) {
	strSql := "delete from " + t.strTableNameSql + t.strWhere + " returning " + t.strColumnsSql
	return strSql, nil
}

func (t *TagObjDelete) ExecRawSql(s string) error {
	sliCol := make([]interface{}, len(t.sliColumnsJson))
	lst := TMgrTable.mapJson2Data[t.strTableNameJson].lstColInfo
	for i := 0; i < len(t.sliColumnsJson); i++ {
		flag := false
		for p := lst.Front(); p != nil; p = p.Next() {
			if p.Value.(tagColInfo).strColJson == t.sliColumnsJson[i] {
				sliCol[i] = p.Value.(tagColInfo).interVal
				flag = true
				continue
			}
		}
		if !flag {
			sliCol[i] = new(interface{})
		}
	}

	mapRetObj := make(map[string]interface{})
	sliOut, err := db.QueryOne(s, sliCol)
	if err == nil {
		for j := 0; j < len(t.sliColumnsJson); j++ {
			mapRetObj[t.sliColumnsJson[j]] = sliOut[j]
		}
	} else {
		return err
	}

	t.mapRet[t.strTableNameJson] = mapRetObj
	return nil
}

func (t *TagObjDelete) ParseObj(mapInput map[string]interface{}) error {
	if len(mapInput) < 1 {
		return errors.New("json对象错误")
	}
	for strTableNameJson, interIn := range mapInput {
		mapInput, bMap := interIn.(map[string]interface{})
		if !bMap {
			return errors.New("对象值错误:" + strTableNameJson)
		}
		if bExist, err := TMgrTable.IsExistTableByJson(strTableNameJson); !bExist {
			return err
		}

		if strTableNameSql, err := TMgrTable.getTableNameSqlbyJson(strTableNameJson); err == nil {
			t.strTableNameSql = strTableNameSql
			t.strTableNameJson = strTableNameJson
		} else {
			return err
		}

		// @column
		if err := t.parseColumn(mapInput); err != nil {
			return err
		}

		// where
		if err := t.parseWhere(mapInput); err != nil {
			return err
		}
	}
	return nil
}
