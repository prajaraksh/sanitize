// Package sanitize goal is to have same fileName irrespective of platform.
// Incase of invalid fileName, a md5Sum of provided file is returned
package sanitize

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Sanitize struct
type Sanitize struct {
	replaceStr string
	maxLen     int
}

var (
	// ErrCntrl represents `replaceStr` contains Control Characters
	ErrCntrl = errors.New("replaceStr can't contain control characters")
	// ErrInval represents `replaceStr` contains Invalid Characters
	ErrInval = errors.New("replaceStr can't contain ., <, >, :, \", /, \\, |, ?, *")

	// ErrLen represents `maxLen` is greater than 255
	ErrLen = errors.New("maxLen can't be greater than 255")

	errGroupLen = errors.New("No.of submatch groups doesn't match with replaceGroup")
)

var iMaxLen = 240

// New returns Sanitize
func New() *Sanitize {
	return &Sanitize{"", iMaxLen}
}

var (
	cntrlExp = regexp.MustCompile("[[:cntrl:]]") // control

	// invalid characters - windows
	// <, >, :, ", /, \, |, ?, *
	invCharExp = regexp.MustCompile(`[<>:"/\\|\?\*]+`)

	// trim right spaces and dot
	rightSDExp = regexp.MustCompile("(?s:[[:space:]]|\\.)+$")

	// trim left dot's
	leftdotExpr = regexp.MustCompile("^\\.+")
)

// NewWithOpts returns Sanitize with opts set
// Certain conditions exist for replace Str and
func NewWithOpts(replaceStr string, maxLen int) (*Sanitize, error) {
	s := &Sanitize{"", iMaxLen}

	if cntrlExp.MatchString(replaceStr) {
		return s, ErrCntrl
	}

	if invCharExp.MatchString(replaceStr) || strings.Contains(replaceStr, ".") {
		return s, ErrInval
	}

	if maxLen > 255 {
		return s, ErrLen
	}

	return &Sanitize{replaceStr, maxLen}, nil
}

// Name sanitizies `fileName`
func (s *Sanitize) Name(fileName string) string {
	return name(fileName, s.replaceStr, s.maxLen)
}

func name(fileName, replaceStr string, maxLen int) string {

	intrStr := strings.ToValidUTF8(fileName, replaceStr) // intermediate string
	intrStr = html.UnescapeString(intrStr)
	intrStr = cntrlExp.ReplaceAllString(intrStr, replaceStr)
	intrStr = invCharExp.ReplaceAllString(intrStr, replaceStr)
	intrStr = rightSDExp.ReplaceAllString(intrStr, replaceStr)
	intrStr = leftdotExpr.ReplaceAllString(intrStr, ".")

	return validName(fileName, intrStr, replaceStr, maxLen)
}

var (
	// Unicode categories

	// Cc - Control
	// Cf - Format
	unicodeControl = regexp.MustCompile("\\p{Cc}|\\p{Cf}")

	// Zl - Line separator
	// Zp - Paragraph separator
	// Zs - Space separator
	unicodeSpace = regexp.MustCompile("\\p{Zl}|\\p{Zp}|\\p{Zs}")
)

var (
	// if below, characters are repeated more than twice,
	// we replace it with single character from `cleanReplaceWith`
	cleanExpr = repeatedCharsExp([]string{
		`[[:space:]]`, `_`, `-`, `\+`, `\.`, `!`})
	cleanReplaceWith = []string{"", " ", "_", "-", "+", ".", "!"}
)

// Clean sanitizes ,removes invisible characters and repeated separators
func (s *Sanitize) Clean(fileName string) string {
	return clean(fileName, s.replaceStr, s.maxLen)
}

func clean(fileName, replaceStr string, maxLen int) string {

	intrStr := strings.ToValidUTF8(fileName, replaceStr)
	intrStr = html.UnescapeString(intrStr)

	// invisible characters
	intrStr = unicodeControl.ReplaceAllString(intrStr, replaceStr)
	intrStr = unicodeSpace.ReplaceAllString(intrStr, " ")

	intrStr = invCharExp.ReplaceAllString(intrStr, replaceStr)

	// repeated separators
	intrStr, _ = replaceAllStringSubmatch(
		cleanExpr, intrStr, cleanReplaceWith)
	var replaceExpr *regexp.Regexp

	// remove repeated `replaceStr`
	if replaceStr != "" {
		replaceExpr = regexp.MustCompile(replaceStr + "{2,}")
		intrStr = replaceExpr.ReplaceAllString(intrStr, replaceStr)
	}

	intrStr = rightSDExp.ReplaceAllString(intrStr, replaceStr)

	// once again remove any repeated `replaceStr`
	if replaceExpr != nil {
		intrStr = replaceExpr.ReplaceAllString(intrStr, replaceStr)
	}

	intrStr = leftdotExpr.ReplaceAllString(intrStr, ".")

	return validName(fileName, intrStr, replaceStr, maxLen)
}

func validName(fileName, intrStr, replaceStr string, maxLen int) string {

	if intrStr != replaceStr {
		if valid(intrStr) {
			if len(intrStr) > maxLen {
				return strings.ToValidUTF8(intrStr[:maxLen], "")
			}
			return intrStr
		}
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(fileName)))
}

func valid(name string) bool {
	if name == "" {
		return false
	}

	if len(name) < 5 {
		for _, invName := range invalidNamesMap[len(name)] {
			if strings.ToLower(name) == invName {
				return false
			}
		}
	}

	return true
}

// https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file?redirectedfrom=MSDN#naming-conventions
// invalidNamesMap with len of value as key
var invalidNamesMap = map[int][]string{
	1: {
		".",
	},

	2: {
		"..",
	},

	3: {
		"con", "prn", "aux", "nul",
	},

	4: {
		"com1", "com2", "com3", "com4",
		"com5", "com6", "com7", "com8",
		"com9", "lpt1", "lpt2", "lpt3",
		"lpt4", "lpt5", "lpt6", "lpt7",
		"lpt8", "lpt9",
	},
}

func repeatedCharsExp(vals []string) *regexp.Regexp {
	var s strings.Builder
	for i, val := range vals {
		if i == 0 {
			s.WriteString("(" + val + "{2,})")
		} else {
			s.WriteString("|(" + val + "{2,})")
		}

	}
	return regexp.MustCompile(s.String())
}

// replaceAllStringSubmatch replaces matched groups with `replaceGroup`
func replaceAllStringSubmatch(re *regexp.Regexp, src string, replaceGroup []string) (string, error) {

	sms := re.FindAllStringSubmatchIndex(src, -1)

	if len(sms) == 0 {
		return src, nil
	}

	if len(sms[0]) != len(replaceGroup)*2 {
		return "", errGroupLen
	}

	var s strings.Builder

	prevPos := 0
	for _, sm := range sms {

		for i := 2; i < len(sm); i += 2 {
			if sm[i] != -1 {
				start := sm[i]
				end := sm[i+1]

				s.WriteString(src[prevPos:start] + replaceGroup[i/2])

				prevPos = end
			}
		}
	}
	s.WriteString(src[prevPos:])
	return s.String(), nil
}
