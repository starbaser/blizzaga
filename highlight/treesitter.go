package highlight

import (
	"fmt"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type queryPriority struct {
	pattern  *int
	captures map[uint]int
}

type languageGrammar struct {
	language   *sitter.Language
	highlights string
	captures   CaptureMap
	priorities map[uint]queryPriority
}

type rawSpan struct {
	startByte   int
	endByte     int
	capture     string
	pattern     uint
	order       uint
	priority    int
	prioritySet bool
	depth       int
	mapping     CaptureMapping
}

func newLanguageGrammar(
	language *sitter.Language,
	highlights string,
	captures CaptureMap,
) (*languageGrammar, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("set Tree-sitter language: %w", err)
	}

	query, queryErr := sitter.NewQuery(language, highlights)
	if queryErr != nil {
		return nil, fmt.Errorf("compile Tree-sitter highlight query: %w", queryErr)
	}
	defer query.Close()
	priorities, err := validateHighlightQuery(query)
	if err != nil {
		return nil, err
	}
	return &languageGrammar{
		language:   language,
		highlights: highlights,
		captures:   captures,
		priorities: priorities,
	}, nil
}

func validateHighlightQuery(query *sitter.Query) (map[uint]queryPriority, error) {
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

func validateInjectionQuery(
	language *sitter.Language,
	querySource string,
	targets map[string]*languageSpec,
) error {
	query, queryErr := sitter.NewQuery(language, querySource)
	if queryErr != nil {
		return fmt.Errorf("compile Tree-sitter injection query: %w", queryErr)
	}
	defer query.Close()
	_, err := injectionTargets(query, targets)
	return err
}

func injectionTargets(
	query *sitter.Query,
	targets map[string]*languageSpec,
) (map[uint]*languageSpec, error) {
	captureNames := query.CaptureNames()
	contentCapture := -1
	for index, name := range captureNames {
		if name == "injection.content" {
			contentCapture = index
			break
		}
	}
	if contentCapture < 0 {
		return nil, fmt.Errorf("injection query has no @injection.content capture")
	}

	patterns := make(map[uint]*languageSpec)
	for pattern := uint(0); pattern < query.PatternCount(); pattern++ {
		if predicates := query.PropertyPredicates(pattern); len(predicates) > 0 {
			return nil, fmt.Errorf("Tree-sitter injection pattern %d uses unsupported property predicates", pattern)
		}
		if predicates := query.GeneralPredicates(pattern); len(predicates) > 0 {
			return nil, fmt.Errorf(
				"Tree-sitter injection pattern %d uses unsupported predicate %q",
				pattern,
				predicates[0].Operator,
			)
		}

		var target *languageSpec
		for _, setting := range query.PropertySettings(pattern) {
			if setting.Key != "injection.language" {
				return nil, fmt.Errorf(
					"Tree-sitter injection pattern %d uses unsupported setting %q",
					pattern,
					setting.Key,
				)
			}
			if setting.CaptureId != nil || setting.Value == nil {
				return nil, fmt.Errorf(
					"Tree-sitter injection pattern %d must set a literal injection.language",
					pattern,
				)
			}
			if target != nil {
				return nil, fmt.Errorf(
					"Tree-sitter injection pattern %d sets injection.language more than once",
					pattern,
				)
			}
			selector := normalizeIdentifier(*setting.Value)
			target = targets[selector]
			if target == nil {
				return nil, fmt.Errorf(
					"Tree-sitter injection pattern %d selects unregistered language %q",
					pattern,
					*setting.Value,
				)
			}
		}
		if target != nil {
			quantifiers := query.CaptureQuantifiers(pattern)
			if contentCapture >= len(quantifiers) || quantifiers[contentCapture] == sitter.CaptureQuantifierZero {
				return nil, fmt.Errorf(
					"Tree-sitter injection pattern %d sets injection.language without @injection.content",
					pattern,
				)
			}
			patterns[pattern] = target
		}
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("injection query has no injection.language settings")
	}
	return patterns, nil
}

type injectionRuntime struct {
	query          *sitter.Query
	cursor         *sitter.QueryCursor
	contentCapture uint
	targets        map[uint]*languageSpec
}

type treeState struct {
	spec        *languageSpec
	source      string
	parser      *sitter.Parser
	parse       func([]byte, *sitter.Tree) *sitter.Tree
	tree        *sitter.Tree
	query       *sitter.Query
	cursor      *sitter.QueryCursor
	injections  []injectionRuntime
	children    []*injectedState
	pendingEdit *Edit
}

type injectedState struct {
	start int
	end   int
	tree  *treeState
}

func newTreeState(spec *languageSpec, source string) (*treeState, error) {
	state := &treeState{spec: spec, source: source}
	state.parser = sitter.NewParser()
	if err := state.parser.SetLanguage(spec.grammar.language); err != nil {
		state.Close()
		return nil, fmt.Errorf("set %s Tree-sitter language: %w", spec.name, err)
	}
	state.parse = state.parser.Parse
	state.tree = state.parse([]byte(source), nil)
	if state.tree == nil {
		state.Close()
		return nil, fmt.Errorf("%s Tree-sitter parser returned no syntax tree", spec.name)
	}
	var queryErr *sitter.QueryError
	state.query, queryErr = sitter.NewQuery(spec.grammar.language, spec.grammar.highlights)
	if queryErr != nil {
		state.Close()
		return nil, fmt.Errorf("compile validated %s highlight query: %w", spec.name, queryErr)
	}
	state.cursor = sitter.NewQueryCursor()

	for _, injection := range spec.injections {
		query, queryErr := sitter.NewQuery(spec.grammar.language, injection.query)
		if queryErr != nil {
			state.Close()
			return nil, fmt.Errorf("compile validated %s injection query: %w", spec.name, queryErr)
		}
		targets, err := injectionTargets(query, injection.languages)
		if err != nil {
			query.Close()
			state.Close()
			return nil, fmt.Errorf("prepare validated %s injection query: %w", spec.name, err)
		}
		contentCapture := uint(0)
		for index, name := range query.CaptureNames() {
			if name == "injection.content" {
				contentCapture = uint(index)
				break
			}
		}
		state.injections = append(state.injections, injectionRuntime{
			query:          query,
			cursor:         sitter.NewQueryCursor(),
			contentCapture: contentCapture,
			targets:        targets,
		})
	}
	return state, nil
}

func (s *treeState) applyEdit(edit Edit, source string) error {
	inputEdit := sitter.InputEdit{
		StartByte:      uint(edit.StartByte),
		OldEndByte:     uint(edit.OldEndByte),
		NewEndByte:     uint(edit.StartByte + len(edit.NewText)),
		StartPosition:  pointAt(s.source, edit.StartByte),
		OldEndPosition: pointAt(s.source, edit.OldEndByte),
		NewEndPosition: pointAfter(pointAt(s.source, edit.StartByte), edit.NewText),
	}
	// Editing a clone preserves the current tree/source pair if the native
	// parser fails. The edited clone is only an incremental-parse input.
	editedTree := s.tree.Clone()
	editedTree.Edit(&inputEdit)
	newTree := s.parse([]byte(source), editedTree)
	editedTree.Close()
	if newTree == nil {
		return fmt.Errorf("%s Tree-sitter parser returned no syntax tree after edit", s.spec.name)
	}
	s.tree.Close()
	s.tree = newTree
	s.source = source
	pending := edit
	s.pendingEdit = &pending
	return nil
}

func (s *treeState) rawSpans(baseOffset, depth int, order *uint) ([]rawSpan, error) {
	spans := make([]rawSpan, 0)
	captures := s.cursor.Captures(s.query, s.tree.RootNode(), []byte(s.source))
	captureNames := s.query.CaptureNames()
	for {
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
		start := int(capture.Node.StartByte())
		end := int(capture.Node.EndByte())
		if start >= end || end > len(s.source) {
			continue
		}
		captureName := captureNames[capture.Index]
		mapping, ok := s.spec.grammar.captures.lookup(captureName)
		if !ok {
			continue
		}
		span := rawSpan{
			startByte: baseOffset + start,
			endByte:   baseOffset + end,
			capture:   captureName,
			pattern:   match.PatternIndex,
			order:     *order,
			depth:     depth,
			mapping:   mapping,
		}
		*order++
		if priority, ok := s.priority(match.PatternIndex, uint(capture.Index)); ok {
			span.priority = priority
			span.prioritySet = true
		}
		spans = append(spans, span)
	}

	injected, err := s.injectionRanges()
	if err != nil {
		return nil, err
	}
	if err := s.reconcileChildren(injected, s.pendingEdit); err != nil {
		return nil, err
	}
	s.pendingEdit = nil
	for _, child := range s.children {
		childSpans, err := child.tree.rawSpans(baseOffset+child.start, depth+1, order)
		if err != nil {
			return nil, err
		}
		spans = append(spans, childSpans...)
	}
	return spans, nil
}

func (s *treeState) priority(patternIndex, captureIndex uint) (int, bool) {
	priority, ok := s.spec.grammar.priorities[patternIndex]
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

type injectionRange struct {
	start  int
	end    int
	target *languageSpec
}

func (s *treeState) injectionRanges() ([]injectionRange, error) {
	var ranges []injectionRange
	for _, injection := range s.injections {
		matches := injection.cursor.Matches(injection.query, s.tree.RootNode(), []byte(s.source))
		for {
			match := matches.Next()
			if match == nil {
				break
			}
			target := injection.targets[match.PatternIndex]
			if target == nil {
				continue
			}
			for _, capture := range match.Captures {
				if uint(capture.Index) != injection.contentCapture {
					continue
				}
				start := int(capture.Node.StartByte())
				end := int(capture.Node.EndByte())
				if start < end && end <= len(s.source) {
					ranges = append(ranges, injectionRange{start: start, end: end, target: target})
				}
			}
		}
	}
	return ranges, nil
}

func (s *treeState) reconcileChildren(ranges []injectionRange, edit *Edit) error {
	used := make([]bool, len(s.children))
	children := make([]*injectedState, 0, len(ranges))
	for _, region := range ranges {
		source := s.source[region.start:region.end]
		var child *injectedState
		var childEdit *Edit
		for index, existing := range s.children {
			if used[index] || existing.tree.spec != region.target {
				continue
			}
			translated, ok := retainedChildEdit(existing, region, edit)
			if !ok {
				continue
			}
			if translated == nil {
				if existing.tree.source != source {
					continue
				}
			} else if applyStringEdit(existing.tree.source, *translated) != source {
				continue
			}
			used[index] = true
			child = existing
			childEdit = translated
			break
		}
		if child == nil {
			state, err := newTreeState(region.target, source)
			if err != nil {
				for _, allocated := range children {
					if !containsChild(s.children, allocated) {
						allocated.tree.Close()
					}
				}
				return err
			}
			child = &injectedState{tree: state}
		} else if childEdit != nil {
			if err := child.tree.applyEdit(*childEdit, source); err != nil {
				for _, allocated := range children {
					if !containsChild(s.children, allocated) {
						allocated.tree.Close()
					}
				}
				return err
			}
		}
		child.start = region.start
		child.end = region.end
		children = append(children, child)
	}
	for index, child := range s.children {
		if !used[index] {
			child.tree.Close()
		}
	}
	s.children = children
	return nil
}

// retainedChildEdit identifies the same injected region across one parent
// edit. Edits wholly inside its captured content are rebased into child byte
// coordinates. Edits outside only move or preserve the region. Any overlap
// with a capture boundary deliberately rejects reuse.
func retainedChildEdit(child *injectedState, region injectionRange, edit *Edit) (*Edit, bool) {
	if edit == nil {
		return nil, child.start == region.start && child.end == region.end
	}
	delta := len(edit.NewText) - (edit.OldEndByte - edit.StartByte)
	if edit.StartByte >= child.start && edit.OldEndByte <= child.end {
		if region.start != child.start || region.end != child.end+delta {
			return nil, false
		}
		translated := Edit{
			StartByte:  edit.StartByte - child.start,
			OldEndByte: edit.OldEndByte - child.start,
			NewText:    edit.NewText,
		}
		return &translated, true
	}
	if edit.OldEndByte <= child.start {
		return nil, region.start == child.start+delta && region.end == child.end+delta
	}
	if edit.StartByte >= child.end {
		return nil, region.start == child.start && region.end == child.end
	}
	return nil, false
}

func applyStringEdit(source string, edit Edit) string {
	var next strings.Builder
	next.Grow(len(source) - (edit.OldEndByte - edit.StartByte) + len(edit.NewText))
	next.WriteString(source[:edit.StartByte])
	next.WriteString(edit.NewText)
	next.WriteString(source[edit.OldEndByte:])
	return next.String()
}

func containsChild(children []*injectedState, target *injectedState) bool {
	for _, child := range children {
		if child == target {
			return true
		}
	}
	return false
}

// Close releases the parser, syntax tree, compiled queries, cursors, and all
// recursively injected trees owned by this state.
func (s *treeState) Close() {
	if s == nil {
		return
	}
	for _, child := range s.children {
		child.tree.Close()
	}
	s.children = nil
	for i := range s.injections {
		if s.injections[i].cursor != nil {
			s.injections[i].cursor.Close()
		}
		if s.injections[i].query != nil {
			s.injections[i].query.Close()
		}
	}
	s.injections = nil
	if s.cursor != nil {
		s.cursor.Close()
		s.cursor = nil
	}
	if s.query != nil {
		s.query.Close()
		s.query = nil
	}
	if s.tree != nil {
		s.tree.Close()
		s.tree = nil
	}
	if s.parser != nil {
		s.parser.Close()
		s.parser = nil
	}
}

func pointAt(source string, offset int) sitter.Point {
	point := sitter.Point{}
	for i := 0; i < offset; i++ {
		b := source[i]
		if b == '\n' {
			point.Row++
			point.Column = 0
		} else {
			point.Column++
		}
	}
	return point
}

func pointAfter(point sitter.Point, text string) sitter.Point {
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b == '\n' {
			point.Row++
			point.Column = 0
		} else {
			point.Column++
		}
	}
	return point
}
