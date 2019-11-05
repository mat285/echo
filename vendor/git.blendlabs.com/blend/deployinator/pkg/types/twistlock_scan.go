package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
)

const (
	// TwistlockScanKeyPrefix is the prefix for configmap key of a twistlock scan result
	TwistlockScanKeyPrefix = "twistlock-scan"
)

// TwistlockScan is a twistlock scan result
type TwistlockScan struct {
	Time   time.Time
	Image  string
	Result string
}

var (
	twistlockScanKey = regexp.MustCompile(`(^.+)\.([[:alnum:]]|(?:[[:alnum:]][[:word:]-]*?[[:alnum:]]))(_(\d+))?$`)
)

// ParseTwistlockScan creates a TwistlockScan object based on a key-value pair
func ParseTwistlockScan(key, value string) (TwistlockScan, error) {
	if strings.HasPrefix(key, TwistlockScanKeyPrefix) {
		matches := twistlockScanKey.FindStringSubmatch(key[len(TwistlockScanKeyPrefix)+1:])
		if matches == nil {
			return TwistlockScan{}, exception.New("Invalid twistlock scan key format")
		}
		var t time.Time
		if len(matches) > 4 && len(matches[4]) > 0 {
			timestamp, err := strconv.ParseInt(matches[4], 10, 64) // 64 bit int
			if err != nil {
				return TwistlockScan{}, exception.New(err)
			}
			t = time.Unix(timestamp, 0)
		}
		return TwistlockScan{
			Time:   t,
			Image:  fmt.Sprintf("%s:%s", matches[1], matches[2]),
			Result: value,
		}, nil
	}
	return TwistlockScan{}, nil
}

// IsZero tests the zero value
func (t TwistlockScan) IsZero() bool {
	return t == (TwistlockScan{})
}

// Key generates the serializable key for the object
func (t TwistlockScan) Key() string {
	return fmt.Sprintf("%s.%s_%d", TwistlockScanKeyPrefix, strings.Replace(t.Image, ":", ".", -1), t.Time.Unix())
}

// TwistlockScansByTime is a sortable interface for a list of TwistlockScans
type TwistlockScansByTime []TwistlockScan

func (t TwistlockScansByTime) Len() int {
	return len(t)
}

func (t TwistlockScansByTime) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TwistlockScansByTime) Less(i, j int) bool {
	return t[i].Time.Before(t[j].Time)
}
