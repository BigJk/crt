package crt

import (
	"fmt"
	"github.com/muesli/termenv"
	"strings"
	"sync"
)

var sgrMtx = &sync.Mutex{}
var sgrCache = map[string][]any{}

// extractSGR extracts an SGR ansi sequence from the beginning of the string.
func extractSGR(s string) (string, bool) {
	if len(s) < 2 {
		return "", false
	}

	if !strings.HasPrefix(s, termenv.CSI) {
		return "", false
	}

	for i := 2; i < len(s); i++ {
		if s[i] == ' ' || s[i] == termenv.CSI[0] {
			return "", false
		}

		if s[i] == 'm' {
			return s[:i+1], true
		}
	}

	return "", false
}

type SGRReset struct{}

type SGRBold struct{}

type SGRUnsetBold struct{}

type SGRItalic struct{}

type SGRUnsetItalic struct{}

type SGRFgTrueColor struct {
	R, G, B byte
}

type SGRBgTrueColor struct {
	R, G, B byte
}

type SGRFgColor struct {
	Id int
}

type SGRBgColor struct {
	Id int
}

// parseSGR parses a single SGR ansi sequence and returns a struct representing the sequence.
func parseSGR(s string) ([]any, bool) {
	if !strings.HasPrefix(s, termenv.CSI) {
		return nil, false
	}

	s = s[len(termenv.CSI):]
	if len(s) == 0 {
		return nil, false
	}

	sgrMtx.Lock()
	if cached, ok := sgrCache[s]; ok {
		sgrMtx.Unlock()
		return cached, true
	}
	sgrMtx.Unlock()

	full := s

	if !strings.HasSuffix(s, "m") {
		return nil, false
	}

	s = s[:len(s)-1]
	if len(s) == 0 {
		return nil, false
	}

	var skips int
	var res []any
	for len(s) > 0 {
		code := strings.SplitN(s, ";", 2)[0]

		if skips > 0 {
			skips--
		} else {
			switch code {
			case "0":
				res = append(res, SGRReset{})
			case "1":
				res = append(res, SGRBold{})
			case "22":
				res = append(res, SGRUnsetBold{})
			case "3":
				res = append(res, SGRItalic{})
			case "23":
				res = append(res, SGRUnsetItalic{})
			default:
				if strings.HasPrefix(s, "38;2;") {
					var r, g, b byte
					_, err := fmt.Sscanf(s, "38;2;%d;%d;%d", &r, &g, &b)
					if err == nil {
						skips = 4
						res = append(res, SGRFgTrueColor{r, g, b})
						continue
					}
				} else if strings.HasPrefix(s, "48;2;") {
					var r, g, b byte
					_, err := fmt.Sscanf(s, "48;2;%d;%d;%d", &r, &g, &b)
					if err == nil {
						skips = 4
						res = append(res, SGRBgTrueColor{r, g, b})
						continue
					}
				} else if strings.HasPrefix(s, "38;5;") {
					var id int
					_, err := fmt.Sscanf(s, "38;5;%d", &id)
					if err == nil {
						skips = 2
						res = append(res, SGRFgColor{id})
						continue
					}
				} else if strings.HasPrefix(s, "48;5;") {
					var id int
					_, err := fmt.Sscanf(s, "48;5;%d", &id)
					if err == nil {
						skips = 2
						res = append(res, SGRBgColor{id})
						continue
					}
				}
			}
		}

		if len(code) >= len(s) {
			break
		}

		s = s[len(code)+1:]
	}

	sgrMtx.Lock()
	sgrCache[full] = res
	sgrMtx.Unlock()

	return res, len(res) > 0
}
