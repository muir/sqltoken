package sqltoken

import (
	"strings"
	"testing"
)

func benchmarkTokenize(b *testing.B, cfg Config, sql string) {
	b.Helper()
	b.ReportAllocs()
	b.SetBytes(int64(len(sql)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(sql, cfg)
	}
}

func comparableDialectSQL(dialect string) string {
	base := strings.Repeat(
		"SELECT {{p1}}, {{p2}}, {{p3}} FROM {{table}} WHERE {{col}} = {{p1}};\n"+
			"UPDATE {{table}} SET {{col}} = {{p2}} WHERE id = {{p3}};\n"+
			"INSERT INTO {{table}}(id, {{col}}) VALUES ({{p3}}, {{str}});\n", 36)

	repl := []string{
		"{{p1}}", "?",
		"{{p2}}", "?",
		"{{p3}}", "?",
		"{{table}}", "t",
		"{{col}}", "c",
		"{{str}}", "'payload_95001'",
	}
	switch dialect {
	case "postgres":
		repl = []string{
			"{{p1}}", "$1",
			"{{p2}}", "$2",
			"{{p3}}", "$3",
			"{{table}}", "t",
			"{{col}}", "c",
			"{{str}}", "'payload_95001'",
		}
	case "sqlite":
		repl = []string{
			"{{p1}}", "?1",
			"{{p2}}", ":p2",
			"{{p3}}", "$3",
			"{{table}}", "t",
			"{{col}}", "c",
			"{{str}}", "'payload_95001'",
		}
	case "oracle":
		repl = []string{
			"{{p1}}", ":p1",
			"{{p2}}", ":p2",
			"{{p3}}", ":p3",
			"{{table}}", "t",
			"{{col}}", "c",
			"{{str}}", "'payload_95001'",
		}
	case "sqlserver":
		repl = []string{
			"{{p1}}", "@p1",
			"{{p2}}", "@p2",
			"{{p3}}", "@p3",
			"{{table}}", "[t]",
			"{{col}}", "[c]",
			"{{str}}", "'payload_95001'",
		}
	}
	return strings.NewReplacer(repl...).Replace(base)
}

func BenchmarkTokenizeDialectModes(b *testing.B) {
	mysqlFileSQL := strings.Repeat(
		"DELIMITER //\n"+
			"CREATE PROCEDURE p_mysql_file()\n"+
			"BEGIN\n"+
			"  INSERT INTO t(id, note) VALUES (1, 'mysql');\n"+
			"  UPDATE t SET note = CONCAT(note, '_x') WHERE id = 1;\n"+
			"END//\n"+
			"DELIMITER ;\n"+
			"SELECT 94001;\n", 16)
	mysqlAPISQL := strings.Repeat(
		"CREATE PROCEDURE p_mysql_api()\n"+
			"BEGIN\n"+
			"  SET @begin = 1;\n"+
			"  SELECT end FROM t WHERE t.end = @begin;\n"+
			"  SELECT 94002 AS begin FROM t;\n"+
			"END;\n", 24)
	singleStoreFileSQL := strings.Repeat(
		"DELIMITER //\n"+
			"CREATE PROCEDURE p_s2_file() AS BEGIN\n"+
			"  INSERT INTO t VALUES (1);\n"+
			"  SELECT 94003;\n"+
			"END//\n"+
			"DELIMITER ;\n", 20)
	singleStoreAPISQL := strings.Repeat(
		"CREATE PROCEDURE p_s2_api() AS BEGIN\n"+
			"  SET @begin = 1;\n"+
			"  SELECT 94004;\n"+
			"END;\n", 28)
	postgresSQL := strings.Repeat(
		"SELECT $1::int, $2::text;\n"+
			"SELECT $tag$postgres_payload$tag$;\n"+
			"SELECT E'line\\ntext';\n"+
			"SELECT U&'d\\0061t\\+000061';\n", 30)
	oracleSQL := strings.Repeat(
		"SELECT q'[oracle_payload]' AS col FROM dual;\n"+
			"SELECT :bind_name, :other_94007 FROM dual;\n"+
			"SELECT 1.25e+3f FROM dual;\n", 34)
	sqlServerSQL := strings.Repeat(
		"SELECT @var_94008, @@spid;\n"+
			"SELECT $10.25 AS money_val;\n"+
			"SELECT [name] FROM [dbo].[t] WHERE id = 1;\n", 32)
	sqliteSQL := strings.Repeat(
		"SELECT ?1, :name, @var, $2;\n"+
			"UPDATE t SET c = :name WHERE id = ?2;\n"+
			"SELECT 94005;\n", 40)

	cases := []struct {
		name string
		cfg  Config
		sql  string
	}{
		{name: "mysql-file-delimiter", cfg: MySQLConfig(), sql: mysqlFileSQL},
		{name: "mysql-api-contextual", cfg: MySQLAPIConfig(), sql: mysqlAPISQL},
		{name: "singlestore-file-delimiter", cfg: SingleStoreConfig(), sql: singleStoreFileSQL},
		{name: "singlestore-api-reserved", cfg: SingleStoreAPIConfig(), sql: singleStoreAPISQL},
		{name: "postgres", cfg: PostgreSQLConfig(), sql: postgresSQL},
		{name: "oracle", cfg: OracleConfig(), sql: oracleSQL},
		{name: "sqlserver", cfg: SQLServerConfig(), sql: sqlServerSQL},
		{name: "sqlite", cfg: SQLiteConfig(), sql: sqliteSQL},
	}
	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			benchmarkTokenize(b, tc.cfg, tc.sql)
		})
	}
}

func BenchmarkTokenizeDialectModesComparable(b *testing.B) {
	cases := []struct {
		name string
		cfg  Config
		sql  string
	}{
		{name: "mysql-file-delimiter", cfg: MySQLConfig(), sql: comparableDialectSQL("mysql")},
		{name: "mysql-api-contextual", cfg: MySQLAPIConfig(), sql: comparableDialectSQL("mysql")},
		{name: "singlestore-file-delimiter", cfg: SingleStoreConfig(), sql: comparableDialectSQL("mysql")},
		{name: "singlestore-api-reserved", cfg: SingleStoreAPIConfig(), sql: comparableDialectSQL("mysql")},
		{name: "postgres", cfg: PostgreSQLConfig(), sql: comparableDialectSQL("postgres")},
		{name: "oracle", cfg: OracleConfig(), sql: comparableDialectSQL("oracle")},
		{name: "sqlserver", cfg: SQLServerConfig(), sql: comparableDialectSQL("sqlserver")},
		{name: "sqlite", cfg: SQLiteConfig(), sql: comparableDialectSQL("sqlite")},
	}
	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			benchmarkTokenize(b, tc.cfg, tc.sql)
		})
	}
}

func BenchmarkTokenizeBeginEndVsDelimiter(b *testing.B) {
	// Same routine body expressed once for DELIMITER mode and once for API mode so
	// BEGIN/END tracking and DELIMITER handling costs can be compared separately.
	mysqlDelimiterSQL := strings.Repeat(
		"DELIMITER //\n"+
			"CREATE PROCEDURE p_mode_delim()\n"+
			"BEGIN\n"+
			"  DECLARE end INT;\n"+
			"  SET end = 1;\n"+
			"  SELECT end FROM t WHERE t.end = end;\n"+
			"END//\n"+
			"DELIMITER ;\n", 24)
	mysqlBeginEndSQL := strings.Repeat(
		"CREATE PROCEDURE p_mode_be()\n"+
			"BEGIN\n"+
			"  DECLARE end INT;\n"+
			"  SET end = 1;\n"+
			"  SELECT end FROM t WHERE t.end = end;\n"+
			"END;\n", 28)
	singleStoreReservedSQL := strings.Repeat(
		"CREATE PROCEDURE p_mode_s2() AS BEGIN\n"+
			"  SET @begin = 1;\n"+
			"  SELECT 94006;\n"+
			"END;\n", 36)

	cases := []struct {
		name string
		cfg  Config
		sql  string
	}{
		{name: "mysql-delimiter-statements", cfg: MySQLConfig(), sql: mysqlDelimiterSQL},
		{name: "mysql-begin-end-contextual", cfg: MySQLAPIConfig(), sql: mysqlBeginEndSQL},
		{name: "singlestore-begin-end-reserved", cfg: SingleStoreAPIConfig(), sql: singleStoreReservedSQL},
	}
	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			benchmarkTokenize(b, tc.cfg, tc.sql)
		})
	}
}
