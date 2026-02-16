package db

import (
	"BackupHunter/types"
	"database/sql"
	"fmt"
	"time"
)

// 获取所有计划任务
func GetAllTasks() ([]types.TableTasks, error) {
	query := "SELECT * FROM scheduled_tasks"
	rows, err := types.DB.Query(query)
	if err != nil {
		fmt.Println("[Error] Failed to get all tasks: " + err.Error())
		return []types.TableTasks{}, err
	}
	defer rows.Close()
	var tasks []types.TableTasks
	for rows.Next() {
		var task types.TableTasks
		if err := rows.Scan(
			&task.Id,
			&task.Name,
			&task.Cron_expr,
			&task.Domain,
			&task.Active,
			&task.Exec_count,
			&task.State,
			&task.Last_exec_time,
			&task.Created_time,
			&task.Updated_time,
		); err != nil {
			fmt.Println("[Error] Failed to scan task: " + err.Error())
			return []types.TableTasks{}, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 设置任务状态
func SetTaskState(taskId string, state int) error {
	_, err := types.DB.Exec(`
        UPDATE scheduled_tasks 
        SET state = ?
        WHERE id = ?
    `, state, taskId)
	return err
}

// 获取任务状态
func GetTaskState(taskId string) (int, error) {
	var state int
	err := types.DB.QueryRow("SELECT state FROM scheduled_tasks WHERE id = ?", taskId).Scan(&state)
	if err != nil {
		fmt.Println("[ERROR] Failed to get task state: " + err.Error())
		return 0, err
	}
	return state, nil
}

// 更新任务执行时间和次数
func UpdateTaskExecInfo(taskId string) error {
	_, err := types.DB.Exec(`
        UPDATE scheduled_tasks 
        SET exec_count = exec_count + 1,
            last_exec_time = NOW()
        WHERE id = ?
    `, taskId)
	return err
}

// GetTaskById 根据 taskId 获取单条任务记录
func GetTaskById(id string) (types.TableTasks, error) {
	query := `
		SELECT * FROM scheduled_tasks WHERE id = ?`

	var task types.TableTasks
	err := types.DB.QueryRow(query, id).Scan(
		&task.Id,
		&task.Name,
		&task.Cron_expr,
		&task.Domain,
		&task.Active,
		&task.Exec_count,
		&task.State,
		&task.Last_exec_time,
		&task.Created_time,
		&task.Updated_time,
	)
	if err != nil {
		return types.TableTasks{}, err
	}
	return task, nil
}

// 获取所有启用的任务
func GetAllActiveTasks() ([]types.TableTasks, error) {
	query := "SELECT * FROM scheduled_tasks WHERE active = TRUE"
	rows, err := types.DB.Query(query)
	if err != nil {
		fmt.Println("[Error] Failed to get active tasks: " + err.Error())
		return nil, err
	}
	defer rows.Close()

	var tasks []types.TableTasks
	for rows.Next() {
		var task types.TableTasks
		if err := rows.Scan(
			&task.Id,
			&task.Name,
			&task.Cron_expr,
			&task.Domain,
			&task.Active,
			&task.Exec_count,
			&task.State,
			&task.Last_exec_time,
			&task.Created_time,
			&task.Updated_time,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 获取所有的表
func GetAllTableName() []string {
	query := `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME
	`
	rows, err := types.DB.Query(query)
	if err != nil {
		fmt.Println("[Error] Failed to get all table names: " + err.Error())
		return nil
	}
	defer rows.Close()
	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Println("[Error] Failed to scan table name: " + err.Error())
			return nil
		}
		tableNames = append(tableNames, tableName)
	}
	return tableNames
}

// 通过主域名从数据库中查询子域名
func GetSubDomain(domain string) []string {
	rows, err := types.DB.Query("SELECT subdomain FROM subdomains WHERE domain = ?", domain)
	if err != nil {
		fmt.Println("[ERROR] scanDomainBak's database query error: ", err.Error())
		return nil
	}
	defer rows.Close()
	var subdomains []string
	for rows.Next() {
		var sub string
		err = rows.Scan(&sub)
		if err != nil {
			fmt.Println("[ERROR] scanDomainBak's database scan error: ", err.Error())
			return nil
		}
		subdomains = append(subdomains, sub)
	}
	return subdomains
}

// 获取所有子域名
func GetAllSubDomain() map[string][]string {
	rows, err := types.DB.Query("SELECT subdomain,domain FROM subdomains")
	if err != nil {
		fmt.Println("[ERROR] scanDomainBak's database query error: ", err.Error())
		return nil
	}
	defer rows.Close()
	subdomains := make(map[string][]string)
	for rows.Next() {
		var sub string
		var domain string
		err = rows.Scan(&sub, &domain)
		if err != nil {
			fmt.Println("[ERROR] scanDomainBak's database scan error: ", err.Error())
			return nil
		}
		subdomains[domain] = append(subdomains[domain], sub)
	}
	return subdomains
}

// 删除目标domain子域名
func DeleteSubDomain(domain string) error {
	_, err := types.DB.Exec("DELETE FROM subdomains WHERE domain = ?", domain)
	return err
}

// 删除单个子域名
func DeleteSingleSubDomain(subdomain, domain string) error {
	_, err := types.DB.Exec("DELETE FROM subdomains WHERE subdomain = ? AND domain = ?", subdomain, domain)
	return err
}

// 删除任务
func DeleteTask(taskId string) error {
	_, err := types.DB.Exec("DELETE FROM scheduled_tasks WHERE id = ?", taskId)
	return err
}

// 设置任务启用状态
func SetTaskActive(taskId string, active bool) error {
	_, err := types.DB.Exec("UPDATE scheduled_tasks SET active = ? WHERE id = ?", active, taskId)
	return err
}

// 批量插入子域名
func BatchInsertSubDomain(subdomains []string, domain string) error {
	tx, err := types.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO subdomains (domain, subdomain) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range subdomains {
		_, err := stmt.Exec(domain, u)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// 将结果存入数据库
func InsResult(result types.Result) error {
	_, err := types.DB.Exec("INSERT INTO results (id,urls,filename, size) VALUES (?, ?, ?, ?)", result.Id, result.URL, result.FileName, result.Size)
	if err != nil {
		return err
	}
	return nil
}

// 添加单个子域名
func AddSubDomain(subdomain string, domain string) error {
	_, err := types.DB.Exec("INSERT INTO subdomains (domain, subdomain) VALUES (?, ?)", domain, subdomain)
	return err
}

// 获取所有扫描到的URL结果
func GetResults() ([]types.Result, error) {
	rows, err := types.DB.Query("SELECT id, urls, filename, size FROM results")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	var results []types.Result
	for rows.Next() {
		var result types.Result
		if err := rows.Scan(&result.Id, &result.URL, &result.FileName, &result.Size); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// 根据id删除result
func DeleteResultById(id string) error {
	_, err := types.DB.Exec("DELETE FROM results WHERE id = ?", id)
	return err
}

// 在results表中，根据id查result
func GetUrlsById(id string) (types.Result, error) {
	rows := types.DB.QueryRow("SELECT id, urls, filename, size FROM results WHERE id = ?", id)
	var url types.Result
	if err := rows.Scan(&url.Id, &url.URL, &url.FileName, &url.Size); err != nil {
		return types.Result{}, err
	}
	return url, nil
}

// 添加计划任务到数据库
func AddCronTask(task types.TableTasks) error {
	_, err := types.DB.Exec("INSERT INTO scheduled_tasks (id, name, cron_expr,domain, active, exec_count, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
		task.Id, task.Name, task.Cron_expr, task.Domain, task.Active, 0, 0)
	return err
}

// 获取最近 N 个月的月度扫描成功数量（默认 12 个月）
func GetMonthlyScanStats(months int) []types.MonthlyScanStat {
	if months <= 0 {
		months = 12
	}
	startTime := time.Now().AddDate(0, -months, 0)

	query := `
		SELECT 
			DATE_FORMAT(create_time, '%Y-%m') AS ` + "`year_month`" + `,
			COUNT(*) AS cnt
		FROM results
		WHERE create_time >= ?
		GROUP BY DATE_FORMAT(create_time, '%Y-%m')
		ORDER BY ` + "`year_month`" + ` ASC
	`

	rows, err := types.DB.Query(query, startTime)
	if err != nil {
		fmt.Println("[ERROR] In GetMonthlyScanStats query error: " + err.Error())
		return fillMissingMonths(map[string]int{}, months)
	}
	defer rows.Close()

	statsMap := make(map[string]int)
	for rows.Next() {
		var yearMonth string
		var count int
		if err := rows.Scan(&yearMonth, &count); err != nil {
			fmt.Println("[ERROR] In GetMonthlyScanStats scan row error: " + err.Error())
			break
		}
		statsMap[yearMonth] = count
	}

	return fillMissingMonths(statsMap, months)
}

// fillMissingMonths 补全最近 N 个月的缺失月份
func fillMissingMonths(data map[string]int, lastNMonths int) []types.MonthlyScanStat {
	now := time.Now()
	start := time.Date(now.Year(), now.Month()-time.Month(lastNMonths-1), 1, 0, 0, 0, 0, now.Location())

	result := make([]types.MonthlyScanStat, 0, lastNMonths)
	for i := 0; i < lastNMonths; i++ {
		t := start.AddDate(0, i, 0)
		key := t.Format("2006-01")
		result = append(result, types.MonthlyScanStat{
			YearMonth: key,
			Count:     data[key],
		})
	}
	return result
}

// 数据库初始化，生成默认表结构
func TablesInit(DB *sql.DB) {
	const createScheduledTasksTableSQL = `
		CREATE TABLE IF NOT EXISTS scheduled_tasks (
			id VARCHAR(100) PRIMARY KEY NOT NULL,
			name VARCHAR(100) NOT NULL,
			cron_expr VARCHAR(50) NOT NULL,
			domain VARCHAR(50) NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			exec_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
			state INT UNSIGNED NOT NULL,
			last_exec_time DATETIME NULL DEFAULT NULL,
			created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
	const createSubDomainTableSQL = `
			CREATE TABLE IF NOT EXISTS subdomains (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			subdomain VARCHAR(253) NOT NULL UNIQUE,
			domain VARCHAR(253) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
	const createResultsTableSQL = `
			CREATE TABLE results (
			id VARCHAR(100) PRIMARY KEY NOT NULL,
			urls TEXT NOT NULL,
			filename VARCHAR(255) NOT NULL,
			size BIGINT UNSIGNED NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
	if _, err := DB.Exec(createScheduledTasksTableSQL); err != nil {
		fmt.Println("[Error] Failed to create scheduled_tasks table: " + err.Error())
		return
	}
	if _, err := DB.Exec(createSubDomainTableSQL); err != nil {
		fmt.Println("[Error] Failed to create subdomains table: " + err.Error())
		return
	}
	if _, err := DB.Exec(createResultsTableSQL); err != nil {
		fmt.Println("[Error] Failed to create results table: " + err.Error())
		return
	}
	fmt.Println("[Success] Tables create successfully")
}
