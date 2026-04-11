package domain

import (
	"strings"
)

const (
	AnsiReset  string = "\x1b[0m"
	AnsiKey    string = "\x1b[36m" // cyan
	AnsiString string = "\x1b[32m" // green
	AnsiNumber string = "\x1b[33m" // yellow
	AnsiBool   string = "\x1b[35m" // magenta
	AnsiPunct  string = "\x1b[90m" // bright black / gray
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (d *Formatter) ColorizeJSON(s string) string {
	var out strings.Builder
	out.Grow(len(s) * 2)

	i := 0
	for i < len(s) {
		c := s[i]

		switch {
		case c == '"':
			// Find matching end quote (respect escapes)
			end := i + 1
			for end < len(s) {
				if s[end] == '\\' && end+1 < len(s) {
					end += 2
					continue
				}

				if s[end] == '"' {
					break
				}

				end++
			}

			if end >= len(s) {
				// Unterminated, emit the rest
				out.WriteString(s[i:])
				return out.String()
			}

			str := s[i : end+1]

			// Is this a key? Look past whitespace for ':'
			j := end + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}

			if j < len(s) && s[j] == ':' {
				out.WriteString(AnsiKey)
			} else {
				out.WriteString(AnsiString)
			}

			out.WriteString(str)
			out.WriteString(AnsiReset)
			i = end + 1

		case c == '-' || (c >= '0' && c <= '9'):
			end := i
			if c == '-' {
				end++
			}

			for end < len(s) {
				d := s[end]
				if (d >= '0' && d <= '9') || d == '.' || d == 'e' || d == 'E' || d == '+' || d == '-' {
					end++
					continue
				}

				break
			}

			out.WriteString(AnsiNumber)
			out.WriteString(s[i:end])
			out.WriteString(AnsiReset)
			i = end

		case c == 't' && strings.HasPrefix(s[i:], "true"):
			out.WriteString(AnsiBool)
			out.WriteString("true")
			out.WriteString(AnsiReset)
			i += 4

		case c == 'f' && strings.HasPrefix(s[i:], "false"):
			out.WriteString(AnsiBool)
			out.WriteString("false")
			out.WriteString(AnsiReset)
			i += 5

		case c == 'n' && strings.HasPrefix(s[i:], "null"):
			out.WriteString(AnsiBool)
			out.WriteString("null")
			out.WriteString(AnsiReset)
			i += 4

		case c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':':
			out.WriteString(AnsiPunct)
			out.WriteByte(c)
			out.WriteString(AnsiReset)
			i++

		default:
			out.WriteByte(c)
			i++
		}
	}

	return out.String()
}
