package sqlbase

import (
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type Query struct {
	SortBy string
	Order  string
	Offset int
	Limit  int
}

// 转化请求的参数变成查询
// 把判断操作都直接放入到查询中
// 把 limit offset order 这些操作提取出来
func ConvertParams2DbQuery(initQuery *gorm.DB, params map[string]string) (*gorm.DB, Query) {
	q := Query{}
	if sortBy, ok := params["sort_by"]; ok {
		q.SortBy = sortBy
		delete(params, "sort_by")
	}
	if order, ok := params["order"]; ok {
		q.Order = order
		delete(params, "order")
	}
	if index, ok := params["index"]; ok {
		q.Offset, _ = strconv.Atoi(index)
		delete(params, "index")
	}
	if offset, ok := params["offset"]; ok {
		q.Offset, _ = strconv.Atoi(offset)
		delete(params, "offset")
	}
	if limit, ok := params["limit"]; ok {
		q.Limit, _ = strconv.Atoi(limit)
		delete(params, "limit")
	}

	for k, v := range params {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if strings.HasSuffix(k, "_like") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-5]
			initQuery = initQuery.Where(k+" like ?", v)
		} else if strings.HasSuffix(k, "_in") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]
			initQuery = initQuery.Where(k+" in (?)", strings.Split(v, ","))
		} else if strings.HasSuffix(k, "_gt") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]
			initQuery = initQuery.Where(k+" > ?", v)
		} else if strings.HasSuffix(k, "_ge") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]
			initQuery = initQuery.Where(k+" >= ?", v)
		} else if strings.HasSuffix(k, "_lt") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]
			initQuery = initQuery.Where(k+" < ?", v)
		} else if strings.HasSuffix(k, "_le") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]
			initQuery = initQuery.Where(k+" <= ?", v)
		} else if strings.HasSuffix(k, "_is") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-3]

			initQuery = initQuery.Where(k + " is " + v + " ")
		} else if strings.HasSuffix(k, "_is_not") {
			if v == "" {
				continue
			}
			k = k[0 : len(k)-7]

			initQuery = initQuery.Where(k + " is not " + v + " ")
		} else {
			initQuery = initQuery.Where(k+" = ?", v)
		}
	}

	return initQuery, q

}

// 执行order limit offset 这些辅助操作
func addAssistQuery(initQuery *gorm.DB, q Query) *gorm.DB {
	if q.SortBy != "" {
		initQuery = initQuery.Order(q.SortBy + " " + q.Order)
	}
	if q.Offset != 0 {
		initQuery = initQuery.Offset(q.Offset)
	}
	if q.Limit != 0 {
		initQuery = initQuery.Limit(q.Limit)
	}
	return initQuery
}
