package secrets

import (
	"regexp"
	"strings"
)

type tokenFamily struct {
	prefixes   []string
	body       func(character byte) bool
	minimumLen int
}

func tokenFamilies() []tokenFamily {
	return []tokenFamily{
		{[]string{"ghp_", "gho_", "ghu_", "ghs_", "ghr_"}, isAlphanumeric, 36},
		{[]string{"github_pat_"}, isWordCharacter, 22},
		{[]string{"AKIA", "ASIA"}, isUpperOrDigit, 16},
		{[]string{"xoxa-", "xoxb-", "xoxp-", "xoxr-", "xoxs-"}, isAlphanumericOrHyphen, 10},
		{[]string{"sk_live_", "rk_live_"}, isAlphanumeric, 16},
		{[]string{"AIza"}, isURLSafe, 35},
		{[]string{"glpat-"}, isURLSafe, 20},
		{[]string{"npm_"}, isAlphanumeric, 36},
		{[]string{"sk-"}, isURLSafe, 40},
	}
}

var webToken = regexp.MustCompile(`(?:^|[^A-Za-z0-9])eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`)

func hasTokenPrefix(text string) bool {
	if webToken.MatchString(text) {
		return true
	}
	for _, family := range tokenFamilies() {
		for _, prefix := range family.prefixes {
			if hasPrefixedBody(text, prefix, family) {
				return true
			}
		}
	}
	return false
}

func hasPrefixedBody(text, prefix string, family tokenFamily) bool {
	for offset := 0; ; {
		found := strings.Index(text[offset:], prefix)
		if found < 0 {
			return false
		}
		start := offset + found
		if startsToken(text, start) && bodyLength(text[start+len(prefix):], family.body) >= family.minimumLen {
			return true
		}
		offset = start + 1
	}
}

func startsToken(text string, start int) bool {
	return start == 0 || !isAlphanumeric(text[start-1])
}

func bodyLength(rest string, belongs func(byte) bool) int {
	length := 0
	for length < len(rest) && belongs(rest[length]) {
		length++
	}
	return length
}
