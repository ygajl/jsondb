package action

import (
	"errors"

	"github.com/ygajl/jsondb/db"
	"github.com/ygajl/jsondb/public"
)

type TagObjCreate struct {
	tagObjBase
	strKeyInsert   string
	strValueInsert string
}

func NewObjCreate() TagObjCreate {
	var tObj TagObjCreate
	tObj.mapRet = make(map[string]interface{})
	return tObj
}

func (t *TagObjCreate) Init() {
	t.mapRet = make(map[string]interface{})
}

func (t *TagObjCreate) GenerateRawSql() (string, error) {
	strSql := "insert into " + t.strTableNameSql + "(" + t.strKeyInsert + ")" + " values" + "(" + t.strValueInsert + ")" + " returning " + t.strColumnsSql
	return strSql, nil
}

func (t *TagObjCreate) ExecRawSql(s string) error {
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

func (t *TagObjCreate) ParseObj(mapInput map[string]interface{}) error {
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

		//
		for strKey, interValue := range mapInput {
			if strKey == KEY_COLUMN {
				continue
			} else if strKey == KEY_SET {
				continue
			} else {
				if err := t.getDefaultWhereSQL(strKey, interValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *TagObjCreate) getDefaultWhereSQL(strkey string, interValue interface{}) error {
	strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strkey)
	if err != nil {
		return errors.New("找不到对应的列:" + strkey)
	}

	if t.strKeyInsert != "" {
		t.strKeyInsert += ","
	}
	if t.strValueInsert != "" {
		t.strValueInsert += ","
	}
	t.strKeyInsert += strColSql
	t.strValueInsert += public.I2S(interValue)
	return nil
}
