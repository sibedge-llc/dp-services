package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/event"
)

var (
	//go:embed queries/select_table_columns.sql
	selectTableColumnsSql string
)

type Converter func(in interface{}) (string, error)

type tableSchema struct {
	Sql            string
	Converters     map[string]Converter
	Columns        []string
	ColumnsIndices map[string]int
}

type Db struct {
	ctx     context.Context
	db      *sqlx.DB
	timeout time.Duration
	cfg     *config.PostgresConfig
	id      uint64
	lock    sync.Mutex
	schema  *tableSchema
}

type sqlColumnType struct {
	DataType   string
	UdtName    string
	IsNullable bool
}

func NewDb(ctx context.Context, id uint64, cfg *config.PostgresConfig, timeout time.Duration) (*Db, error) {

	if cfg.Table == "" {
		return nil, errors.New("table name is empty or not provided")
	}

	db, err := sqlx.ConnectContext(ctx, "postgres", getDataSource(cfg))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	err = db.PingContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return &Db{
		ctx:     ctx,
		db:      db,
		timeout: timeout,
		cfg:     cfg,
		id:      id,
	}, nil
}

func (db *Db) GetConfig() *config.PostgresConfig {
	return db.cfg
}

func (db *Db) Close() {
	err := db.db.Close()
	if err != nil {
		zap.L().Error("failed to close postgres", zap.Error(err))
	}
}

func (db *Db) Flush() {
}

func (db *Db) GetId() uint64 {
	return db.id
}

func (db *Db) Init(evt *event.Event) error {
	return db.updateOrCreateTableSchema(evt.Object)
}

func (db *Db) Send(evt *event.Event) error {
	query, err := db.composeQuery(evt.Object)
	if err != nil {
		return fmt.Errorf("failed to compose query along event: %w", err)
	}

	_, err = db.db.ExecContext(db.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to perform upsert the event: %w", err)
	}
	return nil
}

func getDataSource(cfg *config.PostgresConfig) string {
	parts := make([]string, 0, 6)
	if cfg.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", cfg.Host))
	}
	if cfg.Db != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", cfg.Db))
	}
	if cfg.Port != 0 {
		parts = append(parts, fmt.Sprintf("port=%d", cfg.Port))
	}
	if cfg.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", cfg.User))
	}
	if cfg.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", cfg.Password))
	}

	if !cfg.Ssl {
		parts = append(parts, "sslmode=disable")
	}
	return strings.Join(parts, " ")
}

func (db *Db) updateOrCreateTableSchema(obj event.EventObject) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.schema != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(db.ctx, db.timeout)
	defer cancel()

	params := struct {
		TableName string `db:"table_name"`
	}{
		TableName: db.cfg.Table,
	}

	rows, err := db.db.NamedQueryContext(ctx, selectTableColumnsSql, params)
	if err != nil {
		return err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			zap.L().Error("close rows error", zap.Error(err))
		}
	}()

	sqlColumns := make([]string, 0, 10)
	converters := make(map[string]Converter, 10)

	for rows.Next() {
		row := struct {
			ColumnName string `db:"column_name"`
			DataType   string `db:"data_type"`
			UdtName    string `db:"udt_name"`
			IsNullable bool   `db:"is_nullable"`
		}{}
		err := rows.StructScan(&row)
		if err != nil {
			return err
		}

		v, ok := obj[row.ColumnName]
		if ok {
			sqlColumnType := sqlColumnType{
				row.DataType,
				row.UdtName,
				row.IsNullable,
			}
			converter, err := toConverterByColumnDef(sqlColumnType, v)
			if err != nil {
				return err
			}
			converters[row.ColumnName] = converter
		}
	}

	noTable := len(converters) == 0

	keyColumnNames := make([]string, 0, len(obj))
	updatedColumnNames := make([]string, len(sqlColumns))

	for k, v := range obj {
		if _, ok := converters[k]; ok {
			continue
		}
		sqlColumnDef, err := toSqlColumnDefinition(k, v)
		if err != nil {
			return err
		}
		sqlColumns = append(sqlColumns, sqlColumnDef.ColumnDefinition)
		if sqlColumnDef.IsKey {
			keyColumnNames = append(keyColumnNames, k)
		} else {
			updatedColumnNames = append(updatedColumnNames, fmt.Sprintf("%s=EXCLUDED.%s", k, k))
		}
		converters[k] = sqlColumnDef.Converter
	}

	if len(sqlColumns) == 0 {
		return fmt.Errorf("No columns can be add or any table created")
	}

	sort.Strings(keyColumnNames)

	createOrUpdateTableSql := ""
	if noTable {
		// Create table
		keySql := ""
		if len(keyColumnNames) > 0 {
			keySql = fmt.Sprintf(",\nCONSTRAINT pk_%s PRIMARY KEY (%s)", db.cfg.Table, strings.Join(keyColumnNames, ","))
		}
		createOrUpdateTableSql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s%s)", db.cfg.Table, strings.Join(sqlColumns, ",\n"), keySql)
	} else {
		// Update table
		addSqlColumns := make([]string, len(sqlColumns))
		for i, sqlCol := range sqlColumns {
			addSqlColumns[i] = fmt.Sprintf("ADD COLUMN %s", sqlCol)
		}
		createOrUpdateTableSql = fmt.Sprintf("ALTER TABLE IF EXISTS %s (%s)", db.cfg.Table, strings.Join(addSqlColumns, ",\n"))
	}

	_, err = db.db.ExecContext(ctx, createOrUpdateTableSql)
	if err != nil {
		return err
	}

	columnsIndices := make(map[string]int, len(sqlColumns))
	columns := make([]string, len(sqlColumns))
	for name := range converters {
		columnsIndices[name] = len(columnsIndices)
		columns[columnsIndices[name]] = name
	}

	db.schema = &tableSchema{
		Sql: fmt.Sprintf(
			"INSERT INTO %s(%s) VALUES(%%s) ON CONFLICT(%s) DO UPDATE SET %s",
			db.cfg.Table,
			strings.Join(columns, ","),
			strings.Join(keyColumnNames, ","),
			strings.Join(updatedColumnNames, ","),
		),
		Converters:     converters,
		Columns:        columns,
		ColumnsIndices: columnsIndices,
	}
	return nil
}

func (db *Db) composeQuery(o event.EventObject) (string, error) {
	insertValues := make([]string, len(db.schema.ColumnsIndices))
	for name, index := range db.schema.ColumnsIndices {
		converter := db.schema.Converters[name]
		val, ok := o[name]
		var v string
		if !ok {
			zap.L().Debug("missed field", zap.String("name", name))
		}
		var err error
		v, err = converter(val)
		if err != nil {
			return "", fmt.Errorf("failed to convert field: %s value: %v failed: %w", name, val, err)
		}
		insertValues[index] = v
	}

	return fmt.Sprintf(db.schema.Sql, strings.Join(insertValues, ",")), nil
}
