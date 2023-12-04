package action

import (
	"errors"
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
	wCount     int16
	wPage      int16
	wTotal     int16
	tObj       TagObjReieve
	bJoin      bool
	strPrimary string

	strTableNameJson string
	strJoin          string
	strWhere         string
	strOrder         string
	mapObj           map[string]TagObjReieve
	mapObjOrder      []string

	mapRet map[string]interface{}
}

func NewArray() TagArray {
	var tArray TagArray
	tArray.bJoin = false
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
			} else if strKeyObj == KEY_TOTAL {
				t.wTotal = 0
			} else {
				tObj := NewObj()
				mapObj := make(map[string]interface{})
				mapObj[strKeyObj] = interValueObj

				if err := tObj.ParseObj(mapObj); err == nil {
					if tObj.bForeign == false {
						if t.tObj.strTableNameSql != "" {
							return errors.New("存在两个主对象")
						}
						t.tObj = tObj
					} else {
						t.mapObj[strKeyObj] = tObj
						t.bJoin = true
					}
				} else {
					return err
				}
			}
		}
	}

	for name := range t.mapObj {
		t.mapObjOrder = append(t.mapObjOrder, name)
	}

	return nil
}

func (t *TagArray) GenerateRawSql() (string, error) {
	strSelect := t.tObj.strColumnsSql
	t.strOrder = t.tObj.strOrder

	for _, name := range t.mapObjOrder {
		objValue := t.mapObj[name]
		if strSelect != "" {
			strSelect += ","
		}
		strSelect += objValue.strColumnsSql
		t.strJoin += objValue.strJoin
		t.strOrder += objValue.strOrder
	}

	t.strWhere += " " + t.tObj.strWhere
	strSql := "select " + strSelect + " from " + t.tObj.strTableNameSql + t.strJoin + t.strWhere + t.strOrder

	if t.wCount != 0 {
		if t.wPage != 0 {
			strSql += " limit " + strconv.Itoa(int(t.wCount)) + " offset " + strconv.Itoa(int(t.wPage*t.wCount))
		} else {
			strSql += " limit " + strconv.Itoa(int(t.wCount))
		}
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

	if t.bJoin {
		for _, name := range t.mapObjOrder {
			objValue := t.mapObj[name]
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
		if t.bJoin {
			for _, name := range t.mapObjOrder {
				objValue := t.mapObj[name]
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
