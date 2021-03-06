// package constraints contains methods for decoding a compact CPU and Memory
// constraints format. We specify the "constraints" format as the following:
//
// <cpushare>:<memory limit>
//
// CPUShare can be any number between 2 and 1024. For more details on how the
// --cpu-shares flag works in Docker/cgroups, see
// https://docs.docker.com/reference/run/#cpu-share-constraint
//
// Memory limit can contain a number and optionally the units. The following are
// all equivalent:
//
//	6GB
//	6144MB
//	6291456KB
package constraints

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	. "github.com/remind101/empire/empire/pkg/bytesize"
)

// ConstraintsSeparator separates the individual resource constraints
const ConstraintsSeparator = ":"

var (
	ErrInvalidCPUShare   = errors.New("CPUShare must be a value between 2 and 1024")
	ErrInvalidMemory     = errors.New("invalid memory format")
	ErrInvalidConstraint = errors.New("invalid constraints format")
)

// bytes is used as a multiplier.
const bytes = uint(1)

// CPUShare represents a CPUShare.
type CPUShare int

// NewCPUShare casts i to a CPUShare and ensures it's validity.
func NewCPUShare(i int) (CPUShare, error) {
	if i < 2 || i > 1024 {
		return 0, ErrInvalidCPUShare
	}

	return CPUShare(i), nil
}

func ParseCPUShare(s string) (CPUShare, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return NewCPUShare(i)
}

// memRegex parses the number of units from a string.
var memRegex = regexp.MustCompile(`(\d+)(\S*)?`)

// Memory represents a memory limit.
type Memory uint

// ParseMemory parses a string in memory format and returns the amount of memory
// in bytes.
func ParseMemory(s string) (Memory, error) {
	i, err := parseMemory(s)
	return Memory(i), err
}

// String returns the string representation of Memory, using the following
// algorithm:
//
// * If the memory is less than 1 KB, it will return "x".
// * If the memory is less than 1 MB, it will return "xKB".
// * If the memory is less than 1 GB, it will return "xMB".
// * etc
func (m Memory) String() string {
	v := uint(m)

	switch {
	case v < KB:
		return fmt.Sprintf("%d", v)
	case v < MB:
		return fmtMemory(m, KB)
	case v < GB:
		return fmtMemory(m, MB)
	case v < TB:
		return fmtMemory(m, GB)
	}

	return fmt.Sprintf("%d", v)
}

func fmtMemory(m Memory, units uint) string {
	var u string
	switch units {
	case KB:
		u = "kb"
	case MB:
		u = "mb"
	case GB:
		u = "gb"
	case TB:
		u = "tb"

	}
	return fmt.Sprintf("%.2f%s", float32(m)/float32(units), u)
}

func parseMemory(s string) (uint, error) {
	p := memRegex.FindStringSubmatch(s)

	var (
		// n is the number part of the memory
		n uint
		// u is the units parts
		u string
		// mult is a number that will be used to
		// multiply n to return bytes.
		mult uint
	)

	if len(p) == 0 {
		return n, ErrInvalidMemory
	}

	i, err := strconv.Atoi(p[1])
	if err != nil {
		return n, err
	}

	n = uint(i)

	if len(p) > 2 {
		u = strings.ToUpper(p[2])
	}

	switch u {
	case "":
		mult = bytes
	case "KB":
		mult = KB
	case "MB":
		mult = MB
	case "GB":
		mult = GB
	case "TB":
		mult = TB
	default:
		return n, ErrInvalidMemory
	}

	return n * mult, nil
}

// Constraints is a composition of CPUShares and Memory constraints.
type Constraints struct {
	CPUShare
	Memory
}

func Parse(s string) (Constraints, error) {
	var c Constraints

	p := strings.SplitN(s, ConstraintsSeparator, 2)
	if len(p) != 2 {
		return c, ErrInvalidConstraint
	}

	i, err := ParseCPUShare(p[0])
	if err != nil {
		return c, err
	}

	c.CPUShare = i

	m, err := ParseMemory(p[1])
	if err != nil {
		return c, err
	}

	c.Memory = m

	return c, nil
}
