package tui

import "testing"

func TestTheRailUsesShapesOnAUnicodeTerminalAndLettersOtherwise(t *testing.T) {
	cases := []struct {
		name        string
		locale      map[string]string
		wantUnicode bool
	}{
		{"unset", map[string]string{"LC_ALL": "", "LC_CTYPE": "", "LANG": ""}, true},
		{"utf-8", map[string]string{"LC_ALL": "", "LC_CTYPE": "", "LANG": "en_US.UTF-8"}, true},
		{"utf8 spelled plain", map[string]string{"LC_ALL": "C.utf8", "LC_CTYPE": "", "LANG": ""}, true},
		{"the C locale", map[string]string{"LC_ALL": "", "LC_CTYPE": "", "LANG": "C"}, false},
		{"LC_ALL wins", map[string]string{"LC_ALL": "C", "LC_CTYPE": "", "LANG": "en_US.UTF-8"}, false},
	}
	for _, each := range cases {
		for name, value := range each.locale {
			t.Setenv(name, value)
		}
		got := railGlyphsForTerminal()
		want := plainRailGlyphs()
		if each.wantUnicode {
			want = unicodeRailGlyphs()
		}
		if got != want {
			t.Errorf("%s: got %+v, want %+v", each.name, got, want)
		}
	}
}

func TestEveryRailGlyphIsDistinct(t *testing.T) {
	t.Parallel()
	for name, glyphs := range map[string]railGlyphs{"unicode": unicodeRailGlyphs(), "plain": plainRailGlyphs()} {
		seen := map[string]bool{}
		for _, glyph := range []string{glyphs.answered, glyphs.deferred, glyphs.open, glyphs.here} {
			if seen[glyph] {
				t.Errorf("%s repeats %q", name, glyph)
			}
			seen[glyph] = true
		}
	}
}
