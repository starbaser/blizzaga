// Package chromalexer adapts neutral Tree-sitter capture spans to Chroma v2's
// Lexer interface.
package chromalexer

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"

	tshighlight "github.com/starbaser/blizzaga/internal/highlight/treesitter"
)

var _ chroma.Lexer = (*Lexer)(nil)

// Lexer exposes a neutral Tree-sitter highlighting engine through Chroma v2.
type Lexer struct {
	config   chroma.Config
	engine   *tshighlight.Engine
	captures CaptureMap
	registry *chroma.LexerRegistry
	analyser func(string) float32
}

// New constructs a Chroma lexer around a validated Tree-sitter engine.
func New(config chroma.Config, engine *tshighlight.Engine, captures CaptureMap) (*Lexer, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("Chroma lexer name is empty")
	}
	if engine == nil {
		return nil, fmt.Errorf("Tree-sitter highlighting engine is nil")
	}
	if len(captures) == 0 {
		return nil, fmt.Errorf("Tree-sitter capture map is empty")
	}

	config.Aliases = append([]string(nil), config.Aliases...)
	config.Filenames = append([]string(nil), config.Filenames...)
	config.MimeTypes = append([]string(nil), config.MimeTypes...)
	captureCopy := make(CaptureMap, len(captures))
	for name, style := range captures {
		captureCopy[name] = style
	}

	return &Lexer{
		config:   config,
		engine:   engine,
		captures: captureCopy,
	}, nil
}

// Config returns the metadata Chroma uses for aliases, filenames, and MIME
// type selection.
func (l *Lexer) Config() *chroma.Config {
	return &l.config
}

// Tokenise parses source with Tree-sitter and converts captures into a complete,
// non-overlapping Chroma token stream.
func (l *Lexer) Tokenise(options *chroma.TokeniseOptions, text string) (chroma.Iterator, error) {
	if options == nil {
		options = &chroma.TokeniseOptions{State: "root", EnsureLF: true}
	}
	if options.State != "" && options.State != "root" {
		return nil, fmt.Errorf("Tree-sitter lexers do not support Chroma state %q", options.State)
	}
	if options.EnsureLF {
		text = ensureLF(text)
	}
	if !options.Nested && l.config.EnsureNL && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	spans, err := l.engine.Highlight([]byte(text))
	if err != nil {
		return nil, err
	}
	return chroma.Literator(renderTokens(text, spans, l.captures)...), nil
}

// SetRegistry associates this lexer with a Chroma registry.
func (l *Lexer) SetRegistry(registry *chroma.LexerRegistry) chroma.Lexer {
	l.registry = registry
	return l
}

// SetAnalyser preserves Chroma's content-based language detection contract.
func (l *Lexer) SetAnalyser(analyser func(string) float32) chroma.Lexer {
	l.analyser = analyser
	return l
}

// AnalyseText scores source using the analyser attached during registration.
func (l *Lexer) AnalyseText(text string) float32 {
	if l.analyser == nil {
		return 0
	}
	return l.analyser(text)
}

type styledSpan struct {
	span  tshighlight.Span
	style CaptureStyle
}

type spanEvent struct {
	offset uint
	span   int
	start  bool
}

type spanQueue struct {
	indices []int
	spans   []styledSpan
}

func (q spanQueue) Len() int {
	return len(q.indices)
}

func (q spanQueue) Less(i, j int) bool {
	return outranks(q.spans[q.indices[i]], q.spans[q.indices[j]])
}

func (q spanQueue) Swap(i, j int) {
	q.indices[i], q.indices[j] = q.indices[j], q.indices[i]
}

func (q *spanQueue) Push(value any) {
	q.indices = append(q.indices, value.(int))
}

func (q *spanQueue) Pop() any {
	last := len(q.indices) - 1
	value := q.indices[last]
	q.indices = q.indices[:last]
	return value
}

func renderTokens(source string, spans []tshighlight.Span, captures CaptureMap) []chroma.Token {
	styled := make([]styledSpan, 0, len(spans))
	events := make([]spanEvent, 0, len(spans)*2)
	for _, span := range spans {
		style, ok := captures.lookup(span.Capture)
		if !ok || span.StartByte >= span.EndByte || span.EndByte > uint(len(source)) {
			continue
		}
		index := len(styled)
		styled = append(styled, styledSpan{span: span, style: style})
		events = append(events,
			spanEvent{offset: span.StartByte, span: index, start: true},
			spanEvent{offset: span.EndByte, span: index},
		)
	}

	sort.Slice(events, func(i, j int) bool { return events[i].offset < events[j].offset })
	active := make([]bool, len(styled))
	queue := &spanQueue{spans: styled}
	heap.Init(queue)

	tokens := make([]chroma.Token, 0, len(events)+1)
	position := uint(0)
	eventIndex := 0
	for position < uint(len(source)) {
		for eventIndex < len(events) && events[eventIndex].offset == position {
			event := events[eventIndex]
			active[event.span] = event.start
			if event.start {
				heap.Push(queue, event.span)
			}
			eventIndex++
		}
		for queue.Len() > 0 && !active[queue.indices[0]] {
			heap.Pop(queue)
		}

		next := uint(len(source))
		if eventIndex < len(events) {
			next = events[eventIndex].offset
		}
		value := source[position:next]
		if queue.Len() == 0 {
			appendFallback(&tokens, value)
		} else {
			appendToken(&tokens, styled[queue.indices[0]].style.TokenType, value)
		}
		position = next
	}

	return tokens
}

func outranks(candidate, current styledSpan) bool {
	const defaultQueryPriority = 100

	candidatePriority := defaultQueryPriority
	if candidate.span.PrioritySet {
		candidatePriority = candidate.span.Priority
	}
	currentPriority := defaultQueryPriority
	if current.span.PrioritySet {
		currentPriority = current.span.Priority
	}
	if candidatePriority != currentPriority {
		return candidatePriority > currentPriority
	}

	candidateWidth := candidate.span.EndByte - candidate.span.StartByte
	currentWidth := current.span.EndByte - current.span.StartByte
	if candidateWidth != currentWidth {
		return candidateWidth < currentWidth
	}
	if candidate.style.Priority != current.style.Priority {
		return candidate.style.Priority > current.style.Priority
	}
	if candidate.span.PatternIndex != current.span.PatternIndex {
		return candidate.span.PatternIndex < current.span.PatternIndex
	}
	return candidate.span.Order > current.span.Order
}

func appendFallback(tokens *[]chroma.Token, value string) {
	for value != "" {
		whitespace := startsWithWhitespace(value)
		end := 0
		for end < len(value) {
			r, size := utf8.DecodeRuneInString(value[end:])
			if unicode.IsSpace(r) != whitespace {
				break
			}
			end += size
		}
		if end == 0 {
			end = 1
		}

		tokenType := chroma.Text
		if whitespace {
			tokenType = chroma.TextWhitespace
		}
		appendToken(tokens, tokenType, value[:end])
		value = value[end:]
	}
}

func startsWithWhitespace(value string) bool {
	r, _ := utf8.DecodeRuneInString(value)
	return unicode.IsSpace(r)
}

func appendToken(tokens *[]chroma.Token, tokenType chroma.TokenType, value string) {
	if value == "" {
		return
	}
	if len(*tokens) > 0 && (*tokens)[len(*tokens)-1].Type == tokenType {
		(*tokens)[len(*tokens)-1].Value += value
		return
	}
	*tokens = append(*tokens, chroma.Token{Type: tokenType, Value: value})
}

func ensureLF(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}
