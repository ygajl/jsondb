package action

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/ygajl/jsondb/db"
)

func IsArray(mapInput map[string]interface{}) bool {
	if len(mapInput) == 1 {
		for k := range mapInput {
			if b := strings.HasSuffix(k, KEY_ARRAY); b {
				return true
			}
		}
	}
	return false
}

type TagArray struct {
	wCount int16
	wPage  int16
	wTotal int16
	tObj   TagObjReieve

	strTableNameJson string
	strWhere         string
	strOrder         string
	mapObj           map[string]TagObjReieve
	strJoin          string

	mapRet map[string]interface{}
}

func NewArray() TagArray {
	var tArray TagArray
	tArray.mapObj = make(map[string]TagObjReieve)
	tArray.mapRet = make(map[string]interface{})
	return tArray
}

func (t *TagArray) Init() {
	t.mapObj = make(map[string]TagObjReieve)
	t.mapRet = make(map[string]interface{})
}

func (t *TagArray) GetMapRet() map[string]interface{} {
	return t.mapRet
}

func (t *TagArray) ParseObj(mapInput map[string]interface{}) error {
	t.wTotal = -1
	for strKey, interValue := range mapInput {
		_itnerValue, bMap := interValue.(map[string]interface{})
		if !bMap {
			return errors.New(strKey + "对象错误!")
		}
		t.strTableNameJson = strKey
		n := strings.Index(strKey, KEY_ARRAY)
		strKeyReal := string([]byte(strKey)[0:n])

		for strKeyObj, interValueObj := range _itnerValue {
			if strKeyObj == KEY_COUNT {
				_itnerValue, err := interValueObj.(float64)
				if !err {
					return errors.New("count值错误")
				}
				t.wCount = int16(_itnerValue)
			} else if strKeyObj == KEY_PAGE {
				_itnerValue, err := interValueObj.(float64)
				if !err {
					return errors.New("page值错误")
				}
				t.wPage = int16(_itnerValue)
			} else if strKeyObj == KEY_JOIN {
				_itnerValue, err := interValueObj.(string)
				if !err {
					return errors.New("join值错误")
				}
				t.strJoin = string(_itnerValue)
			} else if strKeyObj == KEY_TOTAL {
				t.wTotal = 0
			} else if strKeyObj == strKeyReal || strKey == KEY_ARRAY {
				tObj := NewObj()
				mapObj := make(map[string]interface{})
				mapObj[strKeyObj] = interValueObj
				if err := tObj.ParseObj(mapObj); err == nil {
					t.tObj = tObj
				} else {
					return err
				}
			} else {
				tObj := NewObj()
				mapObj := make(map[string]interface{})
				mapObj[strKeyObj] = interValueObj
				if err := tObj.ParseObj(mapObj); err == nil {
					t.mapObj[strKeyObj] = tObj
				} else {
					return err
				}
			}
		}

		if strKey == KEY_ARRAY && len(t.mapObj) > 0 {
			return errors.New("[]存在多个对象")
		}
	}
	return nil
}

func (t *TagArray) GenerateRawSql() (string, error) {
	if t.strJoin != "" {
		arrJ := strings.Split(t.strJoin, ",")
		if len(arrJ) <= 0 {
			return "", errors.New("join值错误")
		}
		compile := regexp.MustCompile(`([&><|!])?/(.+)/(.+)@`)
		submatch := compile.FindAllSubmatch([]byte(arrJ[0]), -1)
		strTable1 := string(submatch[0][2])
		strKey1 := string(submatch[0][3])

		table1, _ := TMgrTable.getTableNameSqlbyJson(strTable1)
		colName1, _ := TMgrTable.getColSqlbyJson(strTable1, strKey1)

		submatch2 := compile.FindAllSubmatch([]byte(arrJ[1]), -1)
		strTable2 := string(submatch2[0][2])
		strKey2 := string(submatch2[0][3])
		table2, _ := TMgrTable.getTableNameSqlbyJson(strTable2)
		colName2, _ := TMgrTable.getColSqlbyJson(strTable2, strKey2)
		t.strWhere += ` join ` + table2 + ` on ` + table1 + `.` + colName1 + `=` + table2 + `.` + colName2

		//
		for strKey, tObjValue := range t.mapObj {
			_ = strKey
			if t.tObj.strColumnsSql != "" {
				t.tObj.strColumnsSql += ","
			}
			t.tObj.strColumnsSql += tObjValue.strColumnsSql
		}
	}

	t.strWhere += " " + t.tObj.strWhere
	t.strOrder = t.tObj.strOrder
	strSql := "select " + t.tObj.strColumnsSql + " from " + t.tObj.strTableNameSql + t.strWhere + t.strOrder

	if t.wCount != 0 {
		strSql += " limit " + strconv.Itoa(int(t.wCount))
	}
	return strSql, nil
}

func (t *TagArray) ExecRawSql(s string) error {
	sliRet := make([]interface{}, len(t.tObj.sliColumnsJson))
	lst := TMgrTable.mapJson2Data[t.tObj.strTableNameJson].lstColInfo
	for i := 0; i < len(t.tObj.sliColumnsJson); i++ {
		flag := false
		for p := lst.Front(); p != nil; p = p.Next() {
			if p.Value.(tagColInfo).strColJson == t.tObj.sliColumnsJson[i] {
				sliRet[i] = p.Value.(tagColInfo).interVal
				flag = true
				continue
			}
		}
		if !flag {
			sliRet[i] = new(interface{})
		}
	}

	if t.strJoin != "" {
		for _, objValue := range t.mapObj {
			_sliRet := make([]interface{}, len(objValue.sliColumnsJson))
			lst := TMgrTable.mapJson2Data[objValue.strTableNameJson].lstColInfo
			for i := 0; i < len(objValue.sliColumnsJson); i++ {
				flag := false
				for p := lst.Front(); p != nil; p = p.Next() {
					if p.Value.(tagColInfo).strColJson == objValue.sliColumnsJson[i] {
						_sliRet[i] = p.Value.(tagColInfo).interVal
						flag = true
						continue
					}
				}
				if !flag {
					_sliRet[i] = new(interface{})
				}
			}
			sliRet = append(sliRet, _sliRet...)
		}
	}

	sliOut, _ := db.Query(s, sliRet)
	sliMap := make([]interface{}, len(sliOut))
	for i := 0; i < len(sliOut); i++ {
		n := 0
		mapObj := make(map[string]interface{})
		{
			mapRetObj := make(map[string]interface{})
			for j := 0; j < len(t.tObj.sliColumnsJson); j++ {
				mapRetObj[t.tObj.sliColumnsJson[j]] = sliOut[i].([]interface{})[n]
				n++
			}
			mapObj[t.tObj.strTableNameJson] = mapRetObj
		}
		if t.strJoin != "" {
			for _, objValue := range t.mapObj {
				mapRetObj := make(map[string]interface{})
				for j := 0; j < len(objValue.sliColumnsJson); j++ {
					mapRetObj[objValue.sliColumnsJson[j]] = sliOut[i].([]interface{})[n]
					n++
				}
				mapObj[objValue.strTableNameJson] = mapRetObj
			}
		}
		sliMap[i] = mapObj
	}
	t.mapRet[t.strTableNameJson] = sliMap
	if t.wTotal == 0 {
		t.wTotal = int16(len(sliMap))
		t.mapRet["total"] = t.wTotal
	}
	return nil
}
