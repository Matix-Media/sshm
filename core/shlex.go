package core

import "strings"

// shlexSplit splits a command line into fields the way a POSIX shell would for
// simple cases: whitespace separates fields, and single or double quotes group
// text (including spaces). Unterminated quotes are tolerated — the run so far is
// returned as a field. This covers ssh_config values and custom commands
// without pulling in a dependency.
func shlexSplit(input string) []string {
	var fields []string
	var cur strings.Builder
	inField := false
	var quote rune // 0, '\'' or '"'

	flush := func() {
		if inField {
			fields = append(fields, cur.String())
			cur.Reset()
			inField = false
		}
	}

	for _, r := range input {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inField = true
		case r == ' ' || r == '\t':
			flush()
		default:
			cur.WriteRune(r)
			inField = true
		}
	}
	flush()
	return fields
}
