package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pmaojo/kthulu-go/internal/blueprint"

	// Database drivers registered for database/sql so the console can
	// connect to any database a generated project uses.
	_ "github.com/go-sql-driver/mysql" // mysql driver
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver (pgx)
	_ "modernc.org/sqlite"             // sqlite driver (CGO-free)
)

var (
	consoleDSN    string
	consoleDriver string
	consoleExec   string
)

const consoleHelp = `Open an interactive console connected to the project database.

Run it from a generated project directory: the connection is discovered the
same way the generated app connects (DATABASE_URL, kthulu-plan.yaml, env).

Commands:
  tables                          List tables
  schema <table>                  Show columns of a table
  list <table> [limit]            Show rows (default limit 20)
  find <table> <id>               Show one row by id
  count <table>                   Count rows
  create <table> col=val ...      Insert a row
  update <table> <id> col=val ... Update a row by id
  delete <table> <id>             Delete a row by id
  sql <statement>                 Run raw SQL (SELECT prints rows)
  help                            Show this help
  exit                            Leave the console`

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "🔮 Interactive database console for your project (tinker equivalent)",
	Long:  consoleHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, dsn, err := resolveConsoleDB()
		if err != nil {
			return err
		}
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()
		if driver == "sqlite" {
			db.SetMaxOpenConns(1)
		}
		if err := db.Ping(); err != nil {
			return fmt.Errorf("connect to database (%s): %w", driver, err)
		}

		c := &console{db: db, driver: driver, out: cmd.OutOrStdout()}
		if consoleExec != "" {
			return c.dispatch(consoleExec)
		}

		fmt.Fprintf(c.out, "🔮 kthulu console — connected (%s). Type 'help' for commands, 'exit' to quit.\n", driver)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for {
			fmt.Fprint(c.out, "kthulu> ")
			if !scanner.Scan() {
				return nil
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				return nil
			}
			if err := c.dispatch(line); err != nil {
				fmt.Fprintf(c.out, "❌ %v\n", err)
			}
		}
	},
}

func init() {
	consoleCmd.Flags().StringVar(&consoleDSN, "dsn", "", "Database DSN override (skips auto-discovery)")
	consoleCmd.Flags().StringVar(&consoleDriver, "driver", "", "Database driver override: sqlite, postgres, mysql")
	consoleCmd.Flags().StringVarP(&consoleExec, "exec", "e", "", "Run a single console command and exit")
}

// resolveConsoleDB discovers the project's database connection using the same
// precedence as the generated application: explicit flags, DATABASE_URL, the
// project blueprint plus standard env vars, then the default SQLite path.
func resolveConsoleDB() (driver, dsn string, err error) {
	if consoleDSN != "" {
		return explicitConsoleDriver(), consoleDSN, nil
	}

	if url := os.Getenv("DATABASE_URL"); url != "" {
		return "pgx", url, nil
	}

	dbType, projectName := projectDBSettings()
	return buildConsoleDSN(dbType, projectName)
}

// explicitConsoleDriver normalizes the --driver flag for database/sql.
func explicitConsoleDriver() string {
	switch consoleDriver {
	case "":
		return "sqlite"
	case "postgres":
		return "pgx"
	default:
		return consoleDriver
	}
}

// projectDBSettings reads the database type and project name from
// kthulu-plan.yaml, falling back to sqlite and the directory name.
func projectDBSettings() (dbType, projectName string) {
	dbType, projectName = "sqlite", "app"
	data, err := os.ReadFile("kthulu-plan.yaml")
	if err != nil {
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			projectName = filepath.Base(cwd)
		}
		return dbType, projectName
	}
	var bp blueprint.ProjectBlueprint
	if yaml.Unmarshal(data, &bp) == nil {
		if bp.Database != "" {
			dbType = bp.Database
		}
		if bp.Name != "" {
			projectName = bp.Name
		}
	}
	return dbType, projectName
}

// buildConsoleDSN assembles the DSN for the project database from standard
// environment variables, mirroring the generated app's providers.
func buildConsoleDSN(dbType, projectName string) (driver, dsn string, err error) {
	envOr := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	switch dbType {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			envOr("DB_HOST", "localhost"), envOr("DB_PORT", "5432"),
			envOr("DB_USER", "postgres"), envOr("DB_PASSWORD", "postgres"),
			envOr("DB_NAME", projectName))
		return "pgx", dsn, nil
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			envOr("DB_USER", "root"), envOr("DB_PASSWORD", "password"),
			envOr("DB_HOST", "localhost"), envOr("DB_PORT", "3306"),
			envOr("DB_NAME", projectName))
		return "mysql", dsn, nil
	default:
		path := envOr("SQLITE_PATH", filepath.Join("data", projectName+".db"))
		if _, statErr := os.Stat(path); statErr != nil {
			return "", "", fmt.Errorf("sqlite database not found at %s (run the app once to create it, or pass --dsn)", path)
		}
		return "sqlite", path, nil
	}
}

type console struct {
	db     *sql.DB
	driver string
	out    interface{ Write([]byte) (int, error) }
}

var identPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func (c *console) ident(s string) (string, error) {
	if !identPattern.MatchString(s) {
		return "", fmt.Errorf("invalid identifier %q", s)
	}
	return s, nil
}

func (c *console) placeholder(n int) string {
	if c.driver == "pgx" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

func (c *console) dispatch(line string) error {
	fields := strings.Fields(line)
	cmd, args := fields[0], fields[1:]

	switch strings.ToLower(cmd) {
	case "help":
		fmt.Fprintln(c.out, consoleHelp)
		return nil
	case "tables":
		return c.tables()
	case "schema":
		return c.cmdSchema(args)
	case "list", "all":
		return c.cmdList(args)
	case "find":
		return c.cmdFind(args)
	case "count":
		return c.cmdCount(args)
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("usage: create <table> col=val ...")
		}
		return c.create(args[0], args[1:])
	case "update":
		if len(args) < 3 {
			return fmt.Errorf("usage: update <table> <id> col=val ...")
		}
		return c.update(args[0], args[1], args[2:])
	case "delete":
		return c.cmdDelete(args)
	case "sql":
		return c.cmdSQL(strings.TrimSpace(strings.TrimPrefix(line, cmd)))
	default:
		return fmt.Errorf("unknown command %q (try 'help')", cmd)
	}
}

func (c *console) cmdSchema(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: schema <table>")
	}
	return c.schema(args[0])
}

var limitPattern = regexp.MustCompile(`^\d+$`)

func (c *console) cmdList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: list <table> [limit]")
	}
	limit := "20"
	if len(args) > 1 {
		limit = args[1]
	}
	if !limitPattern.MatchString(limit) {
		return fmt.Errorf("limit must be a number")
	}
	table, err := c.ident(args[0])
	if err != nil {
		return err
	}
	return c.query(fmt.Sprintf("SELECT * FROM %s LIMIT %s", table, limit))
}

func (c *console) cmdFind(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: find <table> <id>")
	}
	table, err := c.ident(args[0])
	if err != nil {
		return err
	}
	return c.queryArgs(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", table, c.placeholder(1)), args[1])
}

func (c *console) cmdCount(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: count <table>")
	}
	table, err := c.ident(args[0])
	if err != nil {
		return err
	}
	return c.query(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", table))
}

func (c *console) cmdDelete(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: delete <table> <id>")
	}
	table, err := c.ident(args[0])
	if err != nil {
		return err
	}
	return c.exec(fmt.Sprintf("DELETE FROM %s WHERE id = %s", table, c.placeholder(1)), args[1])
}

func (c *console) cmdSQL(stmt string) error {
	if stmt == "" {
		return fmt.Errorf("usage: sql <statement>")
	}
	lowered := strings.ToLower(stmt)
	for _, prefix := range []string{"select", "show", "pragma", "explain"} {
		if strings.HasPrefix(lowered, prefix) {
			return c.query(stmt)
		}
	}
	return c.exec(stmt)
}

func (c *console) tables() error {
	switch c.driver {
	case "pgx":
		return c.query("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	case "mysql":
		return c.query("SHOW TABLES")
	default:
		return c.query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	}
}

func (c *console) schema(table string) error {
	t, err := c.ident(table)
	if err != nil {
		return err
	}
	switch c.driver {
	case "pgx":
		return c.queryArgs("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position", t)
	case "mysql":
		return c.queryArgs("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = ? AND table_schema = DATABASE() ORDER BY ordinal_position", t)
	default:
		return c.query(fmt.Sprintf("PRAGMA table_info(%s)", t))
	}
}

func (c *console) create(table string, pairs []string) error {
	t, err := c.ident(table)
	if err != nil {
		return err
	}
	cols, vals, err := c.parsePairs(pairs)
	if err != nil {
		return err
	}
	holders := make([]string, len(cols))
	for i := range cols {
		holders[i] = c.placeholder(i + 1)
	}
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(cols, ", "), strings.Join(holders, ", "))
	return c.exec(stmt, vals...)
}

func (c *console) update(table, id string, pairs []string) error {
	t, err := c.ident(table)
	if err != nil {
		return err
	}
	cols, vals, err := c.parsePairs(pairs)
	if err != nil {
		return err
	}
	sets := make([]string, len(cols))
	for i, col := range cols {
		sets[i] = fmt.Sprintf("%s = %s", col, c.placeholder(i+1))
	}
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE id = %s", t, strings.Join(sets, ", "), c.placeholder(len(cols)+1))
	return c.exec(stmt, append(vals, id)...)
}

func (c *console) parsePairs(pairs []string) (cols []string, vals []interface{}, err error) {
	for _, p := range pairs {
		col, val, ok := strings.Cut(p, "=")
		if !ok {
			return nil, nil, fmt.Errorf("expected col=val, got %q", p)
		}
		if _, err := c.ident(col); err != nil {
			return nil, nil, err
		}
		cols = append(cols, col)
		vals = append(vals, val)
	}
	return cols, vals, nil
}

func (c *console) exec(stmt string, args ...interface{}) error {
	res, err := c.db.Exec(stmt, args...)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	fmt.Fprintf(c.out, "✅ OK (%d row(s) affected)\n", affected)
	return nil
}

func (c *console) query(stmt string) error {
	return c.queryArgs(stmt)
}

func (c *console) queryArgs(stmt string, args ...interface{}) error {
	rows, err := c.db.Query(stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var table [][]string
	for rows.Next() {
		raw := make([]sql.RawBytes, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		row := make([]string, len(cols))
		for i, v := range raw {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = string(v)
			}
		}
		table = append(table, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	c.printTable(cols, table)
	return nil
}

func (c *console) printTable(cols []string, rows [][]string) {
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > 60 {
				cell = cell[:57] + "..."
				row[i] = cell
			}
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	line := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, cell := range cells {
			parts[i] = fmt.Sprintf("%-*s", widths[i], cell)
		}
		return strings.Join(parts, " | ")
	}
	fmt.Fprintln(c.out, line(cols))
	sep := make([]string, len(cols))
	for i := range cols {
		sep[i] = strings.Repeat("-", widths[i])
	}
	fmt.Fprintln(c.out, strings.Join(sep, "-+-"))
	for _, row := range rows {
		fmt.Fprintln(c.out, line(row))
	}
	fmt.Fprintf(c.out, "(%d row(s))\n", len(rows))
}
