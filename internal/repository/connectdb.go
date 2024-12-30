package repository

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (*sql.DB, error) {
	dsn := "root:158334@tcp(127.0.0.1:3306)/blog_db"
	//dsn := "root:zjh20031229.@tcp(127.0.0.1:3306)/blog_db"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
func VerifyCode(mail string, code string) bool {
	db, err := Connect()
	if err != nil {
		return false
	}
	defer db.Close()
	// 查询验证码
	row := db.QueryRow("SELECT code FROM verificationcodes WHERE expiration > NOW() AND email = ? AND code = ?", mail, code)
	var dbCode string
	if err := row.Scan(&dbCode); err != nil {
		return false
	}
	return true
}
