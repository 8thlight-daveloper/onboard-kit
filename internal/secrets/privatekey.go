package secrets

import "regexp"

var privateKeyHeader = regexp.MustCompile(`-----BEGIN (?:[A-Z]+ )*PRIVATE KEY( BLOCK)?-----`)

func hasPrivateKeyHeader(text string) bool {
	return privateKeyHeader.MatchString(text)
}
