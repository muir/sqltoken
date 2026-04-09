package sqltoken

type BeginEndWordMode int

const (
	// BeginEndContextual treats BEGIN/END as context-dependent words that may be identifiers.
	//
	// WARNING: contextual detection is heuristic, not a full SQL parser. It is good for common
	// procedural SQL, but bareword identifiers named begin/end can still be ambiguous and may be
	// misclassified in edge cases. Prefer quoted identifiers (`begin`, `end`) when possible.
	BeginEndContextual BeginEndWordMode = iota
	// BeginEndReserved treats BEGIN/END as reserved keywords unless explicitly quoted/escaped.
	BeginEndReserved
)

// Config specifies the behavior of Tokenize as relates to behavior
// that differs between SQL implementations
type Config struct {
	// Tokenize ? as type Question (used by MySQL, SQLite)
	NoticeQuestionMark bool

	// Tokenize ?<digits> (eg: "?7") as type Question (used by SQLite)
	NoticeQuestionNumber bool

	// Tokenize $<digits> (eg "$7") as type DollarNumber (PostgreSQL, SQLite)
	NoticeDollarNumber bool

	// Tokenize :word as type ColonWord (sqlx, Oracle, SQLite)
	NoticeColonWord bool

	// Tokenize :word with unicode as ColonWord (sqlx, SQLite)
	ColonWordIncludesUnicode bool

	// Tokenize # as type comment (MySQL)
	NoticeHashComment bool

	// $q$ stuff $q$ and $$stuff$$ quoting (PostgreSQL)
	NoticeDollarQuotes bool

	// NoticeHexValues 0xa0 x'af' X'AF' (MySQL)
	NoticeHexNumbers bool

	// NoticeLiteralBackslashEscape 'escape \' quote' (MySQL by default,
	// PostgreSQL optional).
	NoticeLiteralBackslashEscape bool

	// NoticeBinaryValues 0x01 b'01' B'01' (MySQL)
	NoticeBinaryNumbers bool

	// NoticeUAmpPrefix U& utf prefix U&"\0441\043B\043E\043D" (PostgreSQL)
	NoticeUAmpPrefix bool

	// NoticeCharsetLiteral _latin1'string' (MySQL, SingleStore)
	NoticeCharsetLiteral bool

	// NoticeNationalPrefix n'string' N'string' (MySQL)
	NoticeNationalPrefix bool

	// NoticeNotionalStrings [nN]'...''...' (Oracle, SQL Server)
	NoticeNotionalStrings bool

	// NoticeEscapedStrings [eE]'...' (PostgreSQL)
	NoticeEscapedStrings bool

	// NoticeDelimitedStrings [nN]?[qQ]'DELIM .... DELIM' (Oracle)
	NoticeDelimitedStrings bool

	// NoticeTypedNumbers nn.nnEnn[fFdD] (Oracle)
	NoticeTypedNumbers bool

	// NoticeMoneyConstants $10 $10.32 (SQL Server)
	NoticeMoneyConstants bool

	// NoticeAtWord @foo (SQL Server, SQLite)
	NoticeAtWord bool

	// NoticeAtIdentifiers _baz @fo$o @@b#ar #foo ##b@ar (SQL Server)
	NoticeIdentifiers bool

	// SeparatePunctuation prevents merging successive punctuation into a single token
	SeparatePunctuation bool

	// NoticeDelimiterStatement turns on recognition of custom statement delimiters
	// This is not 100%: the delimiter statement is always recognized after a newline even
	// if that's in the middle of a block that the client knows isn't the start of a
	// statement
	NoticeDelimiterStatement bool

	// NoticeBeginEndBlock tracks BEGIN/END block nesting for stored procedures,
	// functions, and triggers. When inside a block, semicolons are treated as
	// punctuation rather than statement delimiters. This is useful when DELIMITER
	// is not available (e.g., when sending SQL directly to database/sql drivers).
	// This option is mutually exclusive with NoticeDelimiterStatement in practice:
	// use NoticeDelimiterStatement for parsing SQL files that use DELIMITER syntax,
	// use NoticeBeginEndBlock for SQL going directly to drivers.
	NoticeBeginEndBlock bool

	// BeginEndWordMode controls whether BEGIN/END are interpreted contextually
	// (e.g. MySQL/MariaDB) or as reserved keywords (e.g. SingleStore).
	BeginEndWordMode BeginEndWordMode
}

// WithNoticeQuestionMark enables parsing question marks using the QuestionMark token
func (c Config) WithNoticeQuestionMark() Config {
	c.NoticeQuestionMark = true
	return c
}

// WithNoticeQuestionNumber enables parsing numbered question marks (?<digits>) using the QuestionMark token.
func (c Config) WithNoticeQuestionNumber() Config {
	c.NoticeQuestionNumber = true
	return c
}

// WithNoticeDollarNumber enables parsing dollar parameters ($1) for PostgreSQL
// and SQLite using the DollarNumber token
func (c Config) WithNoticeDollarNumber() Config {
	c.NoticeDollarNumber = true
	return c
}

// WithNoticeColonWord enables parsing for named parameters using the ColonWord token
func (c Config) WithNoticeColonWord() Config {
	c.NoticeColonWord = true
	return c
}

// WithColonWordIncludesUnicode enables unicode name parsing at a small performance cost
func (c Config) WithColonWordIncludesUnicode() Config {
	c.ColonWordIncludesUnicode = true
	return c
}

// WithNoticeHashComment enables parsing for '#' comments (MySQL) using the Punctuation token
func (c Config) WithNoticeHashComment() Config {
	c.NoticeHashComment = true
	return c
}

// WithNoticeDollarQuotes enables dollar quote $$parsing$$ for PostgreSQL using the DollarNumber token
func (c Config) WithNoticeDollarQuotes() Config {
	c.NoticeDollarQuotes = true
	return c
}

// WithNoticeHexNumbers enables quoted hex number parsing (x'af') using the HexNumber token (MySQL)
func (c Config) WithNoticeHexNumbers() Config {
	c.NoticeHexNumbers = true
	return c
}

// WithNoticeLiteralBackslashEscape enables 'escape \' quote' (MySQL by default,
// PostgreSQL optional).
func (c Config) WithNoticeLiteralBackslashEscape() Config {
	c.NoticeLiteralBackslashEscape = true
	return c
}

// WithNoticeBinaryNumbers enables quoted binary number parsing (b'01') using the BinaryNumber token (MySQL)
func (c Config) WithNoticeBinaryNumbers() Config {
	c.NoticeBinaryNumbers = true
	return c
}

// WithNoticeUAmpPrefix enables U& prefix parsing (U&"\0441\043B\043E\043D") using the Literal token (PostgreSQL)
func (c Config) WithNoticeUAmpPrefix() Config {
	c.NoticeUAmpPrefix = true
	return c
}

// WithNoticeCharsetLiteral enables charset literal parsing (_latin1'string') using the Literal token (MySQL, SingleStore)
func (c Config) WithNoticeCharsetLiteral() Config {
	c.NoticeCharsetLiteral = true
	return c
}

// WithNoticeNationalPrefix enables national string prefix parsing (n'string' N'string') using the Literal token (MySQL)
func (c Config) WithNoticeNationalPrefix() Config {
	c.NoticeNationalPrefix = true
	return c
}

// WithNoticeNotionalStrings enables notional string parsing (n'string') using the Literal token (Oracle, SQL Server)
func (c Config) WithNoticeNotionalStrings() Config {
	c.NoticeNotionalStrings = true
	return c
}

// WithNoticeEscapedStrings enables escaped string parsing (E'string') using the Literal token (PostgreSQL)
func (c Config) WithNoticeEscapedStrings() Config {
	c.NoticeEscapedStrings = true
	return c
}

// WithNoticeDelimitedStrings enables deliminated string parsing (q'DELIM .... DELIM') using the Literal token (Oracle)
func (c Config) WithNoticeDelimitedStrings() Config {
	c.NoticeDelimitedStrings = true
	return c
}

// WithNoticeTypedNumbers enables typed number parsing (nn.nnEnn[fFdD]) using the Number token (Oracle)
func (c Config) WithNoticeTypedNumbers() Config {
	c.NoticeTypedNumbers = true
	return c
}

// WithNoticeMoneyConstants enables money constant parsing ($10 $10.32) using the DollarNumber token (SQL Server)
func (c Config) WithNoticeMoneyConstants() Config {
	c.NoticeMoneyConstants = true
	return c
}

// WithNoticeAtWord enables parsing for '@foo' (SQL Server) using the AtWord token
func (c Config) WithNoticeAtWord() Config {
	c.NoticeAtWord = true
	return c
}

// WithNoticeIdentifiers enables parsing for identifiers (SQL Server) using the Identifier token
func (c Config) WithNoticeIdentifiers() Config {
	c.NoticeIdentifiers = true
	return c
}

// WithSeparatePunctuation enables separating successive punctuation into separate tokens
func (c Config) WithSeparatePunctuation() Config {
	c.SeparatePunctuation = true
	return c
}

// WithNoticeDelimiterStatement enables recognition of custom statement delimiters
// (e.g., DELIMITER // ... DELIMITER ;). This is mutually exclusive with
// NoticeBeginEndBlock - enabling this disables BEGIN/END block tracking.
func (c Config) WithNoticeDelimiterStatement() Config {
	c.NoticeDelimiterStatement = true
	c.NoticeBeginEndBlock = false
	return c
}

// WithNoticeBeginEndBlock enables tracking of BEGIN/END blocks for stored
// procedures, functions, and triggers. Semicolons inside blocks are treated
// as punctuation rather than statement delimiters. This is mutually exclusive
// with NoticeDelimiterStatement - enabling this disables DELIMITER recognition.
func (c Config) WithNoticeBeginEndBlock() Config {
	c.NoticeBeginEndBlock = true
	c.NoticeDelimiterStatement = false
	return c
}

// WithBeginEndWordMode sets how BEGIN/END words are interpreted while block
// tracking is enabled.
//
// For BeginEndContextual, use quoted identifiers if you have tables/columns/variables named
// begin/end and need deterministic behavior in all edge cases.
func (c Config) WithBeginEndWordMode(mode BeginEndWordMode) Config {
	c.BeginEndWordMode = mode
	return c
}

func (c Config) combineOkay(t TokenType) bool {
	// nolint:exhaustive
	switch t {
	case Number, QuestionMark, DollarNumber, ColonWord, Delimiter, DelimiterStatement:
		return false
	case Punctuation:
		return !c.SeparatePunctuation
	}
	return true
}

// OracleConfig returns a parsing configuration that is appropriate
// for parsing Oracle's SQL
func OracleConfig() Config {
	// https://docs.oracle.com/en/database/oracle/oracle-database/19/sqlrf/Literals.html
	return Config{}.
		WithNoticeNotionalStrings().
		WithNoticeDelimitedStrings().
		WithNoticeTypedNumbers().
		WithNoticeColonWord()
}

// SQLServerConfig returns a parsing configuration that is appropriate
// for parsing SQLServer's SQL
func SQLServerConfig() Config {
	return Config{}.
		WithNoticeNotionalStrings().
		WithNoticeHexNumbers().
		WithNoticeMoneyConstants().
		WithNoticeAtWord().
		WithNoticeIdentifiers()
}

// MySQLConfig returns a parsing configuration that is appropriate
// for parsing MySQL and MariaDB SQL. This includes support for DELIMITER and
// is compatible with the mysql client.
func MySQLConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeHashComment().
		WithNoticeHexNumbers().
		WithNoticeBinaryNumbers().
		WithNoticeCharsetLiteral().
		WithNoticeNationalPrefix().
		WithNoticeLiteralBackslashEscape().
		WithNoticeDelimiterStatement()
}

// MySQLAPIConfig returns a parsing configuration that is appropriate
// for parsing MySQL and MariaDB SQL sent through driver APIs. This enables
// BEGIN/END block tracking with contextual begin/end detection.
//
// WARNING: contextual begin/end handling is heuristic (not a full parser).
// If your SQL uses bareword identifiers named begin/end, quote them for
// predictable tokenization in all cases.
func MySQLAPIConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeHashComment().
		WithNoticeHexNumbers().
		WithNoticeBinaryNumbers().
		WithNoticeCharsetLiteral().
		WithNoticeNationalPrefix().
		WithNoticeLiteralBackslashEscape().
		WithNoticeBeginEndBlock().
		WithBeginEndWordMode(BeginEndContextual)
}

// SingleStoreConfig returns a parsing configuration that is appropriate
// for parsing SingleStore SQL. This includes support for DELIMITER and
// is compatible with the mysql/singlestore client.
func SingleStoreConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeHashComment().
		WithNoticeHexNumbers().
		WithNoticeBinaryNumbers().
		WithNoticeCharsetLiteral().
		WithNoticeLiteralBackslashEscape().
		WithNoticeEscapedStrings().
		WithNoticeDelimiterStatement()
}

// SingleStoreAPIConfig returns a parsing configuration that is appropriate
// for parsing SingleStore SQL sent through driver APIs. This enables
// BEGIN/END block tracking (instead of DELIMITER statement handling).
func SingleStoreAPIConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeHashComment().
		WithNoticeHexNumbers().
		WithNoticeBinaryNumbers().
		WithNoticeCharsetLiteral().
		WithNoticeLiteralBackslashEscape().
		WithNoticeEscapedStrings().
		WithNoticeBeginEndBlock().
		WithBeginEndWordMode(BeginEndReserved)
}

// PostgreSQLConfig returns a parsing configuration that is appropriate
// for parsing PostgreSQL and CockroachDB SQL.
func PostgreSQLConfig() Config {
	return Config{}.
		WithNoticeDollarNumber().
		WithNoticeDollarQuotes().
		WithNoticeUAmpPrefix().
		WithNoticeNationalPrefix().
		WithNoticeEscapedStrings()
}

// SQLiteConfig returns a parsing configuration that is appropriate for SQLite SQL.
func SQLiteConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeQuestionNumber().
		WithNoticeColonWord().
		WithColonWordIncludesUnicode().
		WithNoticeAtWord().
		WithNoticeDollarNumber()
}

// TokenizeMySQL breaks up MySQL / MariaDB strings into
// Token objects.
func TokenizeMySQL(s string) Tokens {
	return Tokenize(s, MySQLConfig())
}

// TokenizeMySQLAPI breaks up MySQL / MariaDB strings into
// Token objects. Uses BEGIN/END block tracking for stored procedures.
func TokenizeMySQLAPI(s string) Tokens {
	return Tokenize(s, MySQLAPIConfig())
}

// TokenizeSingleStore breaks up SingleStore SQL strings into
// Token objects.
func TokenizeSingleStore(s string) Tokens {
	return Tokenize(s, SingleStoreConfig())
}

// TokenizeSingleStoreAPI breaks up SingleStore SQL strings into
// Token objects. Uses BEGIN/END block tracking for stored procedures.
func TokenizeSingleStoreAPI(s string) Tokens {
	return Tokenize(s, SingleStoreAPIConfig())
}

// TokenizePostgreSQL breaks up PostgreSQL / CockroachDB SQL strings into
// Token objects.
func TokenizePostgreSQL(s string) Tokens {
	return Tokenize(s, PostgreSQLConfig())
}

// TokenizeSQLite breaks up SQLite strings into Token objects.
func TokenizeSQLite(s string) Tokens {
	return Tokenize(s, SQLiteConfig())
}
