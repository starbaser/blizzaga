; Source: github.com/starbaser/tree-sitter-baml at bbf9ac51dede1bd17c05c83b721ea335223f479e
; SPDX-License-Identifier: MIT

; Declarations and control flow
[
  "enum"
  "class"
  "override"
  "function"
  "template_string"
  "type"
  "type_builder"
  "client"
  "generator"
  "retry_policy"
  "printer"
  "test"
] @keyword

[
  "if"
  "else"
  "for"
  "while"
  "in"
  "return"
  "match"
  "throws"
] @keyword.control

"let" @keyword.storage

; Types
(primitive_type) @type.builtin
(unit_type) @type.builtin
(custom_type (identifier) @type)
(custom_type (qualified_identifier (identifier) @type))
(generic_type constructor: (identifier) @type)
(generic_type constructor: (qualified_identifier (identifier) @type))
(enum_declaration name: (identifier) @type.enum)
(class_declaration name: (identifier) @type)
(type_alias name: (identifier) @type)

; Functions, methods, parameters, and variables
(function_declaration name: (identifier) @function)
(template_string_declaration name: (identifier) @function)
(call_expression function: (identifier) @function.call)
(call_expression
  function: (member_expression member: (identifier) @function.method))
(member_expression member: (identifier) @variable.other.member)
(parameter name: (identifier) @variable.parameter)
(lambda_expression parameters: (lambda_parameters (parameter name: (identifier) @variable.parameter)))
(let_declaration name: (identifier) @variable)
(for_in_loop variable: (identifier) @variable)
(object_field name: (identifier) @property)
(map_entry key: (identifier) @property)
(class_property name: (identifier) @property)
(argument name: (identifier) @variable.parameter)
(identifier) @variable

; Literals
(string_literal) @string
(raw_string_literal) @string
(raw_string_delimiter) @punctuation.delimiter

; Backtick strings: only the delimiters and literal text are strings, so the
; code inside `${...}` keeps its ordinary highlighting.
(backtick_string (string_delimiter) @string)
(backtick_string (string_content) @string)
(escape_sequence) @constant.character.escape
(interpolation_delimiter) @punctuation.special
[
  (template_else)
  (template_endif)
  (template_endfor)
] @keyword.control

(number_literal) @constant.numeric
(boolean_literal) @constant.builtin.boolean
(null_literal) @constant.builtin

; Comments
(line_comment) @comment.line
(doc_comment) @comment.block.documentation
(block_comment) @comment.block

; Operators
[
  "+"
  "-"
  "*"
  "/"
  "%"
  "=="
  "!="
  "<"
  ">"
  "<="
  ">="
  "&&"
  "||"
  "??"
  "!"
  "="
  "+="
  "-="
  "*="
  "/="
  "%="
  "?"
  "|"
  "->"
  "=>"
] @operator

; Attributes and special fields
[
  "@"
  "@@"
] @attribute
(block_attribute name: (identifier) @attribute)
(client_property "client" @keyword.special)
(prompt_property "prompt" @keyword.special)
(config_block type: _ @keyword.special)
(config_block name: (identifier) @constant)

; Punctuation
["(" ")" "{" "}" "[" "]" "<" ">"] @punctuation.bracket
["," ";" ":"] @punctuation.delimiter
"." @punctuation.delimiter
