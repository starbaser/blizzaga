; Blizzaga's backtick delimiters have a distinct color from string content.
(backtick_string (string_delimiter) @string.delimiter)

; Annotation @ has higher priority than the grammar's generic @operator.
(smc_sequence "@" @keyword.operator)
