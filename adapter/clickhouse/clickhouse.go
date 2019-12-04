package clickhouse

import (
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/k0kubun/sqldef/adapter"
)

type ClickhouseDatabase struct {
	config adapter.Config
	db     *sql.DB
}

func CheckError(err error) {
	if err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
	}
}

func NewDatabase(config adapter.Config) (adapter.Database, error) {
	db, err := sql.Open("clickhouse", clickhouseBuildDSN(config))
	if err != nil {
		return nil, err
	}

	return &ClickhouseDatabase{
		db:     db,
		config: config,
	}, nil
}

func (d *ClickhouseDatabase) TableNames() ([]string, error) {
	rows, err := d.db.Query("select name from system.tables where engine != 'MaterializedView' and database != 'system'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (d *ClickhouseDatabase) DumpTableDDL(table string) (string, error) {
	var ddl string
	sql := fmt.Sprintf("show create table `%s`;", table) // TODO: escape table name

	err := d.db.QueryRow(sql).Scan(&ddl)
	if err != nil {
		return "", err
	}

	return ddl, nil
}

func (d *ClickhouseDatabase) DB() *sql.DB {
	return d.db
}

func (d *ClickhouseDatabase) Close() error {
	return d.db.Close()
}

func clickhouseBuildDSN(config adapter.Config) string {
	username := config.User
	password := config.Password
	database := config.DbName
	hostname := ""
	if config.Socket == "" {
		hostname = fmt.Sprintf("%s:%d", config.Host, config.Port)
	} else {
		hostname = config.Socket
	}

	return fmt.Sprintf("tcp://%s?username=%s&password=%s&database=%s", hostname, username, password, database)
}
