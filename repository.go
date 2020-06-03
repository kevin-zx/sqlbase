package sql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Storage struct {
	db *gorm.DB
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
	s.db = db
	return s, err
}

func (p *Storage) Close() {
	_ = p.db.Close()
}

func (p *Storage) Search(dbType interface{}, params map[string]string, needCount bool, preLoads []string, data interface{}) (int, error) {
	queryDb := p.db.Model(dbType)
	//if preLoads != "" {
	//	queryDb = queryDb.Preload(preLoads)
	//}
	for _, load := range preLoads {
		queryDb = queryDb.Preload(load)
	}
	queryDb, q := convertParams2DbQuery(queryDb, params)
	c := 0
	if needCount {
		queryDb = queryDb.Count(&c)
	}
	err := addAssistQuery(queryDb, q).Find(data).Error
	return c, err
}

func (p *Storage) SaveOrCreate(params map[string]string, data interface{}) error {
	queryDb, _ := convertParams2DbQuery(p.db, params)
	return queryDb.FirstOrCreate(data).Error
}

func (p *Storage) Delete(dbType interface{}, params map[string]string) (rowsAffected int64, err error) {
	queryDb := p.db.Model(dbType)
	queryDb, q := convertParams2DbQuery(queryDb, params)
	finalDb := addAssistQuery(queryDb, q).Delete(dbType)
	err = finalDb.Error
	rowsAffected = finalDb.RowsAffected
	return
}

func (p *Storage) Update(data interface{}) error {
	return p.db.Update(data).Error
}
