package db

import (
	"BackupHunter/types"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql驱动
)

func Init() {
	dbUserName := types.GlobalConfig.Server.DbUserName
	dbPassWord := types.GlobalConfig.Server.DbPassword
	dbName := types.GlobalConfig.Server.DbName
	dbPort := types.GlobalConfig.Server.DbPort
	dbHost := types.GlobalConfig.Server.DbHost
	dsn := dbUserName + ":" + dbPassWord + "@tcp(" + dbHost + ":" + strconv.Itoa(dbPort) + ")/" + dbName + "?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("[ERROR] Database connected failed!")
		panic(err)
	}
	DB.SetMaxOpenConns(25)                 // 最大打开连接数
	DB.SetMaxIdleConns(5)                  // 最大空闲连接数
	DB.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期
	if err := DB.Ping(); err != nil {
		fmt.Println("[ERROR] Database connected failed!")
		panic(err)
	}
	types.DB = DB

	// 检查表是否存在，不存在则创建
	tableNames := GetAllTableName()
	findScheduledTasksTableName := false
	findSubDomainTableName := false
	findResults := false
	for _, tableName := range tableNames {
		if tableName == "scheduled_tasks" {
			findScheduledTasksTableName = true
		}
		if tableName == "subdomains" {
			findSubDomainTableName = true
		}
		if tableName == "results" {
			findResults = true
		}
	}
	if !(findScheduledTasksTableName && findSubDomainTableName && findResults) {
		fmt.Println("[INFO] Database Tables not find, auto create table")
		TablesInit(DB)
	}
}

// 关闭数据库
func Close() {
	if types.DB != nil {
		types.DB.Close()
	}
}
