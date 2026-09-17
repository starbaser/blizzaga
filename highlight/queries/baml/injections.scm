; Source: github.com/starbaser/tree-sitter-baml at cc27fb58a2a6687d5ee725c19996b4f67e037f08
; SPDX-License-Identifier: MIT

; Patterns are ordered from most to least specific: when several patterns claim
; the same content node, the earliest pattern's language wins.
;
; A bare raw-string tail is an FML template only when the containing function
; explicitly returns filament.Template; its prose is markup.
(
  (function_declaration
    return_type: (arrow_return_type
      (type_definition
        (custom_type
          (qualified_identifier) @_fml_return_type)))
    body: (function_body
      (expression_statement
        (raw_string_literal
          content: (raw_string_content) @injection.content)) .))
  (#eq? @_fml_return_type "filament.Template")
  (#set! injection.language "fml")
)

(
  (function_declaration
    return_type: (arrow_return_type
      (type_definition
        (custom_type
          (qualified_identifier) @_fml_return_type)))
    body: (function_body
      (return_statement
        value: (raw_string_literal
          content: (raw_string_content) @injection.content)) .))
  (#eq? @_fml_return_type "filament.Template")
  (#set! injection.language "fml")
)

; Every other raw string is a prompt-style template: prose stays string
; content, while Jinja markers, directives, expressions, and FML tags inside
; it are highlighted by the FML grammar.
(
  (raw_string_literal
    content: (raw_string_content) @injection.content)
  (#set! injection.language "baml-template")
)
