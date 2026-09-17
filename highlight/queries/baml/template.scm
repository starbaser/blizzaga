; Source: github.com/starbaser/tree-sitter-baml at 17fd57427537e8fa7d6873007807da83a940604f
; SPDX-License-Identifier: MIT

; BAML raw strings (`#"..."#`) parsed with the FML grammar. This query is
; prepended to the FML highlight query: the prose between markers and tags is
; string content, matching backtick string content, while the FML query keeps
; supplying tags, attributes, template markers, directives, and expressions.
(text) @string
