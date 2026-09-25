package highlight

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestDefaultRegistryUsesTreeSitterBeforeChroma(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}

	source := "package main\r\nfunc render() { println(\"hello\") }"
	result, err := registry.Highlight(source, Query{Language: "golang"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Language != "Go" {
		t.Fatalf("language = %q, want Go", result.Language)
	}
	assertExactSource(t, result, source)
	assertClassifies(t, result, "render", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "println", chroma.NameBuiltin, "tree-sitter.function.builtin")
}

func TestChromaFallbackPreservesExactSource(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		language string
		source   string
	}{
		{language: "css", source: "a { color: red; }\r\n/* tail */"},
		{language: "json", source: "{\r\n  \"hello\": \"世界\"\r\n}"},
		{language: "haskell", source: "main = putStrLn \"hello\"\r\n"},
		{language: "not-a-language", source: "plain\r\ntext"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			t.Parallel()
			result, err := registry.Highlight(test.source, Query{Language: test.language})
			if err != nil {
				t.Fatal(err)
			}
			assertExactSource(t, result, test.source)
		})
	}
}

func TestFCSSRoutesToCSSLexer(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	const source = ".panel {\r\n  color: #ff5c57;\r\n}\r\n"
	tests := []struct {
		name  string
		query Query
	}{
		{name: "css-language", query: Query{Language: "css"}},
		{name: "language", query: Query{Language: "fcss"}},
		{name: "filename", query: Query{Filename: "theme.fcss"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := registry.Highlight(source, test.query)
			if err != nil {
				t.Fatal(err)
			}
			if result.Language != "CSS" {
				t.Fatalf("language = %q, want CSS", result.Language)
			}
			assertExactSource(t, result, source)
			assertClassifies(t, result, "panel", chroma.NameClass, "")
			assertClassifies(t, result, "color", chroma.Keyword, "")
			for _, span := range result.Spans {
				if span.CaptureClass != "" {
					t.Fatalf("FCSS used Tree-sitter capture %q", span.CaptureClass)
				}
			}
		})
	}

	unknown, err := registry.Highlight(source, Query{Language: "unknown-language", Filename: "theme.fcss"})
	if err != nil {
		t.Fatal(err)
	}
	if unknown.Language != fallbackSelection().name {
		t.Fatalf("explicit unknown language selected %q, want %q", unknown.Language, fallbackSelection().name)
	}
	assertExactSource(t, unknown, source)
}

func TestDefaultRegistryBundlesBAML(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "function Center(ctx: training.numeric.Context) -> training.numeric.Program {\r\n  ctx.finish(1024.0)\r\n}"
	result, err := registry.Highlight(source, Query{Filename: "program.baml"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Language != "BAML" {
		t.Fatalf("language = %q, want BAML", result.Language)
	}
	assertExactSource(t, result, source)
	assertClassifies(t, result, "Center", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "finish", chroma.NameFunction, "tree-sitter.function.method")
	assertClassifies(t, result, "1024.0", chroma.LiteralNumber, "tree-sitter.constant.numeric")
}

func TestBAMLSMCUsesBundledGrammarAndCaptures(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	const source = `@tensor.program(profile: "dense-counted-f32")
function Probe(@specialize config: Config, batch: Axis) -> tensor.Pair {
  let slots = tensor.rows {
    input features: F32[config.axis] = [1.0];
    support count: I64[] = 1;
  };
  let ordinary = batch * 2;
  let row = smc { $batch >> $config.row @ ($(choose(config)) * $other) };
  row
}`
	result, err := registry.Highlight(source, Query{Filename: "alloy-smc.baml"})
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
	assertClassifiesAt(t, result, strings.Index(source, "smc {"), "smc", chroma.LiteralNumber, "tree-sitter.constant.numeric")
	splice := strings.Index(source, "$batch")
	assertClassifiesAt(t, result, splice, "$", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, splice+1, "batch", chroma.NameVariable, "tree-sitter.variable")
	assertClassifiesAt(t, result, strings.Index(source, "$config.row"), "$", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, strings.Index(source, "$config.row")+len("$config."), "row", chroma.NameVariable, "tree-sitter.variable")
	assertClassifiesAt(t, result, strings.Index(source, " >> ")+1, ">>", chroma.Operator, "tree-sitter.operator")
	assertClassifiesAt(t, result, strings.Index(source, " @ ")+1, "@", chroma.Operator, "tree-sitter.keyword.operator")
	assertClassifiesAt(t, result, strings.Index(source, " * $other")+1, "*", chroma.Operator, "tree-sitter.operator")
	assertClassifiesAt(t, result, strings.Index(source, "batch * 2")+len("batch "), "*", chroma.Operator, "tree-sitter.operator")
	assertClassifiesAt(t, result, strings.Index(source, "tensor.rows")+len("tensor."), "rows", chroma.Keyword, "tree-sitter.keyword.special")
	assertClassifiesAt(t, result, strings.Index(source, "input features"), "input", chroma.Keyword, "tree-sitter.keyword.special")
	assertClassifiesAt(t, result, strings.Index(source, "F32["), "F32", chroma.KeywordType, "tree-sitter.type.builtin")
}

func TestBAMLBacktickStrings(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "function Ask(text: string) -> string {\n  client: Fast\n  prompt: `Summarize \\`this\\`: ${text} for $5\n${ctx.output_format()}`\n}"
	result, err := registry.Highlight(source, Query{Filename: "ask.baml"})
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
	assertClassifiesAt(t, result, strings.Index(source, "`Summarize"), "`", chroma.LiteralStringDelimiter, "tree-sitter.string.delimiter")
	assertClassifies(t, result, "Summarize ", chroma.LiteralString, "tree-sitter.string")
	assertClassifies(t, result, "\\`", chroma.LiteralStringEscape, "tree-sitter.constant.character.escape")
	assertClassifies(t, result, "${", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, 87, "text", chroma.NameVariable, "tree-sitter.variable")
	assertClassifiesAt(t, result, 91, "}", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifies(t, result, " for $5\n", chroma.LiteralString, "tree-sitter.string")
	assertClassifies(t, result, "output_format", chroma.NameFunction, "tree-sitter.function.method")
	assertClassifiesAt(t, result, strings.LastIndex(source, "`"), "`", chroma.LiteralStringDelimiter, "tree-sitter.string.delimiter")
}

func TestBAMLFourBacktickStrings(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "function Render(value: string) -> string {\n  let markdown = ````Heading\n```text\ninside\n```\n${value}````\n  markdown\n}\n"
	result, err := registry.Highlight(source, Query{Filename: "render.baml"})
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
	start := strings.Index(source, "````Heading")
	assertClassifiesAt(t, result, start, "````", chroma.LiteralStringDelimiter, "tree-sitter.string.delimiter")
	assertClassifiesAt(t, result, strings.Index(source, "```text"), "```", chroma.LiteralString, "tree-sitter.string")
	interpolation := strings.Index(source, "${value}")
	assertClassifiesAt(t, result, interpolation, "${", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, interpolation+2, "value", chroma.NameVariable, "tree-sitter.variable")
	assertClassifiesAt(t, result, interpolation+len("${value}"), "````", chroma.LiteralStringDelimiter, "tree-sitter.string.delimiter")
}

func TestBAMLRawStringsAreTemplates(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "function Ask(events: Event[]) -> string {\n  prompt #\"\nSummarize <instructions role=\"system\">these</instructions>:\n{% for event in events %}\n- {{ event.index }} \\`literal\\`\n{% endfor %}\n\"#\n}"
	result, err := registry.Highlight(source, Query{Filename: "ask.baml"})
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
	assertClassifies(t, result, "#\"", chroma.Punctuation, "tree-sitter.punctuation.delimiter")
	assertClassifies(t, result, "\"#", chroma.Punctuation, "tree-sitter.punctuation.delimiter")
	assertClassifies(t, result, "\nSummarize ", chroma.LiteralString, "tree-sitter.string")
	assertClassifiesAt(t, result, strings.Index(source, "<instructions")+1, "instructions", chroma.NameTag, "tree-sitter.tag")
	assertClassifies(t, result, "role", chroma.NameAttribute, "tree-sitter.attribute")
	assertClassifies(t, result, "these", chroma.LiteralString, "tree-sitter.string")
	assertClassifies(t, result, "{%", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifies(t, result, "for", chroma.Keyword, "tree-sitter.keyword.control")
	assertClassifiesAt(t, result, strings.Index(source, "{% for event")+7, "event", chroma.NameVariable, "tree-sitter.variable")
	assertClassifies(t, result, "{{", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, strings.Index(source, "event.index")+len("event."), "index", chroma.NameProperty, "tree-sitter.variable.other.member")
	assertClassifies(t, result, "}}", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifies(t, result, " \\`literal\\`\n", chroma.LiteralString, "tree-sitter.string")
	assertClassifies(t, result, "endfor", chroma.Keyword, "tree-sitter.keyword.control")
	assertClassifies(t, result, "%}", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
}

func TestDefaultRegistryScopesFMLInjection(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "class Props {\r\n  label string\r\n}\r\n" +
		"function Ordinary(name: string) -> string {\r\n  return ####\"ordinary {{ name }}\"####\r\n}\r\n" +
		"function View(props: Props) -> filament.Template {\r\n  return ###\"<region class=\"board\">\r\n    Plain Unicode ✦\r\n    <text content=\"{{ props.label }}\" />\r\n  </region>\"###\r\n}\r\n"
	result, err := registry.Highlight(source, Query{Language: "baml"})
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
	// The ordinary raw string is a prompt template: prose is string content,
	// the expression inside the marker is code.
	assertClassifies(t, result, "ordinary ", chroma.LiteralString, "tree-sitter.string")
	assertClassifiesAt(t, result, strings.Index(source, "{{ name }}"), "{{", chroma.LiteralStringInterpol, "tree-sitter.punctuation.special")
	assertClassifiesAt(t, result, strings.Index(source, "{{ name }}")+3, "name", chroma.NameVariable, "tree-sitter.variable")
	// The filament.Template tail is FML markup, not string content.
	assertClassifiesAt(t, result, strings.Index(source, "<region")+1, "region", chroma.NameTag, "tree-sitter.tag")
	assertClassifiesAt(t, result, strings.Index(source, "class=\"board\""), "class", chroma.NameAttribute, "tree-sitter.attribute")
	assertClassifiesAt(t, result, strings.Index(source, "Plain Unicode"), "Plain Unicode ✦", chroma.Text, "tree-sitter.markup")
	assertClassifiesAt(t, result, strings.Index(source, "props.label")+len("props."), "label", chroma.NameProperty, "tree-sitter.variable.other.member")
}

func TestFMLInjectionIncrementalMatchesFreshAndRetainsChild(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "class Props { label string }\n" +
		"function View(props: Props) -> filament.Template {\n" +
		"  return ##\"<text content=\"Unicode ✦ {{ props.label }}\" />\"##\n}\n"
	session, err := registry.NewSession(source, Query{Language: "baml"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	if len(session.tree.children) != 1 {
		t.Fatalf("injected children = %d, want 1", len(session.tree.children))
	}
	child := session.tree.children[0]
	parser := child.tree.parser
	nativeTree := child.tree.tree

	start := strings.Index(source, "props.label") + len("props.")
	incremental, err := session.ApplyEdit(Edit{StartByte: start, OldEndByte: start + len("label"), NewText: "title"})
	if err != nil {
		t.Fatal(err)
	}
	assertRetainedInjectedChild(t, session, child, parser, nativeTree, true)
	assertMatchesFresh(t, registry, incremental)
	assertExactSource(t, incremental, incremental.Source)

	childNativeTree := child.tree.tree
	start = strings.Index(incremental.Source, " }}\"")
	incomplete, err := session.ApplyEdit(Edit{StartByte: start, OldEndByte: start + len(" }}\""), NewText: ""})
	if err != nil {
		t.Fatal(err)
	}
	assertRetainedInjectedChild(t, session, child, parser, childNativeTree, true)
	assertMatchesFresh(t, registry, incomplete)
	assertExactSource(t, incomplete, incomplete.Source)
}

func TestSessionClosesRemovedAndRetainedInjectionResources(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	const source = "function View() -> filament.Template {\n  return ##\"<text />\"##\n}\n"
	session, err := registry.NewSession(source, Query{Language: "baml"})
	if err != nil {
		t.Fatal(err)
	}
	removed := session.tree.children[0].tree
	start := strings.Index(source, "filament.Template")
	result, err := session.ApplyEdit(Edit{
		StartByte:  start,
		OldEndByte: start + len("filament.Template"),
		NewText:    "string",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Without the filament.Template return type the raw string is an ordinary
	// prompt template: the FML child is replaced, not merely retained.
	if len(session.tree.children) != 1 {
		t.Fatalf("children after removing template return type = %d, want 1", len(session.tree.children))
	}
	if replacement := session.tree.children[0].tree; replacement == removed || replacement.spec.name != "BAML Template" {
		t.Fatalf("child after removing template return type = %q, want a fresh BAML Template child", replacement.spec.name)
	}
	assertTreeStateClosed(t, removed)
	assertExactSource(t, result, result.Source)
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	session, err = registry.NewSession(source, Query{Language: "baml"})
	if err != nil {
		t.Fatal(err)
	}
	parent := session.tree
	retained := parent.children[0].tree
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	assertTreeStateClosed(t, retained)
	assertTreeStateClosed(t, parent)
}

func TestIncrementalMatchesFreshAcrossEdits(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	session, err := registry.NewSession(
		"package main\n\nfunc greet() { println(\"héllo\") }\n",
		Query{Language: "go"},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	edits := []Edit{
		{StartByte: len("package main\n"), OldEndByte: len("package main\n"), NewText: "// 世界\n"},
		{StartByte: len("package main\n// 世界\n\nfunc "), OldEndByte: len("package main\n// 世界\n\nfunc greet"), NewText: "render"},
		{StartByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo\") "), OldEndByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo\") }"), NewText: ""},
		{StartByte: len("package main\n// 世界\n\nfunc render() { println(\""), OldEndByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo"), NewText: "hi 👩🏽‍💻"},
	}
	for index, edit := range edits {
		incremental, err := session.ApplyEdit(edit)
		if err != nil {
			t.Fatalf("edit %d: %v", index, err)
		}
		fresh, err := registry.Highlight(incremental.Source, Query{Language: "go"})
		if err != nil {
			t.Fatalf("fresh edit %d: %v", index, err)
		}
		if !reflect.DeepEqual(incremental, fresh) {
			t.Fatalf("edit %d incremental differs from fresh\nincremental: %#v\nfresh: %#v", index, incremental, fresh)
		}
		assertExactSource(t, incremental, incremental.Source)
	}
}

func TestSessionRejectsSplitUTF8EditAndCloses(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	session, err := registry.NewSession("package π\n", Query{Language: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.ApplyEdit(Edit{StartByte: 9, OldEndByte: 9, NewText: "x"}); err == nil {
		t.Fatal("expected split UTF-8 edit error")
	}
	first, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	first.Spans[0].TokenType = chroma.Error
	second, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	if second.Spans[0].TokenType == chroma.Error {
		t.Fatal("mutating a returned snapshot changed the retained session result")
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := session.Result(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Result error = %v, want ErrClosed", err)
	}
}

func TestFailedIncrementalParsePreservesCurrentSnapshot(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	const source = "package p\nvar answer = 1\n"
	session, err := registry.NewSession(source, Query{Language: "go"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	nativeTree := session.tree.tree
	parse := session.tree.parse
	session.tree.parse = func([]byte, *sitter.Tree) *sitter.Tree { return nil }

	start := strings.Index(source, "1")
	if _, err := session.ApplyEdit(Edit{StartByte: start, OldEndByte: start + 1, NewText: "2"}); err == nil {
		t.Fatal("expected incremental parse failure")
	}
	if session.tree.tree != nativeTree || session.tree.source != source || session.source != source {
		t.Fatal("failed parse mutated the current native tree/source pair")
	}
	session.tree.parse = parse
	result, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	assertExactSource(t, result, source)
}

func TestRegistryCopiesAndValidatesLanguageDeclarations(t *testing.T) {
	t.Parallel()
	aliases := []string{"test"}
	captures := CaptureMap{"function": {TokenType: chroma.NameFunction, Priority: 10}}
	language := Language{
		Name:       "Test",
		Aliases:    aliases,
		Filenames:  []string{"*.test"},
		Grammar:    goGrammar(),
		Highlights: `(function_declaration name: (identifier) @function)`,
		Captures:   captures,
	}
	registry, err := NewRegistry(language)
	if err != nil {
		t.Fatal(err)
	}
	aliases[0] = "mutated"
	captures["function"] = CaptureMapping{TokenType: chroma.NameVariable}

	result, err := registry.Highlight("package p\nfunc chosen() {}", Query{Language: "test"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "chosen", chroma.NameFunction, "tree-sitter.function")
	infos := registry.Languages()
	infos[0].Aliases[0] = "also-mutated"
	if result, err := registry.Highlight("package p\nfunc chosen() {}", Query{Language: "test"}); err != nil || result.Language != "Test" {
		t.Fatalf("frozen alias lookup = (%q, %v)", result.Language, err)
	}

	invalid := []Language{
		{Name: "", Grammar: goGrammar(), Highlights: `(identifier) @variable`, Captures: StandardCaptures()},
		{Name: "Bad", Grammar: goGrammar(), Highlights: `(not_a_go_node) @variable`, Captures: StandardCaptures()},
		{Name: "Bad", Grammar: goGrammar(), Highlights: `(identifier) @variable`, Captures: nil},
	}
	for index, declaration := range invalid {
		if _, err := NewRegistry(declaration); err == nil {
			t.Errorf("invalid declaration %d succeeded", index)
		}
	}
	if _, err := NewRegistry(language, language); err == nil {
		t.Error("duplicate identifiers succeeded")
	}
}

func TestTextPredicatesAndOverlapPrecedence(t *testing.T) {
	t.Parallel()
	registry, err := NewRegistry(Language{
		Name:       "PredicateGo",
		Aliases:    []string{"predicate-go"},
		Grammar:    goGrammar(),
		Highlights: "((identifier) @variable (#match? @variable \"^pick\"))\n((identifier) @function (#eq? @function \"picked\"))",
		Captures:   StandardCaptures(),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := registry.Highlight("package p\nvar picked = skipped", Query{Language: "predicate-go"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "picked", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "skipped", chroma.Text, "")
}

func TestQueryPriorityOverridesCaptureMapping(t *testing.T) {
	t.Parallel()
	registry, err := NewRegistry(Language{
		Name:    "PriorityGo",
		Grammar: goGrammar(),
		Highlights: `
((identifier) @function)
((identifier) @variable
  (#set! priority "200"))
`,
		Captures: StandardCaptures(),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := registry.Highlight("package p\nvar chosen = 1", Query{Language: "PriorityGo"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "chosen", chroma.NameVariable, "tree-sitter.variable")
}

func TestRegisteredInjectionRebasesSpansAndSurvivesEdits(t *testing.T) {
	t.Parallel()
	inner := Language{
		Name:       "Inner",
		Aliases:    []string{"inner"},
		Grammar:    goGrammar(),
		Highlights: `(raw_string_literal) @function`,
		Captures:   StandardCaptures(),
	}
	outer := Language{
		Name:       "Outer",
		Aliases:    []string{"outer"},
		Grammar:    goGrammar(),
		Highlights: `(raw_string_literal) @string`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `((raw_string_literal) @injection.content (#set! injection.language "fml"))`,
			Languages: map[string]string{"fml": "inner"},
		}},
	}
	registry, err := NewRegistry(inner, outer)
	if err != nil {
		t.Fatal(err)
	}
	source := "package p\nvar template = `hello`\n"
	session, err := registry.NewSession(source, Query{Language: "outer"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	initial, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, initial, "`hello`", chroma.NameFunction, "tree-sitter.function")
	child := session.tree.children[0]
	parser := child.tree.parser
	nativeTree := child.tree.tree

	insert := strings.Index(source, "var template")
	incremental, err := session.ApplyEdit(Edit{StartByte: insert, OldEndByte: insert, NewText: "// moved\n"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, incremental, "`hello`", chroma.NameFunction, "tree-sitter.function")
	assertRetainedInjectedChild(t, session, child, parser, nativeTree, false)
	fresh, err := registry.Highlight(incremental.Source, Query{Language: "outer"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(incremental, fresh) {
		t.Fatalf("injected incremental differs from fresh\nincremental: %#v\nfresh: %#v", incremental, fresh)
	}

	nativeTree = child.tree.tree
	start := strings.Index(incremental.Source, "hello")
	incremental, err = session.ApplyEdit(Edit{StartByte: start, OldEndByte: start + len("hello"), NewText: "héllo"})
	if err != nil {
		t.Fatal(err)
	}
	assertRetainedInjectedChild(t, session, child, parser, nativeTree, true)
	assertMatchesFresh(t, registry, incremental)
}

func TestInjectionValidation(t *testing.T) {
	t.Parallel()
	base := Language{
		Name:       "Base",
		Grammar:    goGrammar(),
		Highlights: `(identifier) @variable`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `(raw_string_literal) @other`,
			Languages: map[string]string{"go": "missing"},
		}},
	}
	if _, err := NewRegistry(base); err == nil {
		t.Fatal("invalid injection succeeded")
	}
	cyclic := Language{
		Name:       "Cyclic",
		Grammar:    goGrammar(),
		Highlights: `(identifier) @variable`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `((raw_string_literal) @injection.content (#set! injection.language "self"))`,
			Languages: map[string]string{"self": "cyclic"},
		}},
	}
	if _, err := NewRegistry(cyclic); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cyclic injection error = %v", err)
	}
}

func TestChromaSourceContractError(t *testing.T) {
	t.Parallel()
	lexer := &changingLexer{config: chroma.Config{Name: "changing"}}
	_, err := highlightChroma("a\r\nb", selection{name: "changing", lexer: lexer})
	var contractErr *SourceContractError
	if !errors.As(err, &contractErr) {
		t.Fatalf("error = %v, want SourceContractError", err)
	}
	if contractErr.Offset != 1 || contractErr.SourceSize != 4 || contractErr.TokenSize != 3 {
		t.Fatalf("contract error = %#v", contractErr)
	}
}

func TestSrceryStyleIsPublic(t *testing.T) {
	t.Parallel()
	style := SrceryStyle()
	if style == nil || style.Name != "srcery" {
		t.Fatalf("SrceryStyle = %#v", style)
	}
	if style.Get(chroma.Keyword).Colour.IsSet() == false {
		t.Fatal("Srcery keyword colour is unset")
	}
}

func TestChromaClass(t *testing.T) {
	t.Parallel()
	if got := ChromaClass(chroma.KeywordDeclaration); got != "chroma.keyword.declaration" {
		t.Fatalf("class = %q", got)
	}
}

func goGrammar() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_go.Language())
}

func assertExactSource(t *testing.T, result Result, want string) {
	t.Helper()
	if result.Source != want {
		t.Fatalf("source = %q, want %q", result.Source, want)
	}
	var joined strings.Builder
	for _, token := range result.Tokens() {
		joined.WriteString(token.Value)
	}
	if joined.String() != want {
		t.Fatalf("token concatenation = %q, want %q", joined.String(), want)
	}
	position := 0
	for index, span := range result.Spans {
		if span.StartByte != position || span.EndByte <= span.StartByte {
			t.Fatalf("span %d = %#v after byte %d", index, span, position)
		}
		position = span.EndByte
	}
	if position != len(want) {
		t.Fatalf("spans end at byte %d, want %d", position, len(want))
	}
}

func assertClassifies(t *testing.T, result Result, text string, tokenType chroma.TokenType, captureClass string) {
	t.Helper()
	start := strings.Index(result.Source, text)
	if start < 0 {
		t.Fatalf("source does not contain %q", text)
	}
	end := start + len(text)
	for _, span := range result.Spans {
		if span.StartByte <= start && span.EndByte >= end {
			if span.TokenType != tokenType || span.CaptureClass != captureClass {
				t.Fatalf("%q span = %#v, want token %s capture %q", text, span, tokenType, captureClass)
			}
			return
		}
	}
	t.Fatalf("no single span classifies %q in %#v", text, result.Spans)
}

func assertClassifiesAt(t *testing.T, result Result, start int, text string, tokenType chroma.TokenType, captureClass string) {
	t.Helper()
	end := start + len(text)
	if start < 0 || end > len(result.Source) {
		t.Fatalf("source range [%d:%d] is outside %d bytes", start, end, len(result.Source))
	}
	if result.Source[start:end] != text {
		t.Fatalf("source[%d:%d] = %q, want %q", start, end, result.Source[start:end], text)
	}
	for _, span := range result.Spans {
		if span.StartByte <= start && span.EndByte >= end {
			if span.TokenType != tokenType || span.CaptureClass != captureClass {
				t.Fatalf("%q span = %#v, want token %s capture %q", text, span, tokenType, captureClass)
			}
			return
		}
	}
	t.Fatalf("no single span classifies %q at byte %d in %#v", text, start, result.Spans)
}

func assertMatchesFresh(t *testing.T, registry *Registry, incremental Result) {
	t.Helper()
	fresh, err := registry.Highlight(incremental.Source, Query{Language: incremental.Language})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(incremental, fresh) {
		t.Fatalf("incremental differs from fresh\nincremental: %#v\nfresh: %#v", incremental, fresh)
	}
}

func assertRetainedInjectedChild(
	t *testing.T,
	session *Session,
	child *injectedState,
	parser *sitter.Parser,
	previousTree *sitter.Tree,
	wantReparse bool,
) {
	t.Helper()
	if len(session.tree.children) != 1 || session.tree.children[0] != child || child.tree.parser != parser {
		t.Fatal("injected child/parser identity was not retained")
	}
	if gotReparse := child.tree.tree != previousTree; gotReparse != wantReparse {
		t.Fatalf("injected native tree changed = %v, want %v", gotReparse, wantReparse)
	}
}

func assertTreeStateClosed(t *testing.T, state *treeState) {
	t.Helper()
	if state.parser != nil || state.tree != nil || state.query != nil || state.cursor != nil || len(state.injections) != 0 || len(state.children) != 0 {
		t.Fatal("Tree-sitter state retained native resources after Close")
	}
}

type changingLexer struct {
	config chroma.Config
}

func (l *changingLexer) Config() *chroma.Config { return &l.config }
func (l *changingLexer) Tokenise(_ *chroma.TokeniseOptions, text string) (chroma.Iterator, error) {
	return chroma.Literator(chroma.Token{Type: chroma.Text, Value: strings.ReplaceAll(text, "\r\n", "\n")}), nil
}
func (l *changingLexer) SetRegistry(_ *chroma.LexerRegistry) chroma.Lexer { return l }
func (l *changingLexer) SetAnalyser(_ func(string) float32) chroma.Lexer  { return l }
func (l *changingLexer) AnalyseText(string) float32                       { return 0 }
