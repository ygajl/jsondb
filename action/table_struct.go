package action

import (
	"container/list"
	"errors"
	"reflect"
)

var (
	TMgrTable *tagMgrTable
)

type tagMgrTable struct {
	mapJson2Data map[string]tagData
}

type tagData struct {
	strTableNameSql string
	interStruct     interface{}
	lstColInfo      *list.List
}

type tagColInfo struct {
	strColJson string
	strColSql  string
	interVal   interface{} //reflect.ValueOf(*t).Elem().Field(i).Addr().Interface()
}

func init() {
	TMgrTable = NewMgrTable()
}

func NewMgrTable() *tagMgrTable {
	tMgrTable := new(tagMgrTable)
	tMgrTable.mapJson2Data = make(map[string]tagData)
	return tMgrTable
}

func Register(strNameSql string, strNameJson string, inter interface{}) {
	var tData tagData
	tData.strTableNameSql = strNameSql
	tData.strTableNameSql = `"` + tData.strTableNameSql + `"`
	tData.interStruct = inter
	lst := list.New()
	refType := reflect.TypeOf(inter).Elem()
	refValue := reflect.ValueOf(inter).Elem()
	for i := 0; i < refType.NumField(); i++ {
		var tColInfo tagColInfo
		field := refType.Field(i)
		tColInfo.strColJson = field.Tag.Get("json")
		tColInfo.strColSql = field.Tag.Get("sql")
		tColInfo.strColSql = `"` + tColInfo.strColSql + `"`
		tColInfo.interVal = refValue.Field(i).Addr().Interface()
		lst.PushBack(tColInfo)
	}
	tData.lstColInfo = lst
	TMgrTable.mapJson2Data[strNameJson] = tData
}

func (t *tagMgrTable) IsExistTableByJson(strTableNameJson string) (bool, error) {
	if _, ok := t.mapJson2Data[strTableNameJson]; !ok {
		return false, errors.New("数据库" + strTableNameJson + "不存在")
	}

	return true, nil
}

func (t *tagMgrTable) getTableNameSqlbyJson(strTableNameJson string) (string, error) {
	if bExist, _ := t.IsExistTableByJson(strTableNameJson); !bExist {
		return "", errors.New("数据库" + strTableNameJson + "不存在")
	}

	return t.mapJson2Data[strTableNameJson].strTableNameSql, nil
}

func (t *tagMgrTable) getColSqlbyJson(strTableNameJson string, strColJson string) (string, error) {
	if bExist, _ := t.IsExistTableByJson(strTableNameJson); !bExist {
		return "", errors.New("数据库" + strTableNameJson + "不存在")
	}

	for p := t.mapJson2Data[strTableNameJson].lstColInfo.Front(); p != nil; p = p.Next() {
		if _tRefValue, ok := p.Value.(tagColInfo); ok {
			if _tRefValue.strColJson == strColJson {
				return _tRefValue.strColSql, nil
			}
		}
	}
	return "", errors.New("找不到对应的列:" + strColJson)
}

// func (t *tagMgrTable) getColJsonbySql(strTableNameJson string, strColSql string) (string, error) {
// 	strTableNameJson = `"` + strTableNameJson + `"`
// 	if bExist, _ := t.IsExistTableByJson(strTableNameJson); !bExist {
// 		return "", errors.New("数据库" + strTableNameJson + "不存在")
// 	}

// 	for p := t.mapJson2Data[strTableNameJson].lstColInfo.Front(); p != nil; p = p.Next() {
// 		if _tRefValue, ok := p.Value.(tagColInfo); ok {
// 			if _tRefValue.strColJson == strColSql {
// 				return _tRefValue.strColJson, nil
// 			}
// 		}
// 	}
// 	return "", errors.New("找不到对应的列:" + strColSql)
// }
