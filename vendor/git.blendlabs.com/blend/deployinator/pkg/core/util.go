package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/airbrake"
	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/mathutil"
	"github.com/blend/go-sdk/ref"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/util/homedir"
)

// FileExists returns if a file exists or not.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirectoryExists returns if a directory exists or not.
func DirectoryExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ContainsString returns if the slice contains the string
func ContainsString(slice []string, s string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}
	return false
}

// DedupStrings deduplicates the string slice while maintaining order
func DedupStrings(strs []string) []string {
	ret := []string{}
	m := map[string]bool{}
	for _, str := range strs {
		if has, ok := m[str]; !ok || !has {
			ret = append(ret, str)
			m[str] = true
		}
	}
	return ret
}

// Chunk splits the byte array into slices of the specified max size
func Chunk(data []byte, maxSize int) [][]byte {
	numChunks := len(data)/maxSize + 1
	chunks := make([][]byte, 0, numChunks)

	for len(data) > 0 {
		nextChunkSize := MinInt(maxSize, len(data))
		chunks = append(chunks, data[:nextChunkSize])
		if nextChunkSize == len(data) {
			break
		}
		data = data[nextChunkSize:]
	}
	return chunks
}

// MinInt wraps util min int so that we can use it for 2 ints
func MinInt(ints ...int) int {
	return mathutil.MinInts(ints)
}

// MaxInt wraps util max int so we can use it for 2 ints
func MaxInt(ints ...int) int {
	return mathutil.MaxInts(ints)
}

// CopyFile copies a file from `src` to `dst`.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return exception.New(err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return exception.New(err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return exception.New(err)
	}
	return nil
}

// ExpandTilde expands ~ in a path with the home dir
func ExpandTilde(path string) string {
	path = strings.TrimSpace(path)
	components := strings.Split(path, string(os.PathSeparator))
	if len(components) > 0 && components[0] == "~" {
		components[0] = homedir.HomeDir()
		return filepath.Join(components...)
	}
	return path
}

// MergeStringMaps merges the string maps without overwriting previous values
func MergeStringMaps(first map[string]string, additional ...map[string]string) map[string]string {
	if first == nil {
		first = make(map[string]string)
	}
	for _, a := range additional {
		for key, value := range a {
			if _, ok := first[key]; !ok {
				first[key] = value
			}
		}
	}
	return first
}

// TimeString returns the current unix time as a string
func TimeString() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// ChdirApp changes the working directory to the app dir under GOPATH and returns a function to cd back to current working dir
func ChdirApp() (func() error, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, exception.New(err)
	}
	if err := os.Chdir(filepath.Join(env.Env().String("GOPATH"), "src/git.blendlabs.com/blend/deployinator")); err != nil {
		return nil, exception.New(err)
	}
	return func() error {
		return exception.New(os.Chdir(wd))
	}, nil
}

// LogCmdInfo logs the command output
func LogCmdInfo(name string, cmd *exec.Cmd, log *logger.Logger) error {
	var wg sync.WaitGroup
	wg.Add(2)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdoutEvent := logger.Flag(name)
	stderrEvent := logger.Flag(fmt.Sprintf("%s-error", name))
	cmd.Start()
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.SyncTrigger(logger.Messagef(stdoutEvent, line).WithFlagTextColor(logger.ColorGreen))
		}
	}()

	all := ""
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			all = fmt.Sprintf("%s\n%s", all, line)
			log.SyncTrigger(logger.Messagef(stderrEvent, line).WithFlagTextColor(logger.ColorRed))
		}
	}()
	wg.Wait()
	err = cmd.Wait()
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s\n%s", err.Error(), all)
}

// PollForInput prints the prompt then reads in up to delim from stdin
func PollForInput(prompt string, delim byte) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(prompt)
	return reader.ReadString(delim)
}

// PollForPasswordInput does not echo back the input
func PollForPasswordInput(prompt string) (string, error) {
	fmt.Println(prompt)
	bytePwd, err := terminal.ReadPassword(int(syscall.Stdin))
	return strings.TrimSpace(string(bytePwd)), exception.New(err)
}

// PollForConfirmation prints the prompt and then polls for y/n
// if yes then no error occurs, else an error is returned
func PollForConfirmation(prompt string) error {
	p := fmt.Sprintf("%s [y/n]", prompt)
	in, err := PollForInput(p, '\n')
	if err != nil {
		return err
	}
	low := strings.TrimSpace(strings.ToLower(in))
	if low == "y" || low == "yes" {
		return nil
	}
	return exception.New(fmt.Sprintf("Confirmation declined for prompt `%s`", prompt))
}

// PtrSliceToStringSlice creates a string slice from the pointers
func PtrSliceToStringSlice(ptrs []*string) []string {
	ret := []string{}
	for _, ptr := range ptrs {
		if ptr != nil {
			ret = append(ret, *ptr)
		}
	}
	return ret
}

// PtrSliceFromStringSlice creates a slice of *string
func PtrSliceFromStringSlice(inputs []string) (outputs []*string) {
	for _, input := range inputs {
		outputs = append(outputs, ref.String(input))
	}
	return
}

// NotifyAirbrake notifies on the error and logs if there is a problem sending the airbrake
func NotifyAirbrake(err error, env string, logger *logger.Logger) {
	logger.SyncError(err)
	config := airbrake.NewConfigFromEnv()
	airbrake, aerr := airbrake.NewClientFromConfig(config)
	if aerr != nil {
		logger.SyncError(aerr)
		return
	}
	notice := airbrake.Notice(err, nil, 2)
	notice.Context["environment"] = env
	airbrake.SendNotice(notice)
}

// ExceptionUnwrap unwraps an exception.Ex object to gets the underlying error
func ExceptionUnwrap(err error) error {
	if ex, ok := err.(*exception.Ex); ok {
		err = ex.Class()
	}
	return err
}

// NewLoggerFromEnvOrAll returns a new logger from the env or logger.All
func NewLoggerFromEnvOrAll() *logger.Logger {
	log, err := logger.NewFromEnv()
	if err != nil {
		log = logger.All()
		log.Error(err)
	}
	return log
}

// MatchesWildcardDomain checks if the fqdn can be served by a wildcard cert on wildcard
func MatchesWildcardDomain(fqdn string, wildcard string) bool {
	wildcard = "." + strings.ToLower(strings.Trim(wildcard, "."))
	fqdn = strings.ToLower(strings.Trim(fqdn, "."))
	if strings.HasSuffix(fqdn, wildcard) {
		remaining := strings.TrimSuffix(fqdn, wildcard)
		return !strings.Contains(remaining, ".")
	}
	return false
}

// WildcardNeededForDomain returns the wildcard cert needed for the domain
func WildcardNeededForDomain(domain string) string {
	domain = strings.ToLower(strings.Trim(domain, "."))
	tldIdx := strings.Index(domain, ".")
	if tldIdx > 0 {
		return fmt.Sprintf("*.%s", domain[tldIdx+1:])
	}
	return ""
}

// ClearEnv clears os.Environ and returns a function to restore it
func ClearEnv() (restore func()) {
	ev := os.Environ()
	os.Clearenv()
	return func() {
		for _, e := range ev {
			parts := strings.Split(e, "=")
			os.Setenv(parts[0], parts[1])
		}
	}
}
