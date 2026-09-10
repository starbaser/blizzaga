; Source: github.com/starbaser/tree-sitter-baml at cc27fb58a2a6687d5ee725c19996b4f67e037f08
; SPDX-License-Identifier: MIT

; A bare raw-string tail is an FML template only when the containing function
; explicitly returns filament.Template. Ordinary BAML prompt/raw strings remain
; strings in the base language.
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
