package clauses

import (
	"sync"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type user struct {
	ID        uint
	OrgID     string `gorm:"uniqueIndex:idx_users_org_email,priority:1"`
	Email     string `gorm:"uniqueIndex:idx_users_org_email,priority:2"`
	Reference string `gorm:"uniqueIndex:idx_users_reference,where:deleted_at IS NULL"`
	Name      string `gorm:"index:idx_users_name"`
}

// newTx builds a statement carrying a parsed schema, no database driver needed.
func newTx(t *testing.T) *gorm.DB {
	t.Helper()
	s, err := schema.Parse(&user{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	return &gorm.DB{Statement: &gorm.Statement{Schema: s, Clauses: map[string]clause.Clause{}}}
}

func onConflictOf(t *testing.T, tx *gorm.DB) (clause.OnConflict, bool) {
	t.Helper()
	c, ok := tx.Statement.Clauses[clause.OnConflict{}.Name()]
	if !ok || c.Expression == nil {
		return clause.OnConflict{}, false
	}
	onConflict, ok := c.Expression.(clause.OnConflict)
	if !ok {
		t.Fatalf("clause expression is %T, want clause.OnConflict", c.Expression)
	}
	return onConflict, true
}

func columnNames(cols []clause.Column) []string {
	names := make([]string, len(cols))
	for i, col := range cols {
		names[i] = col.Name
	}
	return names
}

func TestOnConflictUniqueIndex(t *testing.T) {
	tx := newTx(t)
	OnConflict(tx, "idx_users_org_email")

	onConflict, ok := onConflictOf(t, tx)
	if !ok {
		t.Fatal("no ON CONFLICT clause added, want one")
	}
	if !onConflict.UpdateAll {
		t.Error("UpdateAll = false, want true")
	}

	got, want := columnNames(onConflict.Columns), []string{"org_id", "email"}
	if len(got) != len(want) {
		t.Fatalf("columns = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("columns = %v, want %v", got, want)
		}
	}
	if len(onConflict.TargetWhere.Exprs) != 0 {
		t.Errorf("TargetWhere = %v, want empty", onConflict.TargetWhere.Exprs)
	}
}

func TestOnConflictPartialUniqueIndexSetsTargetWhere(t *testing.T) {
	tx := newTx(t)
	OnConflict(tx, "idx_users_reference")

	onConflict, ok := onConflictOf(t, tx)
	if !ok {
		t.Fatal("no ON CONFLICT clause added, want one")
	}
	if len(onConflict.TargetWhere.Exprs) != 1 {
		t.Fatalf("TargetWhere has %d exprs, want 1", len(onConflict.TargetWhere.Exprs))
	}
	expr, ok := onConflict.TargetWhere.Exprs[0].(clause.Expr)
	if !ok {
		t.Fatalf("TargetWhere expr is %T, want clause.Expr", onConflict.TargetWhere.Exprs[0])
	}
	if expr.SQL != "deleted_at IS NULL" {
		t.Errorf("TargetWhere SQL = %q, want %q", expr.SQL, "deleted_at IS NULL")
	}
}

func TestOnConflictIgnoresNonMatchingIndex(t *testing.T) {
	tests := []struct {
		name string
		idx  string
	}{
		{"non unique index", "idx_users_name"},
		{"unknown index", "idx_users_missing"},
		// idx names an index, never a field: a field name must not select its index.
		{"field name", "Email"},
		{"column name", "email"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newTx(t)
			OnConflict(tx, tt.idx)

			if onConflict, ok := onConflictOf(t, tx); ok {
				t.Errorf("ON CONFLICT clause added for %q: %+v, want none", tt.idx, onConflict)
			}
		})
	}
}

func TestOnConflictWithoutSchema(t *testing.T) {
	tx := &gorm.DB{Statement: &gorm.Statement{Clauses: map[string]clause.Clause{}}}

	OnConflict(tx, "idx_users_org_email")

	if onConflict, ok := onConflictOf(t, tx); ok {
		t.Errorf("ON CONFLICT clause added without a schema: %+v, want none", onConflict)
	}
}
