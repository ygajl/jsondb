package action

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ygajl/jsondb/db"
)

type TagObjReieve struct {
	tagObjBase
	// order
	strOrder string
	mapOrder map[string]string
}

func NewObj() TagObjReieve {
	var tObj TagObjReieve
	tObj.mapOrder = make(map[string]string)
	tObj.mapWhere = make(map[string]string)
	tObj.mapRet = make(map[string]interface{})
	return tObj
}

func (t *TagObjReieve) Init() {
	t.mapOrder = make(map[string]string)
	t.mapWhere = make(map[string]string)
	t.mapRet = make(map[string]interface{})
}

func (t *TagObjReieve) GenerateRawSql() (string, error) {
	strSql := "select " + t.strColumnsSql + " from " + t.strTableNameSql + t.strWhere + t.strOrder
	strSql += " limit " + strconv.Itoa(1)
	return strSql, nil
}

func (t *TagObjReieve) ExecRawSql(s string) error {
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

func (t *TagObjReieve) ParseObj(mapInput map[string]interface{}) error {
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
		// @order
		if err := t.parseOrder(mapInput); err != nil {
			return err
		}
	}

	return nil
}

func (t *TagObjReieve) parseOrder(mapInput map[string]interface{}) error {
	if _, ok := mapInput[KEY_ORDER]; ok {
		strOrder, bError := mapInput[KEY_ORDER].(string)
		if !bError {
			return errors.New("order值错误")
		}
		strOrder = strings.Replace(strOrder, " ", "", -1)
		sliOrder := strings.Split(strOrder, ",")
		for i := 0; i < len(sliOrder); i++ {
			if strings.HasSuffix(sliOrder[i], "-") {
				strKeyReal := string([]byte(sliOrder[i])[0 : len(sliOrder[i])-1])
				t.mapOrder[strKeyReal] = "desc"
			} else if strings.HasSuffix(sliOrder[i], "+") {
				strKeyReal := string([]byte(sliOrder[i])[0 : len(sliOrder[i])-1])
				t.mapOrder[strKeyReal] = "asc"
			} else {
				t.mapOrder[sliOrder[i]] = "asc"
			}
		}
	}

	t.strOrder = ""
	for strColNameJson, strValue := range t.mapOrder {
		if t.strOrder != "" {
			t.strOrder += ","
		}
		if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
			t.strOrder += t.strTableNameSql + `.` + strColSql + " " + strValue
		} else {
			return errors.New("找不到对应的列:" + strColNameJson)
		}
	}
	if t.strOrder != "" {
		t.strOrder = " order by " + t.strOrder
	}
	return nil
}
