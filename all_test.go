package main

import "testing"

func TestIsFlanking(t *testing.T) {
	type Case struct {
		s string
		l bool
		r bool
	}

	cases := []Case{
		{`***abc`, true, false},
		{`  _abc`, true, false},
		{`**"abc"`, true, false},
		{` _"abc"`, true, false},

		{` abc***`, false, true},
		{` abc_`, false, true},
		{`"abc"**`, false, true},
		{`"abc"_`, false, true},

		{` abc***def`, true, true},
		{`"abc"_"def"`, true, true},

		{`abc *** def`, false, false},
		{`a _ b`, false, false},
	}

	for _, c := range cases {
		_, delimiters := parseInlinesToDeimiters(c.s)
		for de := delimiters.Back(); de != nil; de = de.Prev() {
			d := de.Value.(*Delimiter)
			l := d.isLeftFlanking()
			r := d.isRightFlanking()
			if l != c.l || r != c.r {
				t.Errorf("%20s: %v\t%v\n", c.s, l, r)
			}
		}
	}
}
