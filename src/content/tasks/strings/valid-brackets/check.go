package main

func TestBalancedTrue(t *testing.T) {
	for _, s := range []string{"", "()", "([]{})", "{[()()]}", "func() { return a[0] }"} {
		if !Balanced(s) {
			t.Fatalf("%q сбалансирована", s)
		}
	}
}

func TestBalancedWrongNesting(t *testing.T) {
	for _, s := range []string{"(]", "([)]", "{(})"} {
		if Balanced(s) {
			t.Fatalf("%q — скобки вложены неправильно", s)
		}
	}
}

func TestBalancedCounts(t *testing.T) {
	for _, s := range []string{"((", ")(", "())", "[", "}"} {
		if Balanced(s) {
			t.Fatalf("%q — скобки не сбалансированы", s)
		}
	}
}
