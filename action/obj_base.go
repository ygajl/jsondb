package action

import (
	"errors"
	"regexp"
	"strings"

	"github.com/ygajl/jsondb/public"
)

type tagObjBase struct {
	// struct
	strTableNameJson string
	sliColumnsJson   []string
	strTableNameSql  string
	strColumnsSql    string
	bForeign         bool

	// where
	strWhere   string
	mapWhere   map[string]string
	sliCombine []string

	// join
	strJoin string

	// result
	mapRet map[string]interface{}
}

func NewObjBase() tagObjBase {
	var tObjBase tagObjBase
	tObjBase.bForeign = false
	tObjBase.mapRet = make(map[string]interface{})
	return tObjBase
}

func (t *tagObjBase) GetMapRet() map[string]interface{} {
	return t.mapRet
}

func (t *tagObjBase) getColumn(strTableNameJson string, columns []string) (string, error) {
	strColumns := ""
	for k, v := range columns {
		if k != 0 {
			strColumns += ","
		}
		if strColSql, err := TMgrTable.getColSqlbyJson(strTableNameJson, v); err == nil {
			strColSql = t.strTableNameSql + `.` + strColSql
			strColumns += strColSql
		} else {
			return strColumns, err
		}
	}
	return strColumns, nil
}

func (t *tagObjBase) parseColumn(mapInput map[string]interface{}) error {
	if _, ok := mapInput[KEY_COLUMN]; ok {
		if strColumns, ok := mapInput[KEY_COLUMN].(string); ok {
			strColumns = strings.Replace(strColumns, " ", "", -1)
			t.sliColumnsJson = strings.Split(strColumns, ",")
			if strColumnsSql, err := t.getColumn(t.strTableNameJson, t.sliColumnsJson); err == nil {
				t.strColumnsSql = strColumnsSql
			} else {
				return err
			}
		} else {
			return errors.New(KEY_COLUMN + "不合法")
		}
	} else {
		lst := TMgrTable.mapJson2Data[t.strTableNameJson].lstColInfo
		for p := lst.Front(); p != nil; p = p.Next() {
			if p != lst.Front() {
				t.strColumnsSql += ","
			}
			t.strColumnsSql += t.strTableNameSql + `.` + p.Value.(tagColInfo).strColSql
			t.sliColumnsJson = append(t.sliColumnsJson, p.Value.(tagColInfo).strColJson)
		}
	}

	return nil
}

func (t *tagObjBase) parseWhere(mapInput map[string]interface{}) error {
	for strKey, interValue := range mapInput {
		if strKey == KEY_COLUMN {
			continue
		} else if strKey == KEY_SET {
			continue
		} else if strKey == KEY_ORDER {
			continue
		} else if strKey == KEY_COMBINE {
			if strValue, ok := interValue.(string); ok {
				strValue = strings.Replace(strValue, " ", "", -1)
				t.sliCombine = strings.Split(strValue, ",")
			} else {
				return errors.New("combine值错误")
			}
		} else if b := strings.HasSuffix(strKey, KEY_CONTION); b {
			if err := t.getConditionQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_FOREIGN); b {
			t.bForeign = true
			if err := t.getJoinSQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_LIKE); b {
			if err := t.getLikeStringSQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_GREATER_EQUAL); b {
			if err := t.getCmpGreaterEqualSQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_GREATER); b {
			if err := t.getCmpGreaterSQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_LESS_EQUAL); b {
			if err := t.getCmpLessEqualSQL(strKey, interValue); err != nil {
				return err
			}
		} else if b := strings.HasSuffix(strKey, KEY_LESS); b {
			if err := t.getCmpLessSQL(strKey, interValue); err != nil {
				return err
			}
		} else {
			if err := t.getDefaultWhereSQL(strKey, interValue); err != nil {
				return err
			}
		}
	}

	//where and or not
	for strKey, strValue := range t.mapWhere {
		if t.strWhere != "" {
			bFind := false
			for i := 0; i < len(t.sliCombine); i++ {
				if t.sliCombine[i] == strKey {
					bFind = true
					break
				}
			}
			if bFind {
				t.strWhere += " or "
			} else {
				t.strWhere += " and "
			}
		} else {
			t.strWhere = " where "
		}
		t.strWhere += strValue
	}
	return nil
}

func (t *tagObjBase) getDefaultWhereSQL(strkey string, interValue interface{}) error {
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strkey); err == nil {
		strRet := t.strTableNameSql + `.` + strColSql + "=" + public.I2S(interValue)
		t.mapWhere[strkey] = strRet
	} else {
		return err
	}

	return nil
}

func (t *tagObjBase) getConditionQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_CONTION)
	if n < 0 {
		return nil
	}
	compile := regexp.MustCompile(`([^&|!]+)([&|!]?)` + KEY_CONTION)
	submatch := compile.FindAllSubmatch([]byte(strKey), -1)

	if len(submatch) <= 0 {
		return errors.New("join值错误:{}")
	}

	strColNameJson := string(submatch[0][1])
	strUnion := string(submatch[0][2])

	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		_ = strColSql
		if strCondition, ok := interValue.(string); ok {
			arrJ := strings.Split(strCondition, ",")
			for key, value := range arrJ {
				if key != 0 {
					if strUnion == "" || strUnion == "|" {
						t.mapWhere[strKey] += "or"
					} else if strUnion == "&" {
						t.mapWhere[strKey] += "and"
					}
				}
				compile := regexp.MustCompile(`(<=|>=|<|>)(.*)`)
				submatch := compile.FindAllSubmatch([]byte(value), -1)
				if len(submatch) > 0 {
					oper := string(submatch[0][1])
					strValue := string(submatch[0][2])
					t.mapWhere[strKey] = t.mapWhere[strKey] + t.strTableNameSql + `.` + strColSql + oper + public.I2S(strValue)
				}
			}
		} else if arrValue, ok := interValue.([]interface{}); ok {
			strRange := ""
			oper := ""
			for key, value := range arrValue {
				if key != 0 {
					strRange += ","
				}
				strRange += public.I2S(value)
			}

			if strUnion == "!" {
				oper = " not in "
			} else {
				oper = " in "
			}

			t.mapWhere[strKey] = t.strTableNameSql + `.` + strColSql + oper + "(" + strRange + ")"
		}
	} else {
		return err
	}

	return nil
}

func (t *tagObjBase) getJoinSQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_FOREIGN)
	if n < 0 {
		return nil
	}
	strColNameJson := string([]byte(strKey)[0:n])
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		if strJoin, ok := interValue.(string); ok {
			strJoin = strings.Replace(strJoin, " ", "", -1)
			arrJ := strings.Split(strJoin, ",")
			if len(arrJ) <= 0 {
				return errors.New("join值错误")
			}
			compile := regexp.MustCompile(`/(.+)/(.+)`)
			submatch := compile.FindAllSubmatch([]byte(arrJ[0]), -1)
			strTable1 := string(submatch[0][1])
			strKey1 := string(submatch[0][2])

			table1, _ := TMgrTable.getTableNameSqlbyJson(strTable1)
			colName1, _ := TMgrTable.getColSqlbyJson(strTable1, strKey1)

			t.strJoin += ` join ` + t.strTableNameSql + ` on ` + table1 + `.` + colName1 + `=` + t.strTableNameSql + `.` + strColSql
		} else {
			return errors.New("join值错误")
		}
	} else {
		return err
	}

	return nil
}

func (t *tagObjBase) getCmpGreaterEqualSQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_GREATER_EQUAL)
	if n < 0 {
		return nil
	}
	strColNameJson := string([]byte(strKey)[0:n])
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		t.mapWhere[strKey] = t.strTableNameSql + `.` + strColSql + ">=" + public.I2S(interValue)
	} else {
		return err
	}
	return nil
}

func (t *tagObjBase) getCmpGreaterSQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_GREATER)
	if n < 0 {
		return nil
	}
	strColNameJson := string([]byte(strKey)[0:n])
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		t.mapWhere[strKey] = t.strTableNameSql + `.` + strColSql + ">" + public.I2S(interValue)
	} else {
		return err
	}
	return nil
}

func (t *tagObjBase) getCmpLessEqualSQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_LESS_EQUAL)
	if n < 0 {
		return nil
	}
	strColNameJson := string([]byte(strKey)[0:n])
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		t.mapWhere[strKey] = t.strTableNameSql + `.` + strColSql + "<=" + public.I2S(interValue)
	} else {
		return err
	}
	return nil
}

func (t *tagObjBase) getCmpLessSQL(strKey string, interValue interface{}) error {
	n := strings.Index(strKey, KEY_LESS)
	if n < 0 {
		return nil
	}
	strColNameJson := string([]byte(strKey)[0:n])
	if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
		t.mapWhere[strKey] = t.strTableNameSql + `.` + strColSql + "<" + public.I2S(interValue)
	} else {
		return err
	}
	return nil
}

func (t *tagObjBase) getLikeStringSQL(strkey string, interValue interface{}) error {
	if strValue, ok := interValue.(string); ok {
		n := strings.Index(strkey, KEY_LIKE)
		strColNameJson := string([]byte(strkey)[0:n])
		if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
			t.mapWhere[strkey] = t.strTableNameSql + `.` + strColSql + ` like ` + `'` + strValue + `'`
		} else {
			return err
		}
	} else if sliValue, ok := interValue.([]interface{}); ok {
		n := strings.Index(strkey, KEY_LIKE)
		if n < 0 {
			return nil
		}
		strColNameJson := string([]byte(strkey)[0:n])
		strArr := "(array["
		if strColSql, err := TMgrTable.getColSqlbyJson(t.strTableNameJson, strColNameJson); err == nil {
			for k, v := range sliValue {
				if k != 0 {
					strArr += ","
				}
				strArr += `'` + v.(string) + `'`
			}
			strArr += "])"
			t.mapWhere[strkey] = t.strTableNameSql + `.` + strColSql + ` like any ` + strArr
		} else {
			return err
		}
	}
	return nil
}
