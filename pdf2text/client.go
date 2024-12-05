package pdf2text

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/nextunit-io/go-tools/tools"
)

type client struct {
	execClient  tools.ExecInterface
	wrapperFunc func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper
}

type CmdWrapper interface {
	Run() error
}

type Options struct {
	FirstPage     *int     // first page to convert
	LastPage      *int     // last page to convert
	Resolution    *int     // resolution, in DPI (default is 72)
	X             *int     // x-coordinate of the crop area top left corner
	Y             *int     // y-coordinate of the crop area top left corner
	Width         *int     // width of crop area in pixels (default is 0)
	Height        *int     // height of crop area in pixels (default is 0)
	Layout        bool     // maintain original physical layout
	Fixed         *string  // assume fixed-pitch (or tabular) text
	Raw           bool     // keep strings in content stream order
	NoDiag        bool     // discard diagonal text
	HtmlMeta      bool     // generate a simple HTML file, including the meta information
	Tsv           bool     // generate a simple TSV file, including the meta information for bounding boxes
	Enc           *string  // output text encoding name
	Eol           *string  // output end-of-line convention (unix, dos, or mac)
	Nopgbrk       bool     // don't insert page breaks between pages
	Bbox          bool     // output bounding box for each word and page size to html. Sets -htmlmeta
	BboxLayout    bool     // like -bbox but with extra layout bounding box data.  Sets -htmlmeta
	CropBox       bool     // use the crop box rather than media box
	ColSpacing    *float32 // how much spacing we allow after a word before considering adjacent text to be a new column, as a fraction of the font size (default is 0.7, old releases had a 0.3 default)
	OwnerPassword *string  // owner password (for encrypted files)
	UserPassword  *string  // user password (for encrypted files)
}

var wrapCmd func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper = func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper {
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

const (
	client_cli = "pdftotext"

	versionCheck = ">= 24.11.0, < 25.0"
)

// check if the version of pdftotext is working with this library
func (c client) checkVersion() error {
	v, err := c.GetVersion()
	if err != nil {
		return fmt.Errorf("cannot check version of %s", client_cli)
	}

	versionObj, err := version.NewVersion(*v)
	if err != nil {
		return err
	}

	versionConstraint, err := version.NewConstraint(versionCheck)
	if err != nil {
		return err
	}

	if !versionConstraint.Check(versionObj) {
		return fmt.Errorf("version %s does not pass the version constraint %s", *v, versionCheck)
	}

	return nil
}

// Execute function. Some outputs are using the stdin, some the stderr.
// Therefore the three return values are representating stdout, stderr, error
func (c client) exec(args ...string) (*string, *string, error) {
	cmd := tools.GetExecInstance().Command(client_cli, args...)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	wrappedCmd := c.wrapperFunc(cmd, &outBuffer, &errBuffer)

	err := wrappedCmd.Run()

	if err != nil {
		return nil, nil, err
	}

	outputString := outBuffer.String()
	errorString := errBuffer.String()

	var outputStrPtr *string = nil
	var errorStrPtr *string = nil

	if outputString != "" {
		outputStrPtr = &outputString
	}
	if errorString != "" {
		errorStrPtr = &errorString
	}

	return outputStrPtr, errorStrPtr, nil
}

// Get the content for a given file with options
func (c client) Get(filePath string, options Options) (*string, error) {
	args := []string{}
	if options.FirstPage != nil {
		args = append(args, "-f", strconv.Itoa(*options.FirstPage))
	}
	if options.LastPage != nil {
		args = append(args, "-l", strconv.Itoa(*options.LastPage))
	}
	if options.Resolution != nil {
		args = append(args, "-r", strconv.Itoa(*options.Resolution))
	}
	if options.X != nil {
		args = append(args, "-x", strconv.Itoa(*options.X))
	}
	if options.Y != nil {
		args = append(args, "-y", strconv.Itoa(*options.Y))
	}
	if options.Width != nil {
		args = append(args, "-W", strconv.Itoa(*options.Width))
	}
	if options.Height != nil {
		args = append(args, "-H", strconv.Itoa(*options.Height))
	}
	if options.Layout {
		args = append(args, "-layout")
	}
	if options.Fixed != nil {
		args = append(args, "-fixed", *options.Fixed)
	}
	if options.Raw {
		args = append(args, "-raw")
	}
	if options.NoDiag {
		args = append(args, "-nodiag")
	}
	if options.HtmlMeta {
		args = append(args, "-htmlmeta")
	}
	if options.Tsv {
		args = append(args, "-tsv")
	}
	if options.Enc != nil {
		args = append(args, "-enc", *options.Enc)
	}
	if options.Eol != nil {
		args = append(args, "-eol", *options.Eol)
	}
	if options.Nopgbrk {
		args = append(args, "-nopgbrk")
	}
	if options.Bbox {
		args = append(args, "-bbox")
	}
	if options.BboxLayout {
		args = append(args, "-bbox-layout")
	}
	if options.CropBox {
		args = append(args, "-cropbox")
	}
	if options.ColSpacing != nil {
		args = append(args, "-colspacing", fmt.Sprintf("%f", *options.ColSpacing))
	}
	if options.OwnerPassword != nil {
		args = append(args, "-opw", *options.OwnerPassword)
	}
	if options.UserPassword != nil {
		args = append(args, "-upw", *options.UserPassword)
	}

	args = append(args, filePath, "-")

	out, e, err := c.exec(args...)
	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, fmt.Errorf("channel: %s", *e)
	}

	return out, nil
}

// Get the current pdftotext version
func (c client) GetVersion() (*string, error) {
	_, out, err := c.exec("-v")

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("cannot find the version")
	}

	r := regexp.MustCompile("pdftotext version ([^\n]+)\n")
	matches := r.FindStringSubmatch(*out)
	if len(matches) != 2 {
		return nil, fmt.Errorf("cannot find the version")
	}

	return &matches[1], nil
}

// Get the available encodings of pdftotext
func (c client) GetEncodings() ([]string, error) {
	out, _, err := c.exec("-listenc")

	if err != nil {
		return []string{}, err
	}
	if out == nil {
		return []string{}, fmt.Errorf("no valid output given")
	}

	lines := strings.Split(strings.TrimSpace(*out), "\n")

	return lines[1:], nil
}

// Overwrite wrapper function for buffer
func SetWrapperFunc(fn func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper) {
	wrapCmd = fn
}

// Get the pdftotext client
// Will return an error, if the installed CLI version is not valid
func NewClient() (*client, error) {
	c := &client{
		execClient:  tools.GetExecInstance(),
		wrapperFunc: wrapCmd,
	}

	// Check for valid CLI version before
	// Do not do the check for getting the version
	err := c.checkVersion()
	if err != nil {
		return nil, err
	}

	return c, err
}
