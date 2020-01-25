package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chris-ramon/douceur/css"
)

func MustParse(t *testing.T, txt string, nbRules int) *css.Stylesheet {
	stylesheet, err := Parse(txt)
	if err != nil {
		t.Fatal("Failed to parse CSS", err, txt)
	}

	if len(stylesheet.Rules) != nbRules {
		t.Fatal(fmt.Sprintf("Failed to parse CSS \"%s\", expected %d rules but got %d", txt, nbRules, len(stylesheet.Rules)))
	}

	return stylesheet
}

func MustEqualRule(t *testing.T, parsedRule *css.Rule, expectedRule *css.Rule) {
	if !parsedRule.Equal(expectedRule) {
		diff := parsedRule.Diff(expectedRule)

		t.Fatal(fmt.Sprintf("Rule parsing error\nExpected:\n\"%s\"\nGot:\n\"%s\"\nDiff:\n%s", expectedRule, parsedRule, strings.Join(diff, "\n")))
	}
}

func MustEqualCSS(t *testing.T, ruleString string, expected string) {
	if ruleString != expected {
		t.Fatal(fmt.Sprintf("CSS generation error\n   Expected:\n\"%s\"\n   Got:\n\"%s\"", expected, ruleString))
	}
}

func TestQualifiedRule(t *testing.T) {
	input := `/* This is a comment */
p > a {
    color: blue;
    text-decoration: underline; /* This is a comment */
}`

	expectedRule := &css.Rule{
		Kind:    css.QualifiedRule,
		Prelude: "p > a",
		Selectors: []*css.Selector{
			{
				Value:  "p > a",
				Line:   2,
				Column: 1,
			},
		},
		Declarations: []*css.Declaration{
			{
				Property: "color",
				Value:    "blue",
				Line:     3,
				Column:   5,
			},
			{
				Property: "text-decoration",
				Value:    "underline",
				Line:     4,
				Column:   5,
			},
		},
	}

	expectedOutput := `p > a {
  color: blue;
  text-decoration: underline;
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestQualifiedRuleImportant(t *testing.T) {
	input := `/* This is a comment */
p > a {
    color: blue;
    text-decoration: underline !important;
    font-weight: normal   !IMPORTANT    ;
}`

	expectedRule := &css.Rule{
		Kind:    css.QualifiedRule,
		Prelude: "p > a",
		Selectors: []*css.Selector{
			{
				Value:  "p > a",
				Line:   2,
				Column: 1,
			},
		},
		Declarations: []*css.Declaration{
			{
				Property:  "color",
				Value:     "blue",
				Important: false,
				Line:      3,
				Column:    5,
			},
			{
				Property:  "text-decoration",
				Value:     "underline",
				Important: true,
				Line:      4,
				Column:    5,
			},
			{
				Property:  "font-weight",
				Value:     "normal",
				Important: true,
				Line:      5,
				Column:    5,
			},
		},
	}

	expectedOutput := `p > a {
  color: blue;
  text-decoration: underline !important;
  font-weight: normal !important;
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestQualifiedRuleSelectors(t *testing.T) {
	input := `table, tr, td {
  padding: 0;
}

body,
  h1,   h2,
    h3   {
  color: #fff;
}`

	expectedRule1 := &css.Rule{
		Kind:    css.QualifiedRule,
		Prelude: "table, tr, td",
		Selectors: []*css.Selector{
			{
				Value:  "table",
				Line:   1,
				Column: 1,
			},
			{
				Value:  "tr",
				Line:   1,
				Column: 8,
			},
			{
				Value:  "td",
				Line:   1,
				Column: 12,
			},
		},
		Declarations: []*css.Declaration{
			{
				Property: "padding",
				Value:    "0",
				Line:     2,
				Column:   3,
			},
		},
	}

	expectedRule2 := &css.Rule{
		Kind: css.QualifiedRule,
		Prelude: `body,
  h1,   h2,
    h3`,
		Selectors: []*css.Selector{
			{
				Value:  "body",
				Line:   5,
				Column: 1,
			},
			{
				Value:  "h1",
				Line:   6,
				Column: 3,
			},
			{
				Value:  "h2",
				Line:   6,
				Column: 9,
			},
			{
				Value:  "h3",
				Line:   7,
				Column: 5,
			},
		},

		Declarations: []*css.Declaration{
			{
				Property: "color",
				Value:    "#fff",
				Line:     8,
				Column:   3,
			},
		},
	}

	expectedOutput := `table, tr, td {
  padding: 0;
}
body, h1, h2, h3 {
  color: #fff;
}`

	stylesheet := MustParse(t, input, 2)

	MustEqualRule(t, stylesheet.Rules[0], expectedRule1)
	MustEqualRule(t, stylesheet.Rules[1], expectedRule2)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleCharset(t *testing.T) {
	input := `@charset "UTF-8";`

	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@charset",
		Prelude: "\"UTF-8\"",
	}

	expectedOutput := `@charset "UTF-8";`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleCounterStyle(t *testing.T) {
	input := `@counter-style footnote {
  system: symbolic;
  symbols: '*' ⁑ † ‡;
  suffix: '';
}`

	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@counter-style",
		Prelude: "footnote",
		Declarations: []*css.Declaration{
			{
				Property: "system",
				Value:    "symbolic",
				Line:     2,
				Column:   3,
			},
			{
				Property: "symbols",
				Value:    "'*' ⁑ † ‡",
				Line:     3,
				Column:   3,
			},
			{
				Property: "suffix",
				Value:    "''",
				Line:     4,
				Column:   3,
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleDocument(t *testing.T) {
	input := `@document url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*")
{
  /* CSS rules here apply to:
     + The page "http://www.w3.org/".
     + Any page whose URL begins with "http://www.w3.org/Style/"
     + Any page whose URL's host is "mozilla.org" or ends with
       ".mozilla.org"
     + Any page whose URL starts with "https:" */

  /* make the above-mentioned pages really ugly */
  body { color: purple; background: yellow; }
}`

	expectedRule := &css.Rule{
		Kind: css.AtRule,
		Name: "@document",
		Prelude: `url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*")`,
		Rules: []*css.Rule{
			{
				Kind:    css.QualifiedRule,
				Prelude: "body",
				Selectors: []*css.Selector{
					{
						Value:  "body",
						Line:   14,
						Column: 3,
					},
				},
				Declarations: []*css.Declaration{
					{
						Property: "color",
						Value:    "purple",
						Line:     14,
						Column:   10,
					},
					{
						Property: "background",
						Value:    "yellow",
						Line:     14,
						Column:   25,
					},
				},
			},
		},
	}

	expectCSS := `@document url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*") {
  body {
    color: purple;
    background: yellow;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectCSS)
}

func TestAtRuleFontFace(t *testing.T) {
	input := `@font-face {
  font-family: MyHelvetica;
  src: local("Helvetica Neue Bold"),
       local("HelveticaNeue-Bold"),
       url(MgOpenModernaBold.ttf);
  font-weight: bold;
}`

	expectedRule := &css.Rule{
		Kind: css.AtRule,
		Name: "@font-face",
		Declarations: []*css.Declaration{
			{
				Property: "font-family",
				Value:    "MyHelvetica",
				Line:     2,
				Column:   3,
			},
			{
				Property: "src",
				Value: `local("Helvetica Neue Bold"),
       local("HelveticaNeue-Bold"),
       url(MgOpenModernaBold.ttf)`,
				Line:   3,
				Column: 3,
			},
			{
				Property: "font-weight",
				Value:    "bold",
				Line:     6,
				Column:   3,
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleFontFeatureValues(t *testing.T) {
	input := `@font-feature-values Font Two { /* How to activate nice-style in Font Two */
  @styleset {
    nice-style: 4;
  }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@font-feature-values",
		Prelude: "Font Two",
		Rules: []*css.Rule{
			{
				Kind: css.AtRule,
				Name: "@styleset",
				Declarations: []*css.Declaration{
					{
						Property: "nice-style",
						Value:    "4",
						Line:     3,
						Column:   5,
					},
				},
			},
		},
	}

	expectedOutput := `@font-feature-values Font Two {
  @styleset {
    nice-style: 4;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleImport(t *testing.T) {
	input := `@import "my-styles.css";
@import url('landscape.css') screen and (orientation:landscape);`

	expectedRule1 := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@import",
		Prelude: "\"my-styles.css\"",
	}

	expectedRule2 := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@import",
		Prelude: "url('landscape.css') screen and (orientation:landscape)",
	}

	stylesheet := MustParse(t, input, 2)

	MustEqualRule(t, stylesheet.Rules[0], expectedRule1)
	MustEqualRule(t, stylesheet.Rules[1], expectedRule2)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleKeyframes(t *testing.T) {
	input := `@keyframes identifier {
  0% { top: 0; left: 0; }
  100% { top: 100px; left: 100%; }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@keyframes",
		Prelude: "identifier",
		Rules: []*css.Rule{
			{
				Kind:    css.QualifiedRule,
				Prelude: "0%",
				Selectors: []*css.Selector{
					{
						Value:  "0%",
						Line:   2,
						Column: 3,
					},
				},
				Declarations: []*css.Declaration{
					{
						Property: "top",
						Value:    "0",
						Line:     2,
						Column:   8,
					},
					{
						Property: "left",
						Value:    "0",
						Line:     2,
						Column:   16,
					},
				},
			},
			{
				Kind:    css.QualifiedRule,
				Prelude: "100%",
				Selectors: []*css.Selector{
					{
						Value:  "100%",
						Line:   3,
						Column: 3,
					},
				},
				Declarations: []*css.Declaration{
					{
						Property: "top",
						Value:    "100px",
						Line:     3,
						Column:   10,
					},
					{
						Property: "left",
						Value:    "100%",
						Line:     3,
						Column:   22,
					},
				},
			},
		},
	}

	expectedOutput := `@keyframes identifier {
  0% {
    top: 0;
    left: 0;
  }
  100% {
    top: 100px;
    left: 100%;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleMedia(t *testing.T) {
	input := `@media screen, print {
  body { line-height: 1.2 }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@media",
		Prelude: "screen, print",
		Rules: []*css.Rule{
			{
				Kind:    css.QualifiedRule,
				Prelude: "body",
				Selectors: []*css.Selector{
					{
						Value:  "body",
						Line:   2,
						Column: 3,
					},
				},
				Declarations: []*css.Declaration{
					{
						Property: "line-height",
						Value:    "1.2",
						Line:     2,
						Column:   10,
					},
				},
			},
		},
	}

	expectedOutput := `@media screen, print {
  body {
    line-height: 1.2;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleNamespace(t *testing.T) {
	input := `@namespace svg url(http://www.w3.org/2000/svg);`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@namespace",
		Prelude: "svg url(http://www.w3.org/2000/svg)",
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRulePage(t *testing.T) {
	input := `@page :left {
  margin-left: 4cm;
  margin-right: 3cm;
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@page",
		Prelude: ":left",
		Declarations: []*css.Declaration{
			{
				Property: "margin-left",
				Value:    "4cm",
				Line:     2,
				Column:   3,
			},
			{
				Property: "margin-right",
				Value:    "3cm",
				Line:     3,
				Column:   3,
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleSupports(t *testing.T) {
	input := `@supports (animation-name: test) {
    /* specific CSS applied when animations are supported unprefixed */
    @keyframes { /* @supports being a CSS conditional group at-rule, it can includes other relevent at-rules */
      0% { top: 0; left: 0; }
      100% { top: 100px; left: 100%; }
    }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@supports",
		Prelude: "(animation-name: test)",
		Rules: []*css.Rule{
			{
				Kind: css.AtRule,
				Name: "@keyframes",
				Rules: []*css.Rule{
					{
						Kind:    css.QualifiedRule,
						Prelude: "0%",
						Selectors: []*css.Selector{
							{
								Value:  "0%",
								Line:   4,
								Column: 7,
							},
						},
						Declarations: []*css.Declaration{
							{
								Property: "top",
								Value:    "0",
								Line:     4,
								Column:   12,
							},
							{
								Property: "left",
								Value:    "0",
								Line:     4,
								Column:   20,
							},
						},
					},
					{
						Kind:    css.QualifiedRule,
						Prelude: "100%",
						Selectors: []*css.Selector{
							{
								Value:  "100%",
								Line:   5,
								Column: 7,
							},
						},
						Declarations: []*css.Declaration{
							{
								Property: "top",
								Value:    "100px",
								Line:     5,
								Column:   14,
							},
							{
								Property: "left",
								Value:    "100%",
								Line:     5,
								Column:   26,
							},
						},
					},
				},
			},
		},
	}

	expectedOutput := `@supports (animation-name: test) {
  @keyframes {
    0% {
      top: 0;
      left: 0;
    }
    100% {
      top: 100px;
      left: 100%;
    }
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestParseDeclarations(t *testing.T) {
	input := `color: blue; text-decoration:underline;`

	declarations, err := ParseDeclarations(input)
	if err != nil {
		t.Fatal("Failed to parse Declarations:", input)
	}

	expectedOutput := []*css.Declaration{
		{
			Property: "color",
			Value:    "blue",
			Line:     1,
			Column:   1,
		},
		{
			Property: "text-decoration",
			Value:    "underline",
			Line:     1,
			Column:   14,
		},
	}

	if len(declarations) != len(expectedOutput) {
		t.Fatal("Failed to parse Declarations:", input)
	}

	for i, decl := range declarations {
		if !decl.Equal(expectedOutput[i]) {
			t.Fatal("Failed to parse Declarations: ", decl.Str(true), expectedOutput[i].Str(true))
		}
	}
}

func TestMultipleDeclarations(t *testing.T) {
	input := `.btn:focus,
.btn:active:focus,
.btn.active:focus,
.btn.focus,
.btn:active.focus,
.btn.active.focus {
}`
	expectedRule := &css.Rule{
		Kind: css.QualifiedRule,
		Prelude: `.btn:focus,
.btn:active:focus,
.btn.active:focus,
.btn.focus,
.btn:active.focus,
.btn.active.focus`,
		Selectors: []*css.Selector{
			{
				Value:  ".btn:focus",
				Line:   1,
				Column: 1,
			},
			{
				Value:  ".btn:active:focus",
				Line:   2,
				Column: 1,
			},
			{
				Value:  ".btn.active:focus",
				Line:   3,
				Column: 1,
			},
			{
				Value:  ".btn.focus",
				Line:   4,
				Column: 1,
			},
			{
				Value:  ".btn:active.focus",
				Line:   5,
				Column: 1,
			},
			{
				Value:  ".btn.active.focus",
				Line:   6,
				Column: 1,
			},
		},
		Declarations: []*css.Declaration{},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)
}

func TestComments(t *testing.T) {
	input := "td /* © */ { color /* © */: red; }"
	expectedRule := &css.Rule{
		Kind:    css.QualifiedRule,
		Prelude: "td /* © */",
		Selectors: []*css.Selector{
			{
				Value:  "td",
				Line:   1,
				Column: 1,
			},
		},
		Declarations: []*css.Declaration{
			{
				Property: "color",
				Value:    "red",
				Line:     1,
				Column:   14,
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)
}

func TestInfiniteLoop(t *testing.T) {
	input := "{;}"
	_, err := Parse(input)
	if err == nil {
		t.Fatal("Expected an error got nil")
	}
}
