; Source: github.com/starbaser/tree-sitter-baml at cc27fb58a2a6687d5ee725c19996b4f67e037f08
; SPDX-License-Identifier: MIT

; FML markup
(tag_name) @tag
(attribute_name) @attribute
(markup_comment) @comment.block
(text) @markup
(attribute_text_double) @string
(attribute_text_single) @string
(bare_attribute_value) @string

["<" ">" "/"] @punctuation.bracket
"=" @operator

; Template markers are match pairs around embedded code, the same role as the
; `${ }` interpolation markers of BAML backtick strings.
["{{" "}}" "{%" "%}"] @punctuation.special

; Template control flow
(if_directive "if" @keyword.control)
(elif_directive "elif" @keyword.control)
(else_directive) @keyword.control
(endif_directive) @keyword.control
(for_directive ["for" "in"] @keyword.control)
(endfor_directive) @keyword.control

(for_directive variable: (identifier) @variable)

; BAML expressions inside holes and directives
(primitive_type) @type.builtin
(unit_type) @type.builtin
(custom_type (identifier) @type)
(custom_type (qualified_identifier (identifier) @type))
(generic_type constructor: (identifier) @type)
(generic_type constructor: (qualified_identifier (identifier) @type))
(call_expression function: (identifier) @function.call)
(call_expression
  function: (member_expression member: (identifier) @function.method))
(member_expression member: (identifier) @variable.other.member)
(parameter name: (identifier) @variable.parameter)
(let_declaration name: (identifier) @variable)
(object_field name: (identifier) @property)
(map_entry key: (identifier) @property)
(argument name: (identifier) @variable.parameter)
(identifier) @variable
(string_literal) @string
(raw_string_literal) @string
(raw_string_delimiter) @punctuation.delimiter
(number_literal) @constant.numeric
(boolean_literal) @constant.builtin.boolean
(null_literal) @constant.builtin
(line_comment) @comment.line
(doc_comment) @comment.documentation
(block_comment) @comment.block

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

["(" ")" "{" "}" "[" "]"] @punctuation.bracket
["," ";" ":"] @punctuation.delimiter
"." @punctuation.delimiter
