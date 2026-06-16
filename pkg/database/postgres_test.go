package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type failingDialector struct{}

func (d failingDialector) Name() string                                     { return "failing" }
func (d failingDialector) Initialize(db *gorm.DB) error                     { return errors.New("init failed") }
func (d failingDialector) Migrator(db *gorm.DB) gorm.Migrator               { return nil }
func (d failingDialector) DataTypeOf(field *schema.Field) string            { return "" }
func (d failingDialector) DefaultValueOf(field *schema.Field) clause.Expression { return nil }
func (d failingDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}
func (d failingDialector) QuoteTo(writer clause.Writer, str string)           {}
func (d failingDialector) Explain(sql string, vars ...interface{}) string     { return sql }

func TestNewGorm(t *testing.T) {
	db, err := NewGorm(sqlite.Open(":memory:"))
	assert.NoError(t, err)
	assert.NotNil(t, db)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)
	assert.NoError(t, sqlDB.Close())
}

func TestNewGorm_DialectorError(t *testing.T) {
	_, err := NewGorm(&failingDialector{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "init failed")
}

func TestNewPostgres(t *testing.T) {
	db, err := NewPostgres("")
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, db)
	}
}
