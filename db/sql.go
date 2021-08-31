package db

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

var pg *sqlx.DB

func Connect(strDsn string) error {
	db, err := sqlx.Open("postgres", strDsn)
	if err != nil {
		return err
	}
	pg = db
	pg.SetConnMaxLifetime(time.Minute * 3)
	pg.SetMaxOpenConns(10)
	pg.SetMaxIdleConns(10)

	err = pg.Ping()
	if err != nil {
		log.Println(err)
	}
	return err
}

func QueryOne(strRawSql string, sli []interface{}) ([]interface{}, error) {
	log.Println("sql:", strRawSql)
	row := pg.QueryRow(strRawSql)
	if row.Err() != nil {
		return nil, row.Err()
	}

	_sli := getNewStu(sli)
	err := row.Scan(_sli...)
	if err != nil {
		return nil, errors.New("错误" + err.Error())
	}

	for k, v := range _sli {
		if value, ok := v.(*interface{}); ok {
			a := *(*interface{})(unsafe.Pointer(value))
			if value2, ok := a.([]byte); ok {
				reqMap := make(map[string]interface{})
				s := `{` + `"a":` + string(value2) + `}`
				json.Unmarshal([]byte(s), &reqMap)
				_sli[k] = reqMap["a"]
			}
		}
	}

	return _sli, nil
}

func Query(strRawSql string, sli []interface{}) ([]interface{}, error) {
	log.Println("sql:", strRawSql)
	var sliOut []interface{}
	rows, err := pg.Query(strRawSql)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		_sli := getNewStu(sli)
		err := rows.Scan(_sli...)
		if err != nil {
			return nil, errors.New("错误" + err.Error())
		}
		for k, v := range _sli {
			if value, ok := v.(*interface{}); ok {
				a := *(*interface{})(unsafe.Pointer(value))
				if value2, ok := a.([]byte); ok {
					reqMap := make(map[string]interface{})
					s := `{` + `"a":` + string(value2) + `}`
					json.Unmarshal([]byte(s), &reqMap)
					_sli[k] = reqMap["a"]
				}
			}
		}
		sliOut = append(sliOut, _sli)
	}
	return sliOut, nil
}

func getNewStu(sli []interface{}) []interface{} {
	_sli := make([]interface{}, len(sli))
	for i := 0; i < len(sli); i++ {
		v := sli[i]
		_sli[i] = &v
	}
	return _sli
}
