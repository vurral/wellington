package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func printItems(items []Item) {
	for i, item := range items {
		fmt.Printf("%4d: %s %s\n", i, item.Type, item.Value)
	}
}

func TestLexerBools(t *testing.T) {
	if IsEOF('%', 0) != true {
		t.Errorf("Did not detect EOF")
	}
}

func TestLexer(t *testing.T) {

	fvar, _ := ioutil.ReadFile("test/_var.scss")

	items, err := testParse(string(fvar))

	if err != nil {
		t.Errorf("Error parsing string")
	}

	sel := items[0].String()
	if e := "$images"; e != sel {
		t.Errorf("Invalid VAR parsing expected: %s, was: %s",
			e, sel)
	}

	if e := "sprite-map"; e != items[2].String() {
		t.Errorf("Invalid CMD parsing expected: %s, was: %s",
			e, items[1].String())
	}

	sel = items[4].String()
	if e := "*.png"; e != sel {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			e, sel)
	}

	T := items[4].Type
	if e := FILE; e != T {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			e, T)
	}

	sel = items[13].String()
	if e := "#00FF00"; e != sel {
		t.Errorf("Invalid TEXT parsing expected: %s, was: %s",
			e, sel)
	}
}

func TestLexerSub(t *testing.T) {
	in := `$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`
	items, err := testParse(in)

	if err != nil {
		panic(err)
	}
	if e := INT; items[9].Type != e {
		t.Errorf("Invalid token expected: %s, was: %s", e, items[7])
	}
	if e := SUB; items[10].Type != e {
		t.Errorf("Invalid token expected: %s, was: %s", e, items[8])
	}
}

func TestLexerCmds(t *testing.T) {
	in := `$s: sprite-map("test/*.png");
$file: sprite-file($s, 140);
div {
  width: image-width($file, 140);
  height: image-height(sprite-file($s, 140));
}`
	items, err := parse(in)
	if err != nil {
		panic(err)
	}

	types := map[int]ItemType{
		0:  VAR,
		1:  CMDVAR,
		3:  FILE,
		6:  VAR,
		7:  CMD,
		9:  SUB,
		10: FILE,
		15: TEXT,
		16: CMD,
		18: SUB,
		19: FILE,
		23: CMD,
		25: CMD,
		28: FILE,
	}

	for i, tp := range types {
		if tp != items[i].Type {
			t.Errorf("expected: %s, was: %s", tp, items[i].Type)
		}
	}
}

func TestLexerImport(t *testing.T) {
	fvar, _ := ioutil.ReadFile("test/import.scss")
	items, _ := testParse(string(fvar))
	sel := items[0].String()
	if e := "background"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[2].String()
	if e := "purple"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[4].String()
	if e := "@import"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[5].String()
	if e := "var"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
}

// Test disabled due to not working
func TestLexerSubModifiers(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  height: -1 * sprite-height($s,"140");
  width: -sprite-width($s,"140");
  margin: - sprite-height($s, "140")px;
}`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	if e := ":"; items[1].Value != e {
		t.Errorf("Failed to parse symbol expected: %s, was: %s",
			e, items[1].Value)
	}
	if e := "*.png"; items[4].Value != e {
		t.Errorf("Failed to parse file expected: %s, was: %s",
			e, items[4].Value)
	}

	if e := "*"; items[13].Value != e {
		t.Errorf("Failed to parse text expected: %s, was: %s",
			e, items[13].Value)
	}

	if e := MINUS; items[22].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[22].Type)
	}

	if e := CMD; items[23].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[23].Type)
	}

	if e := TEXT; items[37].Type != e {
		t.Errorf("Failed to parse TEXT expected: %s, was: %s",
			e, items[37].Type)
	}
}

func TestLexerVars(t *testing.T) {
	in := `$a: 1;
$b: $1;
$c: ();
$d: $c`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	_ = items
}

func TestLexerWhitespace(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  background:sprite($s,"140");
}`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	if e := TEXT; items[9].Type != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[9].Type)
	}

	if e := CMD; items[11].Type != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[11].Type)
	}

	if e := "sprite"; items[11].Value != e {
		t.Errorf("Command parsed improperly expected: %s, was: %s",
			e, items[11].Value)
	}
}

// create a parser for the language.
func testParse(input string) ([]Item, error) {
	lex := New(func(lex *Lexer) StateFn {
		return lex.Action()
	}, input)

	var status []Item
	for {
		item := lex.Next()
		err := item.Error()

		if err != nil {
			return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		switch item.Type {
		case ItemEOF:
			return status, nil
		case CMD, SPRITE, TEXT, VAR, FILE, SUB:
			fallthrough
		case LPAREN, RPAREN,
			LBRACKET, RBRACKET:
			fallthrough
		case IMPORT:
			fallthrough
		case EXTRA:
			status = append(status, *item)
		default:
			status = append(status, *item)
			//fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}
