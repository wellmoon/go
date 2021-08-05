package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"

	Log "github.com/wellmoon/go/logger"
)

type ListResult struct {
	Columns []string             `json:"columns"`
	List    *[]map[string]string `json:"list"`
}

type Dao struct {
	db *sql.DB
}

func NewDao(ip string, port string, dbUser string, dbPass string, dbName string) *Dao {
	url := dbUser + ":" + dbPass + "@tcp(" + ip + ":" + port + ")/" + dbName + "?charset=utf8"

	db, err := sql.Open("mysql", url)
	if err != nil {
		Log.Debug("connect db faild, url is [{}], err : {}", url, err)
		return nil
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()

	var dbStu Dao
	dbStu.db = db
	return &dbStu
}

func (dao Dao) QueryMap(sql string, args ...interface{}) map[string]string {
	rows, err := dao.db.Query(sql, args...)
	if err != nil {
		Log.Error("查询失败，sql is {}, err : {}", sql, err)
		return nil
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]string)
	for rows.Next() {
		//将行数据保存到record字典
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}
	}
	return record
}

func (dao Dao) QueryList(sql string, args ...interface{}) *ListResult {
	t1 := time.Now()
	rows, err := dao.db.Query(sql, args...)
	Log.Debug("查询list耗时 %.2f 秒，sql is {}", time.Since(t1).Seconds(), sql)
	if err != nil {
		Log.Error("查询失败，sql is {}, err : {}", sql, err)
		return nil
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	list := make([]map[string]string, 0)

	for rows.Next() {
		record := make(map[string]string)
		//将行数据保存到record字典
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}
		list = append(list, record)
	}
	return &ListResult{columns, &list}
}

func (dao Dao) Update(sql string, args ...interface{}) (int64, string) {

	result, err := dao.db.Exec(sql)
	if err != nil {
		Log.Error("exec failed err is {}, sql is {}", err, sql)
		return 0, err.Error()
	}

	idAff, err := result.RowsAffected()
	if err != nil {
		Log.Error("RowsAffected failed {}", err)
		return 0, err.Error()
	}
	return idAff, ""
}
