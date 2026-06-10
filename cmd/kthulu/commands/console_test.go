package commands

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func newTestConsole(t *testing.T) (*console, *bytes.Buffer) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "console.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price INTEGER)`); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	return &console{db: db, driver: "sqlite", out: out}, out
}

func TestConsoleCRUDRoundTrip(t *testing.T) {
	c, out := newTestConsole(t)

	steps := []struct {
		cmd  string
		want string
	}{
		{"create products name=Widget price=5", "1 row(s) affected"},
		{"create products name=Gadget price=9", "1 row(s) affected"},
		{"list products", "Widget"},
		{"count products", "2"},
		{"update products 1 price=42", "1 row(s) affected"},
		{"find products 1", "42"},
		{"sql SELECT name FROM products WHERE id = 2", "Gadget"},
		{"delete products 2", "1 row(s) affected"},
		{"count products", "1"},
		{"tables", "products"},
		{"schema products", "price"},
	}
	for _, step := range steps {
		out.Reset()
		if err := c.dispatch(step.cmd); err != nil {
			t.Fatalf("%q: %v", step.cmd, err)
		}
		if !strings.Contains(out.String(), step.want) {
			t.Fatalf("%q: output %q does not contain %q", step.cmd, out.String(), step.want)
		}
	}
}

func TestConsoleRejectsBadIdentifiers(t *testing.T) {
	c, _ := newTestConsole(t)
	for _, cmd := range []string{
		"list products;DROP",
		"find products--x 1",
		"create products name=ok bad-col=1",
		"delete 'products' 1",
	} {
		if err := c.dispatch(cmd); err == nil {
			t.Fatalf("%q: expected identifier error", cmd)
		}
	}
}

func TestConsoleUnknownCommand(t *testing.T) {
	c, _ := newTestConsole(t)
	if err := c.dispatch("frobnicate things"); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
}
