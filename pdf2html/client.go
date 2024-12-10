package pdf2html

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/nextunit-io/go-tools/tools"
)

type Client struct {
	execClient  tools.ExecInterface
	wrapperFunc func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper
}

type CmdWrapper interface {
	Run() error
}

type Output struct {
	Out      string
	HtmlFile string
	XmlFile  string
}

type Options struct {
	FirstPage     *int     // first page to convert
	LastPage      *int     // last page to convert
	Quite         bool     // don't print any messages or errors
	ExchangeLinks bool     // exchange .pdf links by .html
	ComplexDoc    bool     // generate complex document
	SingleDoc     bool     // generate single document that includes all pages
	IgnoreImages  bool     // ignore images
	NoFrames      bool     // generate no frames
	Stdout        bool     // use standard output
	Zoom          *float32 // zoom the pdf document (default 1.5)
	Xml           bool     // output for XML post-processing
	NoRoundCoord  bool     // do not round coordinates (with XML output only)
	Hidden        bool     // output hidden text
	NoMerge       bool     // do not merge paragraphs
	Enc           *string  // output text encoding name
	Fmt           *string  // image file format for Splash output (png or jpg)
	OwnerPassword *string  // owner password (for encrypted files)
	UserPassword  *string  // user password (for encrypted files)
	NoDrm         bool     // override document DRM settings
	Wbt           *int     // word break threshold (default 10 percent)
	FontFullName  bool     // outputs font full name
}

var wrapCmd func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper = func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper {
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

const (
	client_cli = "pdftohtml"

	versionCheck = ">= 24.11.0, < 25.0"
)

// check if the version of pdftotext is working with this library
func (c Client) checkVersion() error {
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
func (c Client) exec(args ...string) (*string, *string, error) {
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

func cleanup(output *Output) error {
	// Delete XML File if exists
	if _, err := tools.GetOsInstance().Stat(output.XmlFile); err == nil {
		err := tools.GetOsInstance().Remove(output.XmlFile)
		if err != nil {
			return err
		}
	}

	// Delete HTML File if exists
	if _, err := tools.GetOsInstance().Stat(output.HtmlFile); err == nil {
		err = tools.GetOsInstance().Remove(output.HtmlFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) GetXML(filePath string, options Options) (*PdfXmlData, error) {
	dir, err := tools.GetOsInstance().MkdirTemp(tools.GetOsInstance().TempDir(), fmt.Sprintf("%s-*", strings.ReplaceAll(filePath, "/", "_")))

	if err != nil {
		return nil, err
	}

	// Delete the temp directory at the end
	defer func() {
		tools.GetOsInstance().RemoveAll(dir)
	}()

	options.Xml = true
	output, err := c.Get(filePath, dir, options)

	if err != nil {
		return nil, err
	}

	content, err := tools.GetOsInstance().ReadFile(output.XmlFile)
	if err != nil {
		return nil, err
	}

	err = cleanup(output)
	if err != nil {
		return nil, err
	}

	var data PdfXmlData
	err = xml.Unmarshal([]byte(content), &data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (c Client) GetHTML(filePath string, options Options) (*string, error) {
	dir, err := tools.GetOsInstance().MkdirTemp(tools.GetOsInstance().TempDir(), fmt.Sprintf("%s-*", strings.ReplaceAll(filePath, "/", "_")))

	if err != nil {
		return nil, err
	}

	// Delete the temp directory at the end
	defer func() {
		tools.GetOsInstance().RemoveAll(dir)
	}()

	options.Xml = false
	output, err := c.Get(filePath, dir, options)

	if err != nil {
		return nil, err
	}

	content, err := tools.GetOsInstance().ReadFile(output.HtmlFile)
	if err != nil {
		return nil, err
	}

	err = cleanup(output)
	if err != nil {
		return nil, err
	}

	contentString := string(content)

	return &contentString, nil
}

// Get the content for a given file with options
func (c Client) Get(filePath, outputPathPrefix string, options Options) (*Output, error) {
	args := []string{}
	if options.FirstPage != nil {
		args = append(args, "-f", strconv.Itoa(*options.FirstPage))
	}
	if options.LastPage != nil {
		args = append(args, "-l", strconv.Itoa(*options.LastPage))
	}
	if options.Quite {
		args = append(args, "-q")
	}
	if options.ExchangeLinks {
		args = append(args, "-p")
	}
	if options.ComplexDoc {
		args = append(args, "-c")
	}
	if options.SingleDoc {
		args = append(args, "-s")
	}
	if options.IgnoreImages {
		args = append(args, "-i")
	}
	if options.NoFrames {
		args = append(args, "-noframes")
	}
	if options.Stdout {
		args = append(args, "-stdout")
	}
	if options.Zoom != nil {
		args = append(args, "-zoom", fmt.Sprintf("%f", *options.Zoom))
	}
	if options.Xml {
		args = append(args, "-xml")
	}
	if options.NoRoundCoord {
		args = append(args, "-noroundcoord")
	}
	if options.Hidden {
		args = append(args, "-hidden")
	}
	if options.NoMerge {
		args = append(args, "-nomerge")
	}
	if options.Enc != nil {
		args = append(args, "-enc", *options.Enc)
	}
	if options.Fmt != nil {
		args = append(args, "-fmt", *options.Fmt)
	}
	if options.OwnerPassword != nil {
		args = append(args, "-opw", *options.OwnerPassword)
	}
	if options.UserPassword != nil {
		args = append(args, "-upw", *options.UserPassword)
	}
	if options.NoDrm {
		args = append(args, "-nodrm")
	}
	if options.Wbt != nil {
		args = append(args, "-wbt", strconv.Itoa(*options.Wbt))
	}
	if options.FontFullName {
		args = append(args, "-fontfullname")
	}

	htmlPath := fmt.Sprintf("%s.html", outputPathPrefix)
	xmlPath := fmt.Sprintf("%s.xml", outputPathPrefix)
	args = append(args, filePath, outputPathPrefix)

	out, e, err := c.exec(args...)
	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, fmt.Errorf("channel: %s", *e)
	}

	return &Output{
		Out:      *out,
		HtmlFile: htmlPath,
		XmlFile:  xmlPath,
	}, nil
}

// Get the current pdftotext version
func (c Client) GetVersion() (*string, error) {
	_, out, err := c.exec("-v")

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("cannot find the version")
	}

	r := regexp.MustCompile("pdftohtml version ([^\n]+)\n")
	matches := r.FindStringSubmatch(*out)
	if len(matches) != 2 {
		return nil, fmt.Errorf("cannot find the version")
	}

	return &matches[1], nil
}

// Overwrite wrapper function for buffer
func SetWrapperFunc(fn func(cmd *exec.Cmd, stdout, stderr io.Writer) CmdWrapper) {
	wrapCmd = fn
}

// Get the pdftotext client
// Will return an error, if the installed CLI version is not valid
func NewClient() (*Client, error) {
	c := &Client{
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
