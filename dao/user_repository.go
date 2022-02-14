package dao

import (
	"database/sql"
	"github.com/pkg/errors"
	"seckill/common"
	"seckill/dao/db"
	"seckill/models"
	"strconv"
)

type IUserRepository interface {
	Conn()
	Select(string) (*models.User, error)
	Insert(*models.User) (int64, error)
}

type UserManagerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func NewUserRepository(table string, db *sql.DB) IUserRepository {
	return &UserManagerRepository{
		table:     table,
		mysqlConn: db,
	}
}

func (u *UserManagerRepository) Conn() {
	if u.mysqlConn == nil {
		u.mysqlConn = db.DBConn()
	}
	if u.table == "" {
		u.table = "user"
	}
}

func (u *UserManagerRepository) Select(userName string) (*models.User, error) {
	if userName == "" {
		return &models.User{}, errors.New("user_repository#Select: userName is empty")
	}
	u.Conn()

	sql := "SELECT * FROM `" + u.table + "` WHERE userName=?"
	rows, err := u.mysqlConn.Query(sql, userName)
	if err != nil {
		return &models.User{}, errors.Wrap(err, "user_repository#Select: query failed")
	}
	defer rows.Close()

	result := db.GetResultRow(rows)
	if len(result) == 0 {
		return &models.User{}, errors.New("user_repository#Select: userName is not found")
	}

	user := &models.User{}
	common.DataToStructByTagSql(result, user)
	return user, nil
}

func (u *UserManagerRepository) Insert(user *models.User) (int64, error) {
	u.Conn()

	sql := "INSERT `" + u.table + "` SET nickName=?,userName=?,passWord=?,userIp=?"
	stmt, err := u.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, errors.Wrap(err, "user_repository#Insert: sql prepare failed")
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.NickName, user.UserName, user.HashPassword, user.UserIp)
	if err != nil {
		return 0, errors.Wrap(err, "user_repository#Insert: insert failed")
	}

	return result.LastInsertId()
}

func (u *UserManagerRepository) user_repository(userId int64) (*models.User, error) {
	u.Conn()

	sql := "SELECT * FROM `" + u.table + "` WHERE ID=" + strconv.FormatInt(userId, 10)
	row, err := u.mysqlConn.Query(sql)
	if err != nil {
		return &models.User{}, errors.Wrap(err, "user_repository#user_repository: query failed")
	}
	defer row.Close()

	result := db.GetResultRow(row)
	if len(result) == 0 {
		return &models.User{}, errors.New("user_repository#user_repository: user not found")
	}

	user := &models.User{}
	common.DataToStructByTagSql(result, user)
	return user, nil
}
