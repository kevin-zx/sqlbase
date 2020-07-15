package sqlbase

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
)

type Storage struct {
	DB *gorm.DB
}

func NewStorage(engine string, port int, dbName string,
	user string, host string, passWd string) (*Storage, error) {
	var err error
	db, err := gorm.Open(engine, fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		user, passWd, host, port, dbName))
	if err != nil {
		return nil, err
	}
	s := new(Storage)
	s.DB = db
	return s, err
}

func (p *Storage) Close() {
	_ = p.DB.Close()
}

func (p *Storage) Search(dbType interface{}, params map[string]string, needCount bool, preLoads []string, data interface{}) (int, error) {
	queryDb := p.DB.Model(dbType)
	//if preLoads != "" {
	//	queryDb = queryDb.Preload(preLoads)
	//}
	for _, load := range preLoads {
		queryDb = queryDb.Preload(load)
	}
	queryDb, q := ConvertParams2DbQuery(queryDb, params)
	c := 0
	if needCount {
		queryDb = queryDb.Count(&c)
	}
	err := addAssistQuery(queryDb, q).Find(data).Error
	return c, err
}

// RawScan this method can execute raw sql and use the scan function to scan rows
// this can package many are easy to overlook operations, like close.
// baseSql: the main sql body, provide query logic.
// conditionAndLimitPart: the sql condition body and limit part.
// scan: this revoke function to scan result rows.
// values: the values to replace sql placeholders
func (p *Storage) RawScan(baseSql string, conditionAndLimitPart string, scan func(rows *sql.Rows) error, values ...interface{}) error {
	rows, err := p.DB.Raw(baseSql+conditionAndLimitPart, values...).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	exist := false
	for rows.Next() {
		exist = true
		err = scan(rows)
		if err != nil {
			return err
		}
	}
	if !exist {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetLastID: get table last primary id by table and primaryName
func (p *Storage) GetLastID(table string, primaryName string) (uint, error) {
	if primaryName == "" {
		primaryName = "id"
	}
	row := p.DB.Raw("SELECT `" + primaryName + "` FROM `" + table + "`   ORDER BY `" + table + "`.`" + primaryName + "` DESC LIMIT 1").Row()
	lastId := uint(0)

	err := row.Scan(&lastId)
	// if table is empty return 0
	if err != nil && err.Error() == "sql: no rows in result set" {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return lastId, nil

}

func (p *Storage) BatchInsert(baseSql string, valueFmt string, batchSize int, values [][]interface{}) error {
	var tValueFmts []string
	var tValues []interface{}
	var err error
	for i, vls := range values {
		tValues = append(tValues, vls...)
		tValueFmts = append(tValueFmts, valueFmt)
		if len(tValueFmts) > 0 && (i == len(values)-1 || (i%batchSize == 0 && i != 0)) {
			err = p.DB.Raw(baseSql+strings.Join(tValueFmts, ","), tValues...).Error
			if err != nil {
				return err
			}
			tValues = []interface{}{}
			tValueFmts = []string{}
		}
	}
	return nil
}

func (p *Storage) SaveOrCreate(params map[string]string, data interface{}) error {
	queryDb, _ := ConvertParams2DbQuery(p.DB, params)
	return queryDb.FirstOrCreate(data).Error
}

func (p *Storage) Delete(dbType interface{}, params map[string]string) (rowsAffected int64, err error) {
	queryDb := p.DB.Model(dbType)
	queryDb, q := ConvertParams2DbQuery(queryDb, params)
	finalDb := addAssistQuery(queryDb, q).Delete(dbType)
	err = finalDb.Error
	rowsAffected = finalDb.RowsAffected
	return
}

func (p *Storage) Update(data interface{}) error {
	return p.DB.Update(data).Error
}
