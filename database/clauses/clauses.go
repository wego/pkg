package clauses

import (
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

	tx.Statement.AddClause(clause.OnConflict{
		Columns:   cols,
		UpdateAll: true,
	})
}
