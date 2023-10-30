package daos

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/daos/clients/sqls"
	"github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/models"
	log "github.com/sirupsen/logrus"
)

type UserDao struct {
	sqlClient *sqls.MySQLClient
}

func migrateUsers(r *sqls.MySQLClient) error {
	query := `
	CREATE TABLE IF NOT EXISTS users(
		ID int NOT NULL AUTO_INCREMENT,
        
		Name VARCHAR(100) NOT NULL,
	    PRIMARY KEY (ID)
	);
	`
	_, err := r.DB.Exec(query)
	return err
}

func NewUserDao() (*UserDao, error) {
	sqlClient, err := sqls.InitMySQLDB()
	if err != nil {
		return nil, err
	}
	err = migrateUsers(sqlClient)
	if err != nil {
		return nil, err
	}
	return &UserDao{
		sqlClient,
	}, nil
}

func (userDao *UserDao) CreateUser(m *models.User) (*models.User, error) {
	insertQuery := "INSERT INTO users(Name) values(?)"
	res, err := userDao.sqlClient.DB.Exec(insertQuery, m.Name)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return nil, sqls.ErrDuplicate
			}
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	m.Id = id
	log.Debugf("user created")
	return m, nil
}

func (userDao *UserDao) ListUsers() ([]*models.User, error) {
	selectQuery := "SELECT * FROM users"
	rows, err := userDao.sqlClient.DB.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var users []*models.User
	for rows.Next() {
		m := models.User{}
		if err = rows.Scan(&m.Id, &m.Name); err != nil {
			return nil, err
		}
		users = append(users, &m)
	}
	if users == nil {
		users = []*models.User{}
	}
	log.Debugf("user listed")
	return users, nil
}

func (userDao *UserDao) GetUser(id int64) (*models.User, error) {
	selectQuery := "SELECT * FROM users WHERE Id = ?"
	row := userDao.sqlClient.DB.QueryRow(selectQuery, id)

	m := models.User{}
	if err := row.Scan(&m.Id, &m.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sqls.ErrNotExists
		}
		return nil, err
	}
	log.Debugf("user retrieved")
	return &m, nil
}
