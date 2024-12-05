package pdf2text_test

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"

	gomock "github.com/nextunit-io/go-mock"
	"github.com/nextunit-io/go-pdf2X/pdf2text"
	"github.com/nextunit-io/go-tools/tools"
	"github.com/nextunit-io/go-tools/toolsmock"
	"github.com/stretchr/testify/assert"
)

var (
	execMock      *toolsmock.ExecMock
	wrapperFnMock *gomock.ToolMock[
		struct {
			Cmd    *exec.Cmd
			Stdout io.Writer
			Stderr io.Writer
		},
		pdf2text.CmdWrapper,
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

	// Setup wrapper function mock
	wrapperFnMock = gomock.GetMock[
		struct {
			Cmd    *exec.Cmd
			Stdout io.Writer
			Stderr io.Writer
		},
		pdf2text.CmdWrapper,
	](fmt.Errorf("WRAPPER general error"))

	runMock = gomock.GetMock[interface{}, func(cmd []string) (*string, *string, error)](fmt.Errorf("GENERAL ERROR"))

	pdf2text.SetWrapperFunc(func(cmd *exec.Cmd, stdout, stderr io.Writer) pdf2text.CmdWrapper {
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

	wrapperFnMock.SetAlwaysReturnFn(func() (*pdf2text.CmdWrapper, error) {
		var wrapper pdf2text.CmdWrapper = &testVersionWrapper{}
		return &wrapper, nil
	})

	setupInitialVersion()
}

func setupInitialVersion() {
	runMock.Reset()
	fn := func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftotext version 24.11.0
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}

	runMock.AddReturnValue(&fn)
}

func TestGetClient(t *testing.T) {
	t.Helper()
	setupTests()

	client, err := pdf2text.NewClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2text.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "cannot check version of pdftotext", err.Error())
	assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Too low version
	fn := func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftotext version 24.10.100
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}
	runMock.AddReturnValue(&fn)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2text.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "version 24.10.100 does not pass the version constraint >= 24.11.0, < 25.0", err.Error())
	assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)

	// Too high version
	fn = func(cmd []string) (*string, *string, error) {
		versionReturnValue := `pdftotext version 25.0.0
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1996-2011, 2022 Glyph & Cog, LLC`
		return nil, &versionReturnValue, nil
	}
	runMock.AddReturnValue(&fn)

	// Second try should fail, because there will be no version sent back upon the second time
	client, err = pdf2text.NewClient()
	assert.Nil(t, client)
	assert.Equal(t, "version 25.0.0 does not pass the version constraint >= 24.11.0, < 25.0", err.Error())
	assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
}

func TestGetVersion(t *testing.T) {
	t.Helper()
	setupTests()
	client, _ := pdf2text.NewClient()

	versions := []string{"24.11.0", "1.0", "2.5", "100.2.4", "50.0.4-meta"}

	for _, version := range versions {
		t.Run(fmt.Sprintf("Check for version %s", version), func(t *testing.T) {
			runMock.Reset()

			fn := func(cmd []string) (*string, *string, error) {
				versionReturnValue := fmt.Sprintf(`pdftotext version %s
Copyright 2005-2024 The Poppler Developers - http://poppler.freedesktop.org
Copyright 1996-2011, 2022 Glyph & Cog, LLC`, version)
				return nil, &versionReturnValue, nil
			}

			runMock.AddReturnValue(&fn)
			v, err := client.GetVersion()

			assert.Nil(t, err)
			assert.Equal(t, version, *v)
			assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
		})
	}

	t.Run("Version errors", func(t *testing.T) {
		runMock.Reset()

		v, err := client.GetVersion()

		assert.Nil(t, v)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
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
		assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
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
		assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Error on execute", func(t *testing.T) {
		runMock.Reset()

		v, err := client.GetVersion()
		assert.Nil(t, v)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftotext", "-v"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})
}

func TestGetEncodings(t *testing.T) {
	t.Helper()
	setupTests()
	client, _ := pdf2text.NewClient()

	encs := []struct {
		Output         string
		ExpectedReturn []string
	}{
		{
			Output:         "",
			ExpectedReturn: []string{},
		},
		{
			Output:         "ASCII\nBig5\nGBK",
			ExpectedReturn: []string{"ASCII", "Big5", "GBK"},
		},
		{
			Output:         "ISO-8859-6\nISO-8859-8\nShift-JIS",
			ExpectedReturn: []string{"ISO-8859-6", "ISO-8859-8", "Shift-JIS"},
		},
	}

	for _, enc := range encs {
		t.Run(fmt.Sprintf("Check for encodings %s", strings.Join(enc.ExpectedReturn, ", ")), func(t *testing.T) {
			runMock.Reset()

			fn := func(cmd []string) (*string, *string, error) {
				encReturnValue := fmt.Sprintf("Available encodings are:\n%s", enc.Output)
				return &encReturnValue, nil, nil
			}

			runMock.AddReturnValue(&fn)
			e, err := client.GetEncodings()

			assert.Nil(t, err)
			assert.Equal(t, enc.ExpectedReturn, e)
			assert.Equal(t, []string{"pdftotext", "-listenc"}, wrapperFnMock.GetLastInput().Cmd.Args)
		})
	}

	t.Run("Encodings parse error", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			encReturnValue := "invalidValue"
			return &encReturnValue, nil, nil
		}

		runMock.AddReturnValue(&fn)
		e, err := client.GetEncodings()

		assert.Equal(t, []string{}, e)
		assert.Nil(t, err)
		assert.Equal(t, []string{"pdftotext", "-listenc"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Encodings parse error (empty output)", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			encReturnValue := ""
			return &encReturnValue, nil, nil
		}

		runMock.AddReturnValue(&fn)
		e, err := client.GetEncodings()

		assert.Equal(t, []string{}, e)
		assert.Equal(t, "no valid output given", err.Error())
		assert.Equal(t, []string{"pdftotext", "-listenc"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Wrong output channel", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			encReturnValue := "invalidValue"
			return nil, &encReturnValue, nil
		}

		runMock.AddReturnValue(&fn)
		e, err := client.GetEncodings()

		assert.Equal(t, []string{}, e)
		assert.Equal(t, "no valid output given", err.Error())
		assert.Equal(t, []string{"pdftotext", "-listenc"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Error on execute", func(t *testing.T) {
		runMock.Reset()

		e, err := client.GetEncodings()

		assert.Equal(t, []string{}, e)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftotext", "-listenc"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})
}

func TestGet(t *testing.T) {
	t.Helper()
	setupTests()
	client, _ := pdf2text.NewClient()

	outputs := []string{}

	for i, output := range outputs {
		t.Run(fmt.Sprintf("Check for get outputs in loop %d", i), func(t *testing.T) {
			runMock.Reset()

			fn := func(cmd []string) (*string, *string, error) {
				return &output, nil, nil
			}

			runMock.AddReturnValue(&fn)
			o, err := client.Get("filename", pdf2text.Options{})

			assert.Nil(t, err)
			assert.Equal(t, output, *o)
			assert.Equal(t, []string{"pdftotext", "filename", "-"}, wrapperFnMock.GetLastInput().Cmd.Args)
		})
	}

	t.Run("Check for all flags", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			output := "test"
			return &output, nil, nil
		}

		runMock.AddReturnValue(&fn)
		_, err := client.Get("filename", pdf2text.Options{
			FirstPage:     pointerHelperFn(20),
			LastPage:      pointerHelperFn(40),
			Resolution:    pointerHelperFn(60),
			X:             pointerHelperFn(80),
			Y:             pointerHelperFn(100),
			Width:         pointerHelperFn(120),
			Height:        pointerHelperFn(140),
			Layout:        true,
			Fixed:         pointerHelperFn("test-fixed"),
			Raw:           true,
			NoDiag:        true,
			HtmlMeta:      true,
			Tsv:           true,
			Enc:           pointerHelperFn("test-enc"),
			Eol:           pointerHelperFn("test-eol"),
			Nopgbrk:       true,
			Bbox:          true,
			BboxLayout:    true,
			CropBox:       true,
			ColSpacing:    pointerHelperFn(float32(32.5)),
			OwnerPassword: pointerHelperFn("test-owner-password"),
			UserPassword:  pointerHelperFn("test-user-password"),
		})

		assert.Nil(t, err)
		assert.Equal(t, []string{"pdftotext",
			"-f", "20",
			"-l", "40",
			"-r", "60",
			"-x", "80",
			"-y", "100",
			"-W", "120",
			"-H", "140",
			"-layout",
			"-fixed", "test-fixed",
			"-raw",
			"-nodiag",
			"-htmlmeta",
			"-tsv",
			"-enc", "test-enc",
			"-eol", "test-eol",
			"-nopgbrk",
			"-bbox",
			"-bbox-layout",
			"-cropbox",
			"-colspacing", "32.500000",
			"-opw", "test-owner-password",
			"-upw", "test-user-password",
			"filename",
			"-",
		}, wrapperFnMock.GetLastInput().Cmd.Args)

	})

	t.Run("Wrong output channel", func(t *testing.T) {
		runMock.Reset()

		fn := func(cmd []string) (*string, *string, error) {
			returnValue := "invalidValue"
			return nil, &returnValue, nil
		}

		runMock.AddReturnValue(&fn)
		o, err := client.Get("filename", pdf2text.Options{})

		assert.Nil(t, o)
		assert.Equal(t, "channel: invalidValue", err.Error())
		assert.Equal(t, []string{"pdftotext", "filename", "-"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})

	t.Run("Error on execute", func(t *testing.T) {
		runMock.Reset()

		o, err := client.Get("filename", pdf2text.Options{})

		assert.Nil(t, o)
		assert.Equal(t, "GENERAL ERROR", err.Error())
		assert.Equal(t, []string{"pdftotext", "filename", "-"}, wrapperFnMock.GetLastInput().Cmd.Args)
	})
}
