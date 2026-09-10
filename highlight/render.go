package highlight

import (
	"container/heap"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
)

type spanEvent struct {
	offset int
	span   int
	start  bool
}

type spanQueue struct {
	indices []int
	spans   []rawSpan
}

func (q spanQueue) Len() int           { return len(q.indices) }
func (q spanQueue) Less(i, j int) bool { return outranks(q.spans[q.indices[i]], q.spans[q.indices[j]]) }
func (q spanQueue) Swap(i, j int)      { q.indices[i], q.indices[j] = q.indices[j], q.indices[i] }
func (q *spanQueue) Push(value any)    { q.indices = append(q.indices, value.(int)) }
func (q *spanQueue) Pop() any {
	last := len(q.indices) - 1
	value := q.indices[last]
	q.indices = q.indices[:last]
	return value
}

func renderTreeSpans(source string, raw []rawSpan) []Span {
	valid := make([]rawSpan, 0, len(raw))
	events := make([]spanEvent, 0, len(raw)*2)
	for _, span := range raw {
		if span.startByte < 0 || span.startByte >= span.endByte || span.endByte > len(source) {
			continue
		}
		index := len(valid)
		valid = append(valid, span)
		events = append(events,
			spanEvent{offset: span.startByte, span: index, start: true},
			spanEvent{offset: span.endByte, span: index},
		)
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].offset < events[j].offset })

	active := make([]bool, len(valid))
	queue := &spanQueue{spans: valid}
	heap.Init(queue)
	spans := make([]Span, 0, len(events)+1)
	position := 0
	eventIndex := 0
	for position < len(source) {
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

		next := len(source)
		if eventIndex < len(events) {
			next = events[eventIndex].offset
		}
		if queue.Len() == 0 {
			appendFallbackSpans(&spans, source, position, next)
		} else {
			winner := valid[queue.indices[0]]
			appendSpan(&spans, Span{
				StartByte:    position,
				EndByte:      next,
				TokenType:    winner.mapping.TokenType,
				ChromaClass:  ChromaClass(winner.mapping.TokenType),
				CaptureClass: captureClass(winner.capture),
			})
		}
		position = next
	}
	return spans
}

func outranks(candidate, current rawSpan) bool {
	const defaultQueryPriority = 100
	candidatePriority := defaultQueryPriority
	if candidate.prioritySet {
		candidatePriority = candidate.priority
	}
	currentPriority := defaultQueryPriority
	if current.prioritySet {
		currentPriority = current.priority
	}
	if candidatePriority != currentPriority {
		return candidatePriority > currentPriority
	}
	candidateWidth := candidate.endByte - candidate.startByte
	currentWidth := current.endByte - current.startByte
	if candidateWidth != currentWidth {
		return candidateWidth < currentWidth
	}
	if candidate.mapping.Priority != current.mapping.Priority {
		return candidate.mapping.Priority > current.mapping.Priority
	}
	if candidate.depth != current.depth {
		return candidate.depth > current.depth
	}
	if candidate.pattern != current.pattern {
		return candidate.pattern < current.pattern
	}
	return candidate.order > current.order
}

func appendFallbackSpans(spans *[]Span, source string, start, end int) {
	for start < end {
		r, size := utf8.DecodeRuneInString(source[start:end])
		whitespace := unicode.IsSpace(r)
		next := start + size
		for next < end {
			r, size = utf8.DecodeRuneInString(source[next:end])
			if unicode.IsSpace(r) != whitespace {
				break
			}
			next += size
		}
		tokenType := chroma.Text
		if whitespace {
			tokenType = chroma.TextWhitespace
		}
		appendSpan(spans, Span{
			StartByte:   start,
			EndByte:     next,
			TokenType:   tokenType,
			ChromaClass: ChromaClass(tokenType),
		})
		start = next
	}
}

func appendSpan(spans *[]Span, span Span) {
	if span.StartByte >= span.EndByte {
		return
	}
	if len(*spans) > 0 {
		last := &(*spans)[len(*spans)-1]
		if last.EndByte == span.StartByte &&
			last.TokenType == span.TokenType &&
			last.ChromaClass == span.ChromaClass &&
			last.CaptureClass == span.CaptureClass {

			last.EndByte = span.EndByte
			return
		}
	}
	*spans = append(*spans, span)
}
