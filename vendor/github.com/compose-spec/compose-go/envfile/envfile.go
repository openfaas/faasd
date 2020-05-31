/*
   Copyright 2020 The Compose Specification Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package envfile

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/compose-spec/compose-go/types"
)

const whiteSpaces = " \t"

// ErrBadKey typed error for bad environment variable
type ErrBadKey struct {
	msg string
}

func (e ErrBadKey) Error() string {
	return fmt.Sprintf("poorly formatted environment: %s", e.msg)
}

// Parse reads a file with environment variables enumerated by lines
//
// ``Environment variable names used by the utilities in the Shell and
// Utilities volume of IEEE Std 1003.1-2001 consist solely of uppercase
// letters, digits, and the '_' (underscore) from the characters defined in
// Portable Character Set and do not begin with a digit. *But*, other
// characters may be permitted by an implementation; applications shall
// tolerate the presence of such names.''
// -- http://pubs.opengroup.org/onlinepubs/009695399/basedefs/xbd_chap08.html
//
// As of #16585, it's up to application inside docker to validate or not
// environment variables, that's why we just strip leading whitespace and
// nothing more.
// Converts ["key=value"] to {"key":"value"} but set unset keys - the ones with no "=" in them - to nil
// We use this in cases where we need to distinguish between FOO=  and FOO
// where the latter case just means FOO was mentioned but not given a value
func Parse(filename string) (types.MappingWithEquals, error) {
	vars := types.MappingWithEquals{}
	fh, err := os.Open(filename)
	if err != nil {
		return vars, err
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	currentLine := 0
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	for scanner.Scan() {
		scannedBytes := scanner.Bytes()
		if !utf8.Valid(scannedBytes) {
			return vars, fmt.Errorf("env file %s contains invalid utf8 bytes at line %d: %v", filename, currentLine+1, scannedBytes)
		}
		// We trim UTF8 BOM
		if currentLine == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}
		// trim the line from all leading whitespace first
		line := strings.TrimLeftFunc(string(scannedBytes), unicode.IsSpace)
		currentLine++
		// line is not empty, and not starting with '#'
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			data := strings.SplitN(line, "=", 2)

			// trim the front of a variable, but nothing else
			variable := strings.TrimLeft(data[0], whiteSpaces)
			if strings.ContainsAny(variable, whiteSpaces) {
				return vars, ErrBadKey{fmt.Sprintf("variable '%s' contains whitespaces", variable)}
			}
			if len(variable) == 0 {
				return vars, ErrBadKey{fmt.Sprintf("no variable name on line '%s'", line)}
			}

			if len(data) > 1 {
				// pass the value through, no trimming
				vars[variable] = &data[1]
			} else {
				// variable was not given a value but declared
				vars[strings.TrimSpace(line)] = nil
			}
		}
	}
	return vars, scanner.Err()
}
