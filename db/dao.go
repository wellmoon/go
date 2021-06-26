package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	Log "github.com/wellmoon/go/logger"
)

type ListResult struct {
	Columns []string
	List    *[]map[string]string
}

type Dao struct {
	db *sql.DB
}

func NewDao(ip string, port string, dbUser string, dbPass string, dbName string) *Dao {
	url := dbUser + ":" + dbPass + "@tcp(" + ip + ":" + port + ")/" + dbName + "?charset=utf8"

	db, err := sql.Open("mysql", url)
	if err != nil {
		Log.Debug("connect db faild, url is [%v], err : %v\n", url, err)
		return nil
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()

	var dbStu Dao
	dbStu.db = db
	return &dbStu
}

func (dao Dao) QueryMap(sql string, args ...interface{}) *map[string]string {
	rows, err := dao.db.Query(sql, args...)
	if err != nil {
		Log.Error("查询失败，sql is %v, err : %v\n", sql, err)
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
	return &record
}

func (dao Dao) QueryList(sql string, args ...interface{}) *ListResult {
	rows, err := dao.db.Query(sql, args...)
	if err != nil {
		Log.Error("查询失败，sql is %v, err : %v\n", sql, err)
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
