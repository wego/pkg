package example

// String literals passed as arguments to ORM, HTTP, and filter methods are
// column or parameter names, not ISO code values. They must not be flagged
// even when the string happens to match a valid ISO code.

// Mock types to simulate ORM and HTTP framework method signatures.

type mockDB struct{}

func (mockDB) Select(cols ...string) mockDB { return mockDB{} }
func (mockDB) Pluck(col string, dest any) mockDB { return mockDB{} }
func (mockDB) Omit(cols ...string) mockDB { return mockDB{} }

type mockCtx struct{}

func (mockCtx) Query(key string) string         { return "" }
func (mockCtx) QueryParam(key string) string     { return "" }
func (mockCtx) Param(key string) string          { return "" }
func (mockCtx) FormValue(key string) string      { return "" }
func (mockCtx) GetQuery(key string) string       { return "" }
func (mockCtx) DefaultQuery(key, def string) string { return "" }
func (mockCtx) PostForm(key string) string       { return "" }

type mockFilter struct{}

func (mockFilter) Equals(col string, val any) mockFilter    { return mockFilter{} }
func (mockFilter) NotEquals(col string, val any) mockFilter { return mockFilter{} }

func contextSkips() {
	// ORM column-name arguments — not site/currency values
	db := mockDB{}
	db.Select("SG")      // column name, not Singapore
	db.Select("FR", "DE") // column names, not France/Germany
	db.Pluck("TH", nil)  // column name, not Thailand
	db.Omit("MY", "PH")  // column names, not Malaysia/Philippines

	// HTTP parameter-name arguments — not site/currency values
	c := mockCtx{}
	c.Query("AE")          // param name, not UAE
	c.QueryParam("AU")     // param name, not Australia
	c.Param("GB")          // param name, not United Kingdom
	c.FormValue("IN")      // param name, not India
	c.GetQuery("HK")       // param name, not Hong Kong
	c.DefaultQuery("JP", "") // param name, not Japan
	c.PostForm("NZ")       // param name, not New Zealand

	// Filter column-name arguments — not site/currency values
	f := mockFilter{}
	f.Equals("CA", "val")    // column name, not Canada
	f.NotEquals("VN", "val") // column name, not Vietnam

	// Currency codes in call context
	db.Select("USD") // column name, not US Dollar
	c.Query("SGD")   // param name, not Singapore Dollar
}
