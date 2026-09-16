package clauses

import (
	"github.com/wego/pkg/strings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// OnConflict adds an ON CONFLICT clause to the query
func OnConflict(tx *gorm.DB, idx string) {
	index := uniqueIndex(tx.Statement.Schema, idx)
	if index == nil {
		return
	}

	cols := make([]clause.Column, len(index.Fields))
	for i, col := range index.Fields {
		cols[i] = clause.Column{Name: col.DBName}
	}

	onConflictClause := clause.OnConflict{
		Columns:   cols,
		UpdateAll: true,
	}

	if strings.IsNotBlank(index.Where) {
		onConflictClause.TargetWhere = clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: index.Where},
		}}
	}

	tx.Statement.AddClause(onConflictClause)
}

// uniqueIndex returns the unique index named idx, or nil when the schema declares no such index.
// idx matches an index name only, never a field name.
func uniqueIndex(s *schema.Schema, idx string) *schema.Index {
	if s == nil {
		return nil
	}

	for _, index := range s.ParseIndexes() {
		if index.Name == idx && index.Class == "UNIQUE" {
			return index
		}
	}

	return nil
}
