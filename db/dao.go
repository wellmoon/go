package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"

	"github.com/wellmoon/go/logger"
	Log "github.com/wellmoon/go/logger"
	"github.com/wellmoon/go/zjson"
)

type ListResult struct {
	Columns []string            `json:"columns"`
	List    []map[string]string `json:"list"`
}

type Dao struct {
	db *sql.DB
}

func NewDao(ip string, port string, dbUser string, dbPass string, dbName string) (*Dao, error) {

	// err := Ping(ip, port, dbUser, dbPass, dbName)
	// if err != nil {
	// 	Log.Debug("ping connect db fail {}", err)
	// 	return nil, err
	// }

	url := dbUser + ":" + dbPass + "@(" + ip + ":" + port + ")/" + dbName + "?charset=utf8&loc=Local"

	db, err := sql.Open("mysql", url)
	if err != nil {
		Log.Debug("connect db faild, url is [{}], err : {}", url, err)
		return nil, err
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	err = db.Ping()
	if err != nil {
		Log.Debug("ping connect db fail,err : {}", err)
		return nil, err
	}
	var dbStu Dao
	dbStu.db = db
	return &dbStu, nil
}

func Ping(ip string, port string, dbUser string, dbPass string, dbName string) error {
	testUrl := dbUser + ":" + dbPass + "@tcp(" + ip + ":" + port + ")/" + dbName + "?charset=utf8&timeout=3s"
	db, err := sql.Open("mysql", testUrl)
	if err != nil {
		Log.Debug("connect db faild, url is [{}], err : {}", testUrl, err)
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (dao Dao) QueryMap(sql string, args ...interface{}) (map[string]string, error) {
	rows, err := dao.db.Query(sql, args...)
	if err != nil {
		Log.Error("queryMap error, sql is {}, err : {}", sql, err)
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	for i := 0; i < len(columns); i++ {
		columns[i] = strings.ToLower(columns[i])
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]string)
	for rows.Next() {
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = toStr(col)
			}
		}
	}
	return record, nil
}

func (dao Dao) QueryMapInterface(sql string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := dao.db.Query(sql, args...)
	if err != nil {
		Log.Error("queryMap error, sql is {}, err : {}", sql, err)
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	for i := 0; i < len(columns); i++ {
		columns[i] = strings.ToLower(columns[i])
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	for rows.Next() {
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
	}
	return record, nil
}

func (dao Dao) GetOriDb() *sql.DB {
	return dao.db
}

func toStr(inter interface{}) string {
	if inter == nil {
		return ""
	}
	switch value := inter.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case int64:
		return fmt.Sprintf("%v", value)
	case float64:
		return decimal.NewFromFloat(value).String()
	default:
		return string(inter.([]byte))
	}
}

func (dao Dao) QueryList(sql string, args ...interface{}) (*ListResult, error) {
	t1 := time.Now()
	rows, err := dao.db.Query(sql, args...)
	Log.Trace("QueryList cost %.2f seconds, sql is {}, params is {}", time.Since(t1).Seconds(), sql, zjson.ToJSONString(args))
	if err != nil {
		Log.Error("QueryList error, sql is {}, err : {}", sql, err)
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	for i := 0; i < len(columns); i++ {
		columns[i] = strings.ToLower(columns[i])
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	list := make([]map[string]string, 0)

	for rows.Next() {
		record := make(map[string]string)
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = toStr(col)
			}
		}
		list = append(list, record)
	}
	return &ListResult{columns, list}, nil
}

func (dao Dao) Update(sql string, args ...interface{}) (int64, error) {

	result, err := dao.db.Exec(sql, args...)
	if err != nil {
		Log.Error("exec failed err is {}, sql is {}", err, sql)
		return 0, err
	}

	idAff, err := result.RowsAffected()
	if err != nil {
		Log.Error("RowsAffected failed {}", err)
		return 0, err
	}
	// Log.Debug("Update  sql finish, sql is {}, args is {}", sql, args)
	return idAff, nil
}

// 返回插入后的自增id
func (dao Dao) Insert(sql string, args ...interface{}) (int64, error) {
	tx, err := dao.db.Begin()
	if err != nil {
		logger.Error("begin tx err : {}", err)
		return 0, err
	}
	result, err := tx.Exec(sql, args...)
	if err != nil {
		logger.Error("tx exec sql err : {}", err)
		tx.Rollback()
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("get LastInsertId err : {}", err)
		tx.Rollback()
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		logger.Error("tx commit err : {}", err)
		return 0, err
	}
	return id, nil
}

func (dao Dao) SelectCount(sql string, args ...interface{}) (int, error) {
	var count int
	err := dao.db.QueryRow(sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
