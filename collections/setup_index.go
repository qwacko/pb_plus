package collections

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

type IndexConfig struct {
	Fields []string `mapstructure:"fields" json:"fields"`
	Unique bool     `mapstructure:"unique" json:"unique"`
	Id     string   `mapstructure:"id" json:"id"`
}

type IndexReturn struct {
	Name       string `json:"name"`
	Unique     bool   `json:"unique"`
	ColumnExpr string `json:"column_expr"`
}

func (i *IndexConfig) getIndexQuery(collection *core.Collection) IndexReturn {

	for _, field := range i.Fields {
		if collection.Fields.GetByName(field) == nil {
			log.Panicf("Unknown field %s when configuring indexes on collection %s", field, collection.Name)
		}
	}

	columns_expr := ""
	for i, field := range i.Fields {
		if i > 0 {
			columns_expr += ", "
		}
		columns_expr += "`" + field + "`"
	}

	return IndexReturn{
		Name:       i.Id,
		Unique:     i.Unique,
		ColumnExpr: columns_expr,
	}

}
