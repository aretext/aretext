package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

// SQLParseFunc returns a parse func for SQL.
func SQLParseFunc() parser.Func {
	return sqlLineCommentParseFunc().
		Or(sqlBlockCommentParseFunc()).
		Or(sqlStringLiteralParseFunc()).
		Or(sqlStringNameParseFunc()).
		Or(sqlIdentifierOrKeywordParseFunc()).
		Or(sqlOperatorParseFunc()).
		Or(sqlNumberLiteralParseFunc())
}

func sqlLineCommentParseFunc() parser.Func {
	return consumeString("--").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
}

func sqlBlockCommentParseFunc() parser.Func {
	return consumeString("/*").
		Then(consumeToString("*/")).
		Map(recognizeToken(parser.TokenRoleComment))
}

func sqlStringLiteralParseFunc() parser.Func {
	return consumeString("'").
		Then(consumeToString("'")).
		Map(recognizeToken(parser.TokenRoleString))
}

func sqlStringNameParseFunc() parser.Func {
	return consumeString("\"").
		Then(consumeToString("\"")).
		Map(recognizeToken(parser.TokenRoleCustom1))
}

// TODO: Case insensitive keywords
func sqlIdentifierOrKeywordParseFunc() parser.Func {
	isLetter := func(r rune) bool { return unicode.IsLetter(r) || r == '_' }
	isLetterOrDigit := func(r rune) bool { return isLetter(r) || unicode.IsDigit(r) }
	keywords := []string{
		"ABORT", "ABS", "ABSENT", "ABSOLUTE", "ACCESS", "ACCORDING", "ACOS",
		"ACTION", "ADA", "ADD", "ADMIN", "AFTER", "AGGREGATE", "ALL", "ALLOCATE",
		"ALSO", "ALTER", "ALWAYS", "ANALYSE", "ANALYZE", "AND", "ANY", "ANY_VALUE",
		"ARE", "ARRAY", "ARRAY_AGG", "AS", "ASC", "ASENSITIVE", "ASIN", "ASSERTION",
		"ASSIGNMENT", "ASYMMETRIC", "AT", "ATAN", "ATOMIC", "ATTACH", "ATTRIBUTE", "ATTRIBUTES",
		"AUTHORIZATION", "AVG", "BACKWARD", "BASE64", "BEFORE", "BEGIN", "BEGIN_FRAME", "BEGIN_PARTITION",
		"BERNOULLI", "BETWEEN", "BIGINT", "BINARY", "BIT", "BIT_LENGTH", "BLOB", "BLOCKED",
		"BOM", "BOOLEAN", "BOTH", "BREADTH", "BTRIM", "BY", "C", "CACHE",
		"CALL", "CALLED", "CARDINALITY", "CASCADE", "CASCADED", "CASE", "CAST", "CATALOG",
		"CATALOG_NAME", "CEIL", "CEILING", "CHAIN", "CHAINING", "CHAR", "CHARACTER", "CHARACTERISTICS",
		"CHARACTERS", "CHARACTER_LENGTH", "CHARACTER_SET_NAME", "CHARACTER_SET_SCHEMA", "CHAR_LENGTH", "CHECK", "CHECKPOINT", "CLASS",
		"CLASSIFIER", "CLASS_ORIGIN", "CLOB", "CLOSE", "CLUSTER", "COALESCE", "COBOL", "COLLATE",
		"COLLATION", "COLLATION_CATALOG", "COLLATION_NAME", "COLLATION_SCHEMA", "COLLECT", "COLUMN", "COLUMNS", "COLUMN_NAME",
		"COMMAND_FUNCTION", "COMMENT", "COMMENTS", "COMMIT", "COMMITTED", "COMPRESSION", "CONCURRENTLY", "CONDITION",
		"CONDITIONAL", "CONDITION_NUMBER", "CONFIGURATION", "CONFLICT", "CONNECT", "CONNECTION", "CONNECTION_NAME", "CONSTRAINT",
		"CONSTRAINTS", "CONSTRAINT_CATALOG", "CONSTRAINT_NAME", "CONSTRAINT_SCHEMA", "CONSTRUCTOR", "CONTAINS", "CONTENT", "CONTINUE",
		"CONTROL", "CONVERSION", "CONVERT", "COPARTITION", "COPY", "CORR", "CORRESPONDING", "COS",
		"COSH", "COST", "COUNT", "COVAR_POP", "COVAR_SAMP", "CREATE", "CROSS", "CSV",
		"CUBE", "CUME_DIST", "CURRENT", "CURRENT_CATALOG", "CURRENT_DATE", "CURRENT_PATH", "CURRENT_ROLE", "CURRENT_ROW",
		"CURRENT_SCHEMA", "CURRENT_TIME", "CURRENT_TIMESTAMP", "CURRENT_USER", "CURSOR", "CURSOR_NAME", "CYCLE", "DATA",
		"DATABASE", "DATALINK", "DATE", "DAY", "DB", "DEALLOCATE", "DEC", "DECFLOAT",
		"DECIMAL", "DECLARE", "DEFAULT", "DEFAULTS", "DEFERRABLE", "DEFERRED", "DEFINE", "DEFINED",
		"DEFINER", "DEGREE", "DELETE", "DELIMITER", "DELIMITERS", "DENSE_RANK", "DEPENDS", "DEPTH",
		"DEREF", "DERIVED", "DESC", "DESCRIBE", "DESCRIPTOR", "DETACH", "DETERMINISTIC", "DIAGNOSTICS",
		"DICTIONARY", "DISABLE", "DISCARD", "DISCONNECT", "DISPATCH", "DISTINCT", "DLNEWCOPY", "DLPREVIOUSCOPY",
		"DLURLCOMPLETE", "DLURLCOMPLETEONLY", "DLURLCOMPLETEWRITE", "DLURLPATH", "DLURLPATHONLY", "DLURLPATHWRITE", "DLURLSCHEME", "DLURLSERVER",
		"DLVALUE", "DO", "DOCUMENT", "DOMAIN", "DOUBLE", "DROP", "DYNAMIC", "DYNAMIC_FUNCTION",
		"EACH", "ELEMENT", "ELSE", "EMPTY", "ENABLE", "ENCODING", "ENCRYPTED", "END",
		"END-EXEC", "END_FRAME", "END_PARTITION", "ENFORCED", "ENUM", "EQUALS", "ERROR", "ESCAPE",
		"EVENT", "EVERY", "EXCEPT", "EXCEPTION", "EXCLUDE", "EXCLUDING", "EXCLUSIVE", "EXEC",
		"EXECUTE", "EXISTS", "EXP", "EXPLAIN", "EXPRESSION", "EXTENSION", "EXTERNAL", "EXTRACT",
		"FALSE", "FAMILY", "FETCH", "FILE", "FILTER", "FINAL", "FINALIZE", "FINISH",
		"FIRST", "FIRST_VALUE", "FLAG", "FLOAT", "FLOOR", "FOLLOWING", "FOR", "FORCE",
		"FOREIGN", "FORMAT", "FORTRAN", "FORWARD", "FOUND", "FRAME_ROW", "FREE", "FREEZE",
		"FROM", "FS", "FULFILL", "FULL", "FUNCTION", "FUNCTIONS", "FUSION", "G",
		"GENERAL", "GENERATED", "GET", "GLOBAL", "GO", "GOTO", "GRANT", "GRANTED",
		"GREATEST", "GROUP", "GROUPING", "GROUPS", "HANDLER", "HAVING", "HEADER", "HEX",
		"HIERARCHY", "HOLD", "HOUR", "ID", "IDENTITY", "IF", "IGNORE", "ILIKE",
		"IMMEDIATE", "IMMEDIATELY", "IMMUTABLE", "IMPLEMENTATION", "IMPLICIT", "IMPORT", "IN", "INCLUDE",
		"INCLUDING", "INCREMENT", "INDENT", "INDEX", "INDEXES", "INDICATOR", "INHERIT", "INHERITS",
		"INITIAL", "INITIALLY", "INLINE", "INNER", "INOUT", "INPUT", "INSENSITIVE", "INSERT",
		"INSTANCE", "INSTANTIABLE", "INSTEAD", "INT", "INTEGER", "INTEGRITY", "INTERSECT", "INTERSECTION",
		"INTERVAL", "INTO", "INVOKER", "IS", "ISNULL", "ISOLATION", "JOIN", "JSON",
		"JSON_ARRAY", "JSON_ARRAYAGG", "JSON_EXISTS", "JSON_OBJECT", "JSON_OBJECTAGG", "JSON_QUERY", "JSON_SCALAR", "JSON_SERIALIZE",
		"JSON_TABLE", "JSON_TABLE_PRIMITIVE", "JSON_VALUE", "K", "KEEP", "KEY", "KEYS", "KEY_MEMBER",
		"KEY_TYPE", "LABEL", "LAG", "LANGUAGE", "LARGE", "LAST", "LAST_VALUE", "LATERAL",
		"LEAD", "LEADING", "LEAKPROOF", "LEAST", "LEFT", "LENGTH", "LEVEL", "LIBRARY",
		"LIKE", "LIKE_REGEX", "LIMIT", "LINK", "LISTAGG", "LISTEN", "LN", "LOAD",
		"LOCAL", "LOCALTIME", "LOCALTIMESTAMP", "LOCATION", "LOCATOR", "LOCK", "LOCKED", "LOG",
		"LOG10", "LOGGED", "LOWER", "LPAD", "LTRIM", "M", "MAP", "MAPPING",
		"MATCH", "MATCHED", "MATCHES", "MATCH_NUMBER", "MATCH_RECOGNIZE", "MATERIALIZED", "MAX", "MAXVALUE",
		"MEASURES", "MEMBER", "MERGE", "MERGE_ACTION", "MESSAGE_LENGTH", "MESSAGE_OCTET_LENGTH", "MESSAGE_TEXT", "METHOD",
		"MIN", "MINUTE", "MINVALUE", "MOD", "MODE", "MODIFIES", "MODULE", "MONTH",
		"MORE", "MOVE", "MULTISET", "MUMPS", "NAME", "NAMES", "NAMESPACE", "NATIONAL",
		"NATURAL", "NCHAR", "NCLOB", "NESTED", "NESTING", "NEW", "NEXT", "NFC",
		"NFD", "NFKC", "NFKD", "NIL", "NO", "NONE", "NORMALIZE", "NORMALIZED",
		"NOT", "NOTHING", "NOTIFY", "NOTNULL", "NOWAIT", "NTH_VALUE", "NTILE", "NULL",
		"NULLABLE", "NULLIF", "NULLS", "NULL_ORDERING", "NUMBER", "NUMERIC", "OBJECT", "OCCURRENCE",
		"OCCURRENCES_REGEX", "OCTETS", "OCTET_LENGTH", "OF", "OFF", "OFFSET", "OIDS", "OLD",
		"OMIT", "ON", "ONE", "ONLY", "OPEN", "OPERATOR", "OPTION", "OPTIONS",
		"OR", "ORDER", "ORDERING", "ORDINALITY", "OTHERS", "OUT", "OUTER", "OUTPUT",
		"OVER", "OVERFLOW", "OVERLAPS", "OVERLAY", "OVERRIDING", "OWNED", "OWNER", "P",
		"PAD", "PARALLEL", "PARAMETER", "PARAMETER_MODE", "PARAMETER_NAME", "PARSER", "PARTIAL", "PARTITION",
		"PASCAL", "PASS", "PASSING", "PASSTHROUGH", "PASSWORD", "PAST", "PATH", "PATTERN",
		"PER", "PERCENT", "PERCENTILE_CONT", "PERCENTILE_DISC", "PERCENT_RANK", "PERIOD", "PERMISSION", "PERMUTE",
		"PIPE", "PLACING", "PLAN", "PLANS", "PLI", "POLICY", "PORTION", "POSITION",
		"POSITION_REGEX", "POWER", "PRECEDES", "PRECEDING", "PRECISION", "PREPARE", "PREPARED", "PRESERVE",
		"PREV", "PRIMARY", "PRIOR", "PRIVATE", "PRIVILEGES", "PROCEDURAL", "PROCEDURE", "PROCEDURES",
		"PROGRAM", "PRUNE", "PTF", "PUBLIC", "PUBLICATION", "QUOTE", "QUOTES", "RANGE",
		"RANK", "READ", "READS", "REAL", "REASSIGN", "RECHECK", "RECOVERY", "RECURSIVE",
		"REF", "REFERENCES", "REFERENCING", "REFRESH", "REGR_AVGX", "REGR_AVGY", "REGR_COUNT", "REGR_INTERCEPT",
		"REGR_R2", "REGR_SLOPE", "REGR_SXX", "REGR_SXY", "REGR_SYY", "REINDEX", "RELATIVE", "RELEASE",
		"RENAME", "REPEATABLE", "REPLACE", "REPLICA", "REQUIRING", "RESET", "RESPECT", "RESTART",
		"RESTORE", "RESTRICT", "RESULT", "RETURN", "RETURNED_CARDINALITY", "RETURNED_LENGTH", "RETURNED_SQLSTATE", "RETURNING",
		"RETURNS", "REVOKE", "RIGHT", "ROLE", "ROLLBACK", "ROLLUP", "ROUTINE", "ROUTINES",
		"ROUTINE_CATALOG", "ROUTINE_NAME", "ROUTINE_SCHEMA", "ROW", "ROWS", "ROW_COUNT", "ROW_NUMBER", "RPAD",
		"RTRIM", "RULE", "RUNNING", "SAVEPOINT", "SCALAR", "SCALE", "SCHEMA", "SCHEMAS",
		"SCHEMA_NAME", "SCOPE", "SCOPE_CATALOG", "SCOPE_NAME", "SCOPE_SCHEMA", "SCROLL", "SEARCH", "SECOND",
		"SECTION", "SECURITY", "SEEK", "SELECT", "SELECTIVE", "SELF", "SEMANTICS", "SENSITIVE",
		"SEQUENCE", "SEQUENCES", "SERIALIZABLE", "SERVER", "SERVER_NAME", "SESSION", "SESSION_USER", "SET",
		"SETOF", "SETS", "SHARE", "SHOW", "SIMILAR", "SIMPLE", "SIN", "SINH",
		"SIZE", "SKIP", "SMALLINT", "SNAPSHOT", "SOME", "SORT_DIRECTION", "SOURCE", "SPACE",
		"SPECIFIC", "SPECIFICTYPE", "SPECIFIC_NAME", "SQL", "SQLCODE", "SQLERROR", "SQLEXCEPTION", "SQLSTATE",
		"SQLWARNING", "SQRT", "STABLE", "STANDALONE", "START", "STATE", "STATEMENT", "STATIC",
		"STATISTICS", "STDDEV_POP", "STDDEV_SAMP", "STDIN", "STDOUT", "STORAGE", "STORED", "STRICT",
		"STRING", "STRIP", "STRUCTURE", "STYLE", "SUBCLASS_ORIGIN", "SUBMULTISET", "SUBSCRIPTION", "SUBSET",
		"SUBSTRING", "SUBSTRING_REGEX", "SUCCEEDS", "SUM", "SUPPORT", "SYMMETRIC", "SYSID", "SYSTEM",
		"SYSTEM_TIME", "SYSTEM_USER", "T", "TABLE", "TABLES", "TABLESAMPLE", "TABLESPACE", "TABLE_NAME",
		"TAN", "TANH", "TARGET", "TEMP", "TEMPLATE", "TEMPORARY", "TEXT", "THEN",
		"THROUGH", "TIES", "TIME", "TIMESTAMP", "TIMEZONE_HOUR", "TIMEZONE_MINUTE", "TO", "TOKEN",
		"TOP_LEVEL_COUNT", "TRAILING", "TRANSACTION", "TRANSACTION_ACTIVE", "TRANSFORM", "TRANSFORMS", "TRANSLATE", "TRANSLATE_REGEX",
		"TRANSLATION", "TREAT", "TRIGGER", "TRIGGER_CATALOG", "TRIGGER_NAME", "TRIGGER_SCHEMA", "TRIM", "TRIM_ARRAY",
		"TRUE", "TRUNCATE", "TRUSTED", "TYPE", "TYPES", "UESCAPE", "UNBOUNDED", "UNCOMMITTED",
		"UNCONDITIONAL", "UNDER", "UNENCRYPTED", "UNION", "UNIQUE", "UNKNOWN", "UNLINK", "UNLISTEN",
		"UNLOGGED", "UNMATCHED", "UNNAMED", "UNNEST", "UNTIL", "UNTYPED", "UPDATE", "UPPER",
		"URI", "USAGE", "USER", "USING", "UTF16", "UTF32", "UTF8", "VACUUM",
		"VALID", "VALIDATE", "VALIDATOR", "VALUE", "VALUES", "VALUE_OF", "VARBINARY", "VARCHAR",
		"VARIADIC", "VARYING", "VAR_POP", "VAR_SAMP", "VERBOSE", "VERSION", "VERSIONING", "VIEW",
		"VIEWS", "VOLATILE", "WHEN", "WHENEVER", "WHERE", "WHITESPACE", "WIDTH_BUCKET", "WINDOW",
		"WITH", "WITHIN", "WITHOUT", "WORK", "WRAPPER", "WRITE", "XML", "XMLAGG",
		"XMLATTRIBUTES", "XMLBINARY", "XMLCAST", "XMLCOMMENT", "XMLCONCAT", "XMLDECLARATION", "XMLDOCUMENT", "XMLELEMENT",
		"XMLEXISTS", "XMLFOREST", "XMLITERATE", "XMLNAMESPACES", "XMLPARSE", "XMLPI", "XMLQUERY", "XMLROOT",
		"XMLSCHEMA", "XMLSERIALIZE", "XMLTABLE", "XMLTEXT", "XMLVALIDATE", "YEAR", "YES", "ZONE",
	}

	predeclaredIdentifiers := []string{
		"INT", "TINYINT", "BIGINT", "FLOAT", "REAL",
		"DATE", "TIME", "DATETIME",
		"CHAR", "VARCHAR", "TEXT",
		"NCHAR", "NVARCHAR", "NTEXT",
		"BINARY", "VARBINARY",
		"CLOB", "BLOB", "XML", "JSON",
		"CURSOR", "TABLE",
	}

	return consumeSingleRuneLike(isLetter).
		ThenMaybe(consumeRunesLike(isLetterOrDigit)).
		MapWithInput(recognizeKeywordOrConsume(append(keywords, predeclaredIdentifiers...)))
}

func sqlOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		"+", "-", "*", "/", "%",
		"=", "!=",
		"<", "<=", ">", ">=",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}

func sqlNumberLiteralParseFunc() parser.Func {
	consumeDecimalZero := consumeString("0").
		ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool { return r == '0' }))

	consumeDecimalNonZero := consumeSingleRuneLike(func(r rune) bool {
		return r >= '1' && r <= '9'
	}).ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool {
		return r >= '0' && r <= '9'
	}))

	consumeDecimalLiteral := consumeDecimalZero.Or(consumeDecimalNonZero)

	consumeBinaryLiteral := (consumeString("0b").Or(consumeString("0B"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return r == '0' || r == '1'
		}))

	consumeHexLiteral := (consumeString("0x").Or(consumeString("0X"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		}))

	consumeIntLiteral := consumeBinaryLiteral.
		Or(consumeHexLiteral).
		Or(consumeDecimalLiteral)

	consumeDigitPart := consumeDigitsAndSeparators(false, func(r rune) bool {
		return r >= '0' && r <= '9'
	})
	consumePointFloat := (consumeDigitPart.Then(consumeString(".")).ThenMaybe(consumeDigitPart)).
		Or(consumeString(".").Then(consumeDigitPart))

	consumeExponentFloat := ((consumePointFloat).Or(consumeDigitPart)).
		Then((consumeString("e").Or(consumeString("E")))).
		ThenMaybe((consumeString("+").Or(consumeString("-")))).
		Then(consumeDigitPart)

	consumeFloatLiteral := consumeExponentFloat.Or(consumePointFloat)

	consumeImaginaryLiteral := (consumeFloatLiteral.Or(consumeDigitPart)).
		Then(consumeString("j").Or(consumeString("J")))

	return consumeImaginaryLiteral.
		Or(consumeFloatLiteral).
		Or(consumeIntLiteral).
		Map(recognizeToken(parser.TokenRoleNumber))
}
