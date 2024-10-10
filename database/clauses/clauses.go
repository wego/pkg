package clauses

import (
	"github.com/wego/pkg/strings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OnConflict adds an ON CONFLICT clause to the query
func OnConflict(tx *gorm.DB, idx string) {
	index, ok := tx.Statement.Schema.ParseIndexes()[idx]
	if !ok || index.Class != "UNIQUE" {
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
