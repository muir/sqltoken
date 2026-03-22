package sqltoken

// Config specifies the behavior of Tokenize as relates to behavior
// that differs between SQL implementations
type Config struct {
	// Tokenize ? as type Question (used by MySQL, SQLite)
	NoticeQuestionMark bool

	// Tokenize ?1 as type Question; implies NoticeQuestionMark (used by SQLite)
	NoticeQuestionNumber bool

	// Tokenize $7 as type DollarNumber (PostgreSQL, SQLite)
	NoticeDollarNumber bool

	// Tokenize :word as type ColonWord (sqlx, Oracle, SQLite)
	NoticeColonWord bool

	// Tokenize :word with unicode as ColonWord (sqlx, SQlite)
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
}

// WithNoticeQuestionMark enables parsing question marks using the QuestionMark token
func (c Config) WithNoticeQuestionMark() Config {
	c.NoticeQuestionMark = true
	return c
}

// WithNoticeQuestionNumber enables parsing question marks followed by a number (?1) for SQLite.
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

func (c Config) WithNoticeDelimiterStatement() Config {
	c.NoticeDelimiterStatement = true
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
// for parsing MySQL and MariaDB SQL.
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

// SingleStoreConfig returns a parsing configuration that is appropriate
// for parsing SingleStore SQL.
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

// SQLiteConfig returns a parsing configuration that is appropriate for parsing
// SQLite SQL.
func SQLiteConfig() Config {
	return Config{}.
		WithNoticeQuestionNumber().
		WithNoticeQuestionMark().
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

// TokenizeSingleStore breaks up SingleStore SQL strings into
// Token objects.
func TokenizeSingleStore(s string) Tokens {
	return Tokenize(s, SingleStoreConfig())
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
