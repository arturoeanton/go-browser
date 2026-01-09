package webapi

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/dop251/goja"
)

// Storage implements the Web Storage API (localStorage/sessionStorage)
type Storage struct {
	db        *sql.DB
	tableName string
	mu        sync.RWMutex
	vm        *goja.Runtime
}

// NewLocalStorage creates a persistent SQLite-based localStorage
func NewLocalStorage(vm *goja.Runtime) (*Storage, error) {
	// Store in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dbPath := filepath.Join(homeDir, ".spidergopher", "storage.db")

	// Create directory if not exists
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db:        db,
		tableName: "localStorage",
		vm:        vm,
	}

	if err := storage.ensureTable(); err != nil {
		return nil, err
	}

	return storage, nil
}

// NewSessionStorage creates an in-memory storage (cleared on restart)
func NewSessionStorage(vm *goja.Runtime) (*Storage, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db:        db,
		tableName: "sessionStorage",
		vm:        vm,
	}

	if err := storage.ensureTable(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) ensureTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + s.tableName + ` (
			key TEXT PRIMARY KEY,
			value TEXT
		)
	`)
	return err
}

// GetItem retrieves a value by key
func (s *Storage) GetItem(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Null()
	}

	key := call.Argument(0).String()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var value string
	err := s.db.QueryRow("SELECT value FROM "+s.tableName+" WHERE key = ?", key).Scan(&value)
	if err != nil {
		return goja.Null()
	}

	return s.vm.ToValue(value)
}

// SetItem stores a key-value pair
func (s *Storage) SetItem(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return goja.Undefined()
	}

	key := call.Argument(0).String()
	value := call.Argument(1).String()

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO `+s.tableName+` (key, value) VALUES (?, ?)
	`, key, value)

	if err != nil {
		// Could throw an exception here
		return goja.Undefined()
	}

	return goja.Undefined()
}

// RemoveItem removes a key
func (s *Storage) RemoveItem(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	key := call.Argument(0).String()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("DELETE FROM "+s.tableName+" WHERE key = ?", key)

	return goja.Undefined()
}

// Clear removes all items
func (s *Storage) Clear(call goja.FunctionCall) goja.Value {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("DELETE FROM " + s.tableName)

	return goja.Undefined()
}

// Key returns the key at the given index
func (s *Storage) Key(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Null()
	}

	index := call.Argument(0).ToInteger()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var key string
	err := s.db.QueryRow("SELECT key FROM "+s.tableName+" LIMIT 1 OFFSET ?", index).Scan(&key)
	if err != nil {
		return goja.Null()
	}

	return s.vm.ToValue(key)
}

// Length returns the number of items
func (s *Storage) Length() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM " + s.tableName).Scan(&count)
	return count
}

// ToJSObject creates a JS object with all Storage methods
func (s *Storage) ToJSObject() *goja.Object {
	obj := s.vm.NewObject()
	obj.Set("getItem", s.GetItem)
	obj.Set("setItem", s.SetItem)
	obj.Set("removeItem", s.RemoveItem)
	obj.Set("clear", s.Clear)
	obj.Set("key", s.Key)

	// length as a property (getter)
	obj.DefineAccessorProperty("length",
		s.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return s.vm.ToValue(s.Length())
		}),
		goja.Undefined(),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	return obj
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}
