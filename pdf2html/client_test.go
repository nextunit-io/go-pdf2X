package pdf2html_test

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"testing"

	gomock "github.com/nextunit-io/go-mock"
	"github.com/nextunit-io/go-pdf2X/pdf2html"
	"github.com/nextunit-io/go-tools/tools"
	"github.com/nextunit-io/go-tools/toolsmock"
	"github.com/stretchr/testify/assert"
)

var (
	osMock       *toolsmock.OsMock
	execMock     *toolsmock.ExecMock
	fileInfoMock *toolsmock.FileInfoMock

	wrapperFnMock *gomock.ToolMock[
		struct {
			Cmd    *exec.Cmd
			Stdout io.Writer
			Stderr io.Writer
		},
		pdf2html.CmdWrapper,
	]
	runMock *gomock.ToolMock[
		interface{},
		func(cmd []string) (*string, *string, error),
	]

	versionWrapperMock *gomock.ToolMock[
		interface{},
		string,
	]
)

var xmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE pdf2xml SYSTEM "pdf2xml.dtd">

<pdf2xml producer="poppler" version="24.11.0">
<page number="1" position="absolute" top="0" left="0" height="1262" width="892">
	<fontspec id="0" size="8" family="ArialMT" color="#000000"/>
	<fontspec id="1" size="9" family="Arial" color="#AAAAAA"/>
	<text top="452" left="106" width="227" height="19" font="0"><b>Test Bold</b></text>
	<text top="471" left="106" width="188" height="19" font="0"><b>Test</b> mixed</text>
	<text top="272" left="106" width="61" height="19" font="0">Only text</text>
</page>
<page number="2" position="absolute" top="0" left="0" height="1264" width="894">
	<text top="452" left="106" width="227" height="19" font="0"><b>Test Bold</b> p2</text>
	<text top="471" left="106" width="188" height="19" font="0"><b>Test</b> mixed p2</text>
	<text top="272" left="106" width="61" height="19" font="0">Only text p2</text>
</page>
</pdf2xml>`

var expectedXMLObj = pdf2html.PdfXmlData{
	XMLName:  xml.Name{Local: "pdf2xml"},
	Producer: pointerHelperFn("poppler"),
	Version:  pointerHelperFn("24.11.0"),
	Pages: []pdf2html.PdfXmlPage{
		{
			XMLName: xml.Name{
				Local: "page",
			},
			PageNumber: pointerHelperFn(1),
			Position:   pointerHelperFn("absolute"),
			Top:        pointerHelperFn(0),
			Left:       pointerHelperFn(0),
			Width:      pointerHelperFn(892),
			Height:     pointerHelperFn(1262),
			FontSpecs: []pdf2html.PdfXmlFontSpec{
				{
					XMLName: xml.Name{
						Local: "fontspec",
					},
					ID:     pointerHelperFn(0),
					Size:   pointerHelperFn(8),
					Family: pointerHelperFn("ArialMT"),
					Color:  pointerHelperFn("#000000"),
				},
				{
					XMLName: xml.Name{
						Local: "fontspec",
					},
					ID:     pointerHelperFn(1),
					Size:   pointerHelperFn(9),
					Family: pointerHelperFn("Arial"),
					Color:  pointerHelperFn("#AAAAAA"),
				},
			},
			Texts: []pdf2html.PdfXmlText{
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(452),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(227),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn(""),
					BoldText: pointerHelperFn("Test Bold"),
				},
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(471),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(188),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn(" mixed"),
					BoldText: pointerHelperFn("Test"),
				},
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(272),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(61),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn("Only text"),
					BoldText: nil,
				},
			},
		},
		{
			XMLName: xml.Name{
				Local: "page",
			},
			PageNumber: pointerHelperFn(2),
			Position:   pointerHelperFn("absolute"),
			Top:        pointerHelperFn(0),
			Left:       pointerHelperFn(0),
			Width:      pointerHelperFn(894),
			Height:     pointerHelperFn(1264),

			Texts: []pdf2html.PdfXmlText{
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(452),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(227),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn(" p2"),
					BoldText: pointerHelperFn("Test Bold"),
				},
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(471),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(188),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn(" mixed p2"),
					BoldText: pointerHelperFn("Test"),
				},
				{
					XMLName: xml.Name{
						Local: "text",
					},
					Top:      pointerHelperFn(272),
					Left:     pointerHelperFn(106),
					Width:    pointerHelperFn(61),
					Height:   pointerHelperFn(19),
					Text:     pointerHelperFn("Only text p2"),
					BoldText: nil,
				},
			},
		},
	},
}

type testVersionWrapper struct{}

func (testVersionWrapper) Run() error {
	runMock.AddInput(nil)

	result, err := runMock.GetNextResult()
	if err != nil {
		return err
	}
	fn := *result

	cmdInput := wrapperFnMock.GetLastInput()
	outString, errString, err := fn(cmdInput.Cmd.Args)

	if outString != nil {
		cmdInput.Stdout.Write([]byte(*outString))
	}
	if errString != nil {
		cmdInput.Stderr.Write([]byte(*errString))
	}

	return err
}

func pointerHelperFn[T any](x T) *T {
	return &x
}

func setupTests() {
	execMock = toolsmock.GetExecMock()
	tools.SetExecInstance(execMock)

	osMock = toolsmock.GetOsMock()
	tools.SetOsInstance(osMock)

	fileInfoMock = toolsmock.GetFileInfoMock()

	// Setup wrapper function mock
	wrapperFnMock = gomock.GetMock[
		struct {
			Cmd    *exec.Cmd
			Stdout io.Writer
			Stderr io.Writer
		},
		pdf2html.CmdWrapper,
	](fmt.Errorf("WRAPPER general error"))

	runMock = gomock.GetMock[interface{}, func(cmd []string) (*string, *string, error)](fmt.Errorf("GENERAL ERROR"))

	pdf2html.SetWrapperFunc(func(cmd *exec.Cmd, stdout, stderr io.Writer) pdf2html.CmdWrapper {
		wrapperFnMock.AddInput(struct {
			Cmd    *exec.Cmd
			Stdout io.Writer
			Stderr io.Writer
		}{
			Cmd:    cmd,
			Stdout: stdout,
			Stderr: stderr,
		})

		result, err := wrapperFnMock.GetNextResult()
		if err != nil {
			panic(err.Error())
		}

		return *result
	})

	versionWrapperMock = gomock.GetMock[interface{}, string](fmt.Errorf("VERSIONWRAPPER general error"))

	execMock.Mock.Command.SetAlwaysReturnFn(func() (**exec.Cmd, error) {
		lastCommand := execMock.Mock.Command.GetLastInput()
		cmd := exec.Command(lastCommand.Name, lastCommand.Arg...)
		return &cmd, nil
	})

	wrapperFnMock.SetAlwaysReturnFn(func() (*pdf2html.CmdWrapper, error) {
		var wrapper pdf2html.CmdWrapper = &testVersionWrapper{}
		return &wrapper, nil
	})

	osMock.Mock.TempDir.SetAlwaysReturn("test-tmp-dir")
	osMock.Mock.MkdirTemp.SetAlwaysReturn("test-mkdir-tmpdir")
	osMock.Mock.RemoveAll.AddReturnValue(pointerHelperFn(false))
	osMock.Mock.ReadFile.SetAlwaysReturn([]byte("test-read-file"))

	var fileInfo fs.FileInfo = fileInfoMock
	osMock.Mock.Stat.AddReturnValue(&fileInfo)
	osMock.Mock.Stat.AddReturnValue(&fileInfo)

	osMock.Mock.Remove.SetAlwaysReturn(false)

	setupInitialVersion()
}

func setupXmlTests() {
	setupTests()
	osMock.Mock.ReadFile.Reset()
	osMock.Mock.ReadFile.AddReturnValue(pointerHelperFn([]byte(xmlContent)))
}

func setupInitialVersion() {
	runMock.Reset()
	fn := func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftohtml version 24.11.0
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1999-2003 Gueorgui Ovtcharov and Rainer Dorsch
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}

	runMock.AddReturnValue(&fn)
}

func TestGetClient(t *testing.T) {
	t.Helper()
	setupTests()

	client, err := pdf2html.NewClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2html.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "cannot check version of pdftohtml", err.Error())
	assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Too low version
	fn := func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftohtml version 24.10.1000
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1999-2003 Gueorgui Ovtcharov and Rainer Dorsch
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}
	runMock.AddReturnValue(&fn)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2html.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "version 24.10.1000 does not pass the version constraint >= 24.11.0, < 25.0", err.Error())
	assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Too high version
	fn = func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftohtml version 25.0.0
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1999-2003 Gueorgui Ovtcharov and Rainer Dorsch
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}
	runMock.AddReturnValue(&fn)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2html.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "version 25.0.0 does not pass the version constraint >= 24.11.0, < 25.0", err.Error())
	assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
}

func TestGetVersion(t *testing.T) {
	t.Helper()
	setupTests()
	client, _ := pdf2html.NewClient()

	versions := []string{"24.11.0", "1.0", "2.5", "100.2.4", "50.0.4-meta"}

	for _, version := range versions {
		t.Run(fmt.Sprintf("Check for version %s", version), func(t *testing.T) {
			runMock.Reset()

			fn := func(cmd []string) (*string, *string, error) {
				versionReturnValue := fmt.Sprintf(`pdftohtml version %s
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1999-2003 Gueorgui Ovtcharov and Rainer Dorsch
Copyright 1996-2011, 2022 Glyph & Cog, LLC`, version)
				return nil, &versionReturnValue, nil
			}

			runMock.AddReturnValue(&fn)
			v, err := client.GetVersion()

			assert.Nil(t, err)
			assert.Equal(t, version, *v)
			assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
		})
	}

	t.Run("Version errors", func(t *testing.T) {
		runMock.Reset()

		v, err := client.GetVersion()

		assert.Nil(t, v)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Version parse error", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			versionReturnValue := "invalidValue"
			return nil, &versionReturnValue, nil
		}

		runMock.AddReturnValue(&fn)
		v, err := client.GetVersion()
		assert.Nil(t, v)
		assert.Equal(t, "cannot find the version", err.Error())
		assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Wrong output channel", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			versionReturnValue := "invalidValue"
			return &versionReturnValue, nil, nil
		}

		runMock.AddReturnValue(&fn)
		v, err := client.GetVersion()
		assert.Nil(t, v)
		assert.Equal(t, "cannot find the version", err.Error())
		assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Error on execute", func(t *testing.T) {
		runMock.Reset()

		v, err := client.GetVersion()
		assert.Nil(t, v)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftohtml", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})
}

func TestGet(t *testing.T) {
	t.Helper()
	setupTests()
	client, _ := pdf2html.NewClient()

	outputs := []string{"test-output", "test-output-2"}

	for i, output := range outputs {
		t.Run(fmt.Sprintf("Check for get outputs in loop %d", i), func(t *testing.T) {
			runMock.Reset()

			fn := func(cmd []string) (*string, *string, error) {
				return &output, nil, nil
			}

			runMock.AddReturnValue(&fn)
			o, err := client.Get("filename", "test-output-path", pdf2html.Options{})

			assert.Nil(t, err)
			assert.Equal(t, output, o.Out)
			assert.Equal(t, "test-output-path.html", o.HtmlFile)
			assert.Equal(t, "test-output-path.xml", o.XmlFile)
			assert.Equal(t, []string{"pdftohtml", "filename", "test-output-path"}, wrapperFnMock.GetLastInput().Cmd.Args)
		})
	}

	t.Run("Check for all flags", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			output := "test"
			return &output, nil, nil
		}

		runMock.AddReturnValue(&fn)
		_, err := client.Get("filename", "test-output-path", pdf2html.Options{
			FirstPage:     pointerHelperFn(20),
			LastPage:      pointerHelperFn(40),
			Quite:         true,
			ExchangeLinks: true,
			ComplexDoc:    true,
			SingleDoc:     true,
			IgnoreImages:  true,
			NoFrames:      true,
			Stdout:        true,
			Zoom:          pointerHelperFn(float32(27.2)),
			Xml:           true,
			NoRoundCoord:  true,
			Hidden:        true,
			NoMerge:       true,
			Enc:           pointerHelperFn("test-enc-string"),
			Fmt:           pointerHelperFn("test-fmt-string"),
			OwnerPassword: pointerHelperFn("test-owner-string"),
			UserPassword:  pointerHelperFn("test-userpassword-string"),
			NoDrm:         true,
			Wbt:           pointerHelperFn(123),
			FontFullName:  true,
		})

		assert.Nil(t, err)
		assert.Equal(t, []string{
			"pdftohtml",
			"-f", "20",
			"-l", "40",
			"-q",
			"-p",
			"-c",
			"-s",
			"-i",
			"-noframes",
			"-stdout",
			"-zoom", "27.200001",
			"-xml",
			"-noroundcoord",
			"-hidden",
			"-nomerge",
			"-enc", "test-enc-string",
			"-fmt", "test-fmt-string",
			"-opw", "test-owner-string",
			"-upw", "test-userpassword-string",
			"-nodrm",
			"-wbt", "123",
			"-fontfullname",
			"filename",
			"test-output-path",
		}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Wrong output channel", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			returnValue := "invalidValue"
			return nil, &returnValue, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.Get("filename", "test-output-path", pdf2html.Options{})

		assert.Nil(t, o)
		assert.Equal(t, "channel: invalidValue", err.Error())
		assert.Equal(t, []string{"pdftohtml", "filename", "test-output-path"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Error on execute", func(t *testing.T) {
		runMock.Reset()

		o, err := client.Get("filename", "test-output-path", pdf2html.Options{})

		assert.Nil(t, o)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftohtml", "filename", "test-output-path"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})
}

func TestGetXml(t *testing.T) {
	t.Helper()

	t.Run("Check for successful GetXML", func(t *testing.T) {
		setupXmlTests()

		client, _ := pdf2html.NewClient()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})

		assert.Nil(t, err)
		assert.Equal(t, expectedXMLObj, *o)
		assert.Equal(t, []string{"pdftohtml", "-xml", "filename", "test-mkdir-tmpdir"}, wrapperFnMock.GetLastInput().Cmd.Args)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, "test-tmp-dir", osMock.Mock.MkdirTemp.GetInput(0).Dir)
		assert.Equal(t, "filename-*", osMock.Mock.MkdirTemp.GetInput(0).Pattern)

		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.ReadFile.GetInput(0).Name)

		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Stat.GetInput(0).Name)
		assert.Equal(t, "test-mkdir-tmpdir.html", osMock.Mock.Stat.GetInput(1).Name)

		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Remove.GetInput(0).Name)
		assert.Equal(t, "test-mkdir-tmpdir.html", osMock.Mock.Remove.GetInput(1).Name)

		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir", osMock.Mock.RemoveAll.GetInput(0).Path)
	})

	t.Run("Check cleanup cannot find HTML files", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Stat.Reset()

		var fileInfo fs.FileInfo = fileInfoMock
		osMock.Mock.Stat.AddReturnValue(&fileInfo)

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, expectedXMLObj, *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())

		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Remove.GetInput(0).Name)
	})

	t.Run("Check cleanup cannot find files at all", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Stat.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, expectedXMLObj, *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Check removeall should not let the process fail", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.RemoveAll.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, expectedXMLObj, *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Cleanup failes (XML)", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Remove.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Equal(t, "Remove general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Cleanup failes (HTML)", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Remove.Reset()
		osMock.Mock.Remove.AddReturnValue(pointerHelperFn(false))

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Equal(t, "Remove general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Readfile failes", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.ReadFile.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Equal(t, "ReadFile general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Get fails", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.ReadFile.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return nil, nil, fmt.Errorf("GET error")
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Equal(t, "GET error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Error on tmp dir", func(t *testing.T) {
		setupXmlTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.MkdirTemp.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetXML("filename", pdf2html.Options{})
		assert.Equal(t, "MkdirTemp general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.RemoveAll.HasBeenCalled())
	})
}

func TestGetHTML(t *testing.T) {
	t.Helper()

	t.Run("Check for successful GetHTML", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		// it is going to remove the XML, since we want HTML
		o, err := client.GetHTML("filename", pdf2html.Options{
			Xml: true,
		})

		assert.Nil(t, err)
		assert.Equal(t, "test-read-file", *o)
		assert.Equal(t, []string{"pdftohtml", "filename", "test-mkdir-tmpdir"}, wrapperFnMock.GetLastInput().Cmd.Args)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, "test-tmp-dir", osMock.Mock.MkdirTemp.GetInput(0).Dir)
		assert.Equal(t, "filename-*", osMock.Mock.MkdirTemp.GetInput(0).Pattern)

		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.html", osMock.Mock.ReadFile.GetInput(0).Name)

		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Stat.GetInput(0).Name)
		assert.Equal(t, "test-mkdir-tmpdir.html", osMock.Mock.Stat.GetInput(1).Name)

		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Remove.GetInput(0).Name)
		assert.Equal(t, "test-mkdir-tmpdir.html", osMock.Mock.Remove.GetInput(1).Name)

		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
		assert.Equal(t, "test-mkdir-tmpdir", osMock.Mock.RemoveAll.GetInput(0).Path)
	})

	t.Run("Check cleanup cannot find HTML files", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Stat.Reset()

		var fileInfo fs.FileInfo = fileInfoMock
		osMock.Mock.Stat.AddReturnValue(&fileInfo)

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, "test-read-file", *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())

		assert.Equal(t, "test-mkdir-tmpdir.xml", osMock.Mock.Remove.GetInput(0).Name)
	})

	t.Run("Check cleanup cannot find files at all", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Stat.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, "test-read-file", *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Check removeall should not let the process fail", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.RemoveAll.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Nil(t, err)
		assert.Equal(t, "test-read-file", *o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Cleanup failes (XML)", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Remove.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Equal(t, "Remove general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Cleanup failes (HTML)", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.Remove.Reset()
		osMock.Mock.Remove.AddReturnValue(pointerHelperFn(false))

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Equal(t, "Remove general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 2, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Readfile failes", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.ReadFile.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Equal(t, "ReadFile general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Get fails", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.ReadFile.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return nil, nil, fmt.Errorf("GET error")
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Equal(t, "GET error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.RemoveAll.HasBeenCalled())
	})

	t.Run("Error on tmp dir", func(t *testing.T) {
		setupTests()
		client, _ := pdf2html.NewClient()

		osMock.Mock.MkdirTemp.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			return pointerHelperFn("test-output"), nil, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.GetHTML("filename", pdf2html.Options{})
		assert.Equal(t, "MkdirTemp general error", err.Error())
		assert.Nil(t, o)

		assert.Equal(t, 1, osMock.Mock.TempDir.HasBeenCalled())
		assert.Equal(t, 1, osMock.Mock.MkdirTemp.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.ReadFile.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Stat.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.Remove.HasBeenCalled())
		assert.Equal(t, 0, osMock.Mock.RemoveAll.HasBeenCalled())
	})
}
