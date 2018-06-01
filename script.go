package migrate

import "strings"

// for simplicity sake ignores edge cases,
// it will error-out on incorrect syntax anyway later;
// not production-ready solution

func SplitScript(script string) []string {
	var stmts []string
	s := splitter{str: []rune(script)}
	for s.Scan() {
		stmts = append(stmts, s.stmt)
	}
	return stmts
}

type splitter struct {
	str []rune

	i1, i2 int
	c1, c2 rune // s[i2-1], s[i2]

	inCurved       int  // (
	inLineComment  bool // --
	inMultiComment bool // $$
	inDoubleQuotes bool // "
	inQuote        bool // '
	inWtf          bool // `

	stmt string
}

func (s *splitter) Scan() bool {
	for s.step() {
		if s.updateState() && s.c2 == ';' {
			s.stmt = strings.TrimSpace(string(s.str[s.i1:s.i2]))
			s.i2++
			s.i1 = s.i2
			return true
		}
	}

	s.stmt = ""
	return false
}

func (s *splitter) Stmt() string {
	return s.stmt
}

func (s *splitter) step() bool {
	s.i2++
	if s.i2 >= len(s.str) {
		return false
	}
	s.c1, s.c2 = s.str[s.i2-1], s.str[s.i2]
	return true
}

// updates the comment/quote state and returns true if in global scope
func (s *splitter) updateState() bool {
	if s.inComment() || s.inQuotes() {
		switch {
		case s.inLineComment && s.c2 == '\n':
			s.inLineComment = false
		case s.inMultiComment && s.c2 == '$' && s.c1 == '$':
			s.inMultiComment = false
		case s.inDoubleQuotes && s.c2 == '"':
			s.inDoubleQuotes = false
		case s.inQuote && s.c2 == '\'':
			s.inQuote = false
		case s.inWtf && s.c2 == '`':
			s.inWtf = false
		}
	} else {
		switch {
		case s.c2 == '-' && s.c1 == '-':
			s.inLineComment = true
		case s.c2 == '$' && s.c1 == '$':
			s.inMultiComment = true
		case s.c2 == '"':
			s.inDoubleQuotes = true
		case s.c2 == '\'':
			s.inQuote = true
		case s.c2 == '`':
			s.inWtf = true
		}
	}

	return !(s.inComment() || s.inQuotes())
}

func (s *splitter) inComment() bool {
	return s.inLineComment || s.inMultiComment
}

func (s *splitter) inQuotes() bool {
	return s.inDoubleQuotes || s.inQuote || s.inWtf
}
