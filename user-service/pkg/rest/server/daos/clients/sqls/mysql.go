package sqls

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"os"
	"sync"
	"time"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

var (
	user     = os.Getenv("MYSQL_DB_USER")
	password = os.Getenv("MYSQL_DB_PASSWORD")
	host     = os.Getenv("MYSQL_DB_HOST")
	port     = os.Getenv("MYSQL_DB_PORT")
	database = os.Getenv("MYSQL_DB_DATABASE")
)

type MySQLClient struct {
	DB *sql.DB
}

func NewMySQLClient(db *sql.DB) *MySQLClient {
	return &MySQLClient{
		DB: db,
	}
}

var o sync.Once
var err error
var mySQLClient *MySQLClient

func InitMySQLDB() (*MySQLClient, error) {
	o.Do(func() {
		dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
		var db *sql.DB
		serviceName := os.Getenv("SERVICE_NAME")
		collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		if len(serviceName) > 0 && len(collectorURL) > 0 {
			// add opentel
			db, err = otelsql.Open("mysql", dataSource, otelsql.WithAttributes(semconv.DBSystemMySQL))
		} else {
			db, err = sql.Open("mysql", dataSource)
		}
		if err != nil {
			log.Debugf("unable to connect to database, %v", err)
			os.Exit(1)
		}
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		mySQLClient = NewMySQLClient(db)
	})

	return mySQLClient, nil
}
