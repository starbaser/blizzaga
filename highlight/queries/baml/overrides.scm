; Blizzaga's backtick delimiters have a distinct color from string content.
(backtick_string (string_delimiter) @string.delimiter)

; Annotation @ has higher priority than the grammar's generic @operator.
(smc_sequence "@" @keyword.operator)

; Chroma's NameVariableInstance inherits the neutral foreground. Preserve the
; grammar's scoped-variable meaning while selecting Blizzaga's blue name role.
(smc_splice path: (identifier) @variable.other.member)
(smc_splice path: (qualified_identifier (identifier) @variable.other.member))
