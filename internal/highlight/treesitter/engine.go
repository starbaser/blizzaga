// Package treesitter turns Tree-sitter highlight-query captures into neutral
// byte spans. Renderer-specific token types belong in adapters outside this
// package.
package treesitter

import (
	"fmt"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Span is one capture produced by a Tree-sitter highlight query. Byte offsets
// refer to the original UTF-8 source passed to Engine.Highlight.
type Span struct {
	StartByte    uint
	EndByte      uint
	Capture      string
	PatternIndex uint
	Order        uint
	Priority     int
	PrioritySet  bool
}

type queryPriority struct {
	pattern  *int
	captures map[uint]int
}

// Engine owns an immutable grammar/query specification. Each Highlight call
// uses short-lived native parser, tree, query, and cursor values so all native
// allocations have an explicit lifetime.
type Engine struct {
	language   *sitter.Language
	query      string
	priorities map[uint]queryPriority
}

// NewEngine validates the grammar ABI, query syntax, predicates, and supported
// query settings before the engine is registered with a renderer.
func NewEngine(language *sitter.Language, querySource string) (*Engine, error) {
	if language == nil {
		return nil, fmt.Errorf("Tree-sitter language is nil")
	}
	if strings.TrimSpace(querySource) == "" {
		return nil, fmt.Errorf("Tree-sitter highlight query is empty")
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("set Tree-sitter language: %w", err)
	}

	query, queryErr := sitter.NewQuery(language, querySource)
	if queryErr != nil {
		return nil, fmt.Errorf("compile Tree-sitter highlight query: %w", queryErr)
	}
	defer query.Close()

	priorities, err := validateQuery(query)
	if err != nil {
		return nil, err
	}

	return &Engine{
		language:   language,
		query:      querySource,
		priorities: priorities,
	}, nil
}

// Highlight parses source and returns all valid captures in query-cursor order.
func (e *Engine) Highlight(source []byte) ([]Span, error) {
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(e.language); err != nil {
		return nil, fmt.Errorf("set Tree-sitter language: %w", err)
	}

	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil, fmt.Errorf("Tree-sitter parser returned no syntax tree")
	}
	defer tree.Close()

	query, queryErr := sitter.NewQuery(e.language, e.query)
	if queryErr != nil {
		return nil, fmt.Errorf("compile validated Tree-sitter highlight query: %w", queryErr)
	}
	defer query.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	captures := cursor.Captures(query, tree.RootNode(), source)
	captureNames := query.CaptureNames()

	spans := make([]Span, 0)
	for order := uint(0); ; order++ {
		match, captureIndex := captures.Next()
		if match == nil {
			break
		}
		if captureIndex >= uint(len(match.Captures)) {
			return nil, fmt.Errorf("Tree-sitter returned capture %d outside match bounds", captureIndex)
		}

		capture := match.Captures[captureIndex]
		if uint(capture.Index) >= uint(len(captureNames)) {
			return nil, fmt.Errorf("Tree-sitter returned unknown capture index %d", capture.Index)
		}

		start := capture.Node.StartByte()
		end := capture.Node.EndByte()
		if start >= end || end > uint(len(source)) {
			continue
		}

		span := Span{
			StartByte:    start,
			EndByte:      end,
			Capture:      captureNames[capture.Index],
			PatternIndex: match.PatternIndex,
			Order:        order,
		}
		if priority, ok := e.priority(match.PatternIndex, uint(capture.Index)); ok {
			span.Priority = priority
			span.PrioritySet = true
		}
		spans = append(spans, span)
	}

	return spans, nil
}

func (e *Engine) priority(patternIndex, captureIndex uint) (int, bool) {
	priority, ok := e.priorities[patternIndex]
	if !ok {
		return 0, false
	}
	if value, found := priority.captures[captureIndex]; found {
		return value, true
	}
	if priority.pattern != nil {
		return *priority.pattern, true
	}
	return 0, false
}

func validateQuery(query *sitter.Query) (map[uint]queryPriority, error) {
	priorities := make(map[uint]queryPriority)
	for pattern := uint(0); pattern < query.PatternCount(); pattern++ {
		if predicates := query.PropertyPredicates(pattern); len(predicates) > 0 {
			return nil, fmt.Errorf("Tree-sitter highlight pattern %d uses unsupported property predicates", pattern)
		}
		if predicates := query.GeneralPredicates(pattern); len(predicates) > 0 {
			return nil, fmt.Errorf(
				"Tree-sitter highlight pattern %d uses unsupported predicate %q",
				pattern,
				predicates[0].Operator,
			)
		}

		for _, setting := range query.PropertySettings(pattern) {
			if setting.Key != "priority" {
				return nil, fmt.Errorf(
					"Tree-sitter highlight pattern %d uses unsupported setting %q",
					pattern,
					setting.Key,
				)
			}
			if setting.Value == nil {
				return nil, fmt.Errorf("Tree-sitter highlight pattern %d has a priority without a value", pattern)
			}
			value, err := strconv.Atoi(*setting.Value)
			if err != nil {
				return nil, fmt.Errorf(
					"Tree-sitter highlight pattern %d has invalid priority %q: %w",
					pattern,
					*setting.Value,
					err,
				)
			}

			priority := priorities[pattern]
			if setting.CaptureId == nil {
				priority.pattern = &value
			} else {
				if priority.captures == nil {
					priority.captures = make(map[uint]int)
				}
				priority.captures[*setting.CaptureId] = value
			}
			priorities[pattern] = priority
		}
	}

	return priorities, nil
}
