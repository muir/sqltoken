#!/bin/bash
#
# Validate SQL literal patterns against MySQL, PostgreSQL, and SingleStore.
# Produces a markdown table showing which patterns each database accepts.
#
# Prerequisites:
#   mysql -u ubuntu                          (MySQL on default socket)
#   psql                                     (PostgreSQL on default socket)
#   mysql -u root -ptest -h 127.0.0.1 -P 3307  (SingleStore in Docker)

set -o pipefail

run_mysql()       { mysql -u ubuntu -N -e "$1" 2>/dev/null; }
run_postgres()    { psql -t -A -c "$1" 2>/dev/null; }
run_singlestore() { mysql -u root -ptest -h 127.0.0.1 -P 3307 -N -e "$1" 2>/dev/null; }

check() {
    local fn="$1"
    local sql="$2"
    if $fn "$sql" >/dev/null 2>&1; then
        echo -n "yes"
    else
        echo -n "no"
    fi
}

# Each row: description | SQL to execute
patterns=(
    "backslash-escaped single quote \\\\' |SELECT '\\''"
    "double-escaped single quote ''      |SELECT 'a''b'"
    "backslash-escaped double quote \\\\\"  |SELECT \"\\\"\""
    "double-escaped double quote \"\"     |SELECT \"a\"\"b\""
    "n'string'                            |SELECT n'test'"
    "N'string'                            |SELECT N'test'"
    "_utf8'string'                        |SELECT _utf8'test'"
    "_utf8mb4'string'                     |SELECT _utf8mb4'test'"
    "_latin1'string'                      |SELECT _latin1'test'"
    "_binary'string'                      |SELECT _binary'test'"
    "E'string'                            |SELECT E'test'"
    "E'backslash escape'                  |SELECT E'it\\'s'"
    "0x hex number                        |SELECT 0x1f"
    "X'hex'                               |SELECT X'1f'"
    "x'hex'                               |SELECT x'1f'"
    "0b binary number                     |SELECT 0b01"
    "B'binary'                            |SELECT B'01'"
    "b'binary'                            |SELECT b'01'"
    "# line comment                       |SELECT 1 # comment"
    "-- line comment                      |SELECT 1 -- comment"
    "/* block comment */                  |SELECT 1 /* comment */"
    "? placeholder                        |PREPARE s FROM 'SELECT ?'"
)

printf "| %-42s | %-7s | %-10s | %-12s |\n" "Pattern" "MySQL" "PostgreSQL" "SingleStore"
printf "|%-44s|%-9s|%-12s|%-14s|\n" "$(printf '%0.s-' {1..44})" "$(printf '%0.s-' {1..9})" "$(printf '%0.s-' {1..12})" "$(printf '%0.s-' {1..14})"

for entry in "${patterns[@]}"; do
    label="${entry%%|*}"
    sql="${entry##*|}"
    label="$(echo "$label" | sed 's/[[:space:]]*$//')"
    
    m=$(check run_mysql "$sql")
    p=$(check run_postgres "$sql")
    s=$(check run_singlestore "$sql")
    
    printf "| %-42s | %-7s | %-10s | %-12s |\n" "$label" "$m" "$p" "$s"
done
