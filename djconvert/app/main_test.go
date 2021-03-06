package app

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	tests := []struct {
		Args           []string
		ExpectError    bool
		ExpectedOutput map[string]string
	}{
		// No arguments
		{
			[]string{},
			true,
			map[string]string{},
		},
		// Bad arguments
		{
			[]string{"foo"},
			true,
			map[string]string{},
		},
		// Insufficient arguments
		{
			[]string{"up"},
			true,
			map[string]string{},
		},
		// Bad file locations
		{
			[]string{"up", "{{ .Dir }}", "."},
			true,
			map[string]string{},
		},
		// Up not pretty
		{
			[]string{"up",
				"{{ .Dir }}/input_hello_world.json",
				"{{ .Dir }}/doc.json",
			},
			false,
			map[string]string{"doc.json": `{"events":{"aa582b4df04ba01af5205e702d4d16ed0b2c0705":{"parent":"","handler":"SET","args":{"path":[],"value":{"hello":"world"}}}},"timestamps":["aa582b4df04ba01af5205e702d4d16ed0b2c0705"]}` + "\n"},
		},
		// Up pretty
		{
			[]string{"up",
				"{{ .Dir }}/input_hello_world.json",
				"{{ .Dir }}/doc.json",
				"--pretty",
			},
			false,
			map[string]string{"doc.json": `{
    "events": {
        "aa582b4df04ba01af5205e702d4d16ed0b2c0705": {
            "parent": "",
            "handler": "SET",
            "args": {
                "path": [],
                "value": {
                    "hello": "world"
                }
            }
        }
    },
    "timestamps": [
        "aa582b4df04ba01af5205e702d4d16ed0b2c0705"
    ]
}` + "\n"},
		},
		// Down not pretty
		{
			[]string{"down",
				"{{ .Dir }}/doc_hello_world.json",
				"{{ .Dir }}/static.json",
				"aa58",
			},
			false,
			map[string]string{"static.json": `{"hello":"world"}` + "\n"},
		},
		// Down pretty
		{
			[]string{"down",
				"{{ .Dir }}/doc_hello_world.json",
				"{{ .Dir }}/static.json",
				"aa58",
				"--pretty",
			},
			false,
			map[string]string{"static.json": `{
    "hello": "world"
}` + "\n"},
		},
	}
	for _, test := range tests {
		dir := setupTestDir(t)
		defer removeTestDir(t, dir)

		for i, arg := range test.Args {
			arg_transformed, err := tmplExecute(arg, dir,
				struct{ Dir string }{dir})
			if err != nil {
				t.Fatal(err)
			}
			test.Args[i] = arg_transformed
		}

		err := Main(test.Args, false)
		if test.ExpectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		for filename, content := range test.ExpectedOutput {
			assertFileHasContent(t, path.Join(dir, filename), content)
		}
	}
}

func TestGetFilehandles(t *testing.T) {
	tests := []struct {
		InputFilename  string
		OutputFilename string
		Pretty         bool
		ExpectedError  string
		ExpectedOutput map[string]string
	}{
		// Bad path for input
		{
			"foo", "bar",
			true,
			"open {{ .Dir }}/{{ .InputFilename }}: no such file or directory",
			map[string]string{},
		},
		// Bad path for output
		{
			"input_hello_world.json", "",
			true,
			"open {{ .Dir }}: is a directory",
			map[string]string{},
		},
		// Good paths
		{
			"input_hello_world.json", "output.json",
			true,
			"",
			map[string]string{
				"input_hello_world.json": `{ "hello": "world" }`,
				"output.json": `{
    "output": "was written here"
}` + "\n",
			},
		},
		// Again, with prettiness turned off
		{
			"input_hello_world.json", "output.json",
			false,
			"",
			map[string]string{
				"input_hello_world.json": `{ "hello": "world" }`,

				"output.json": `{"output":"was written here"}` + "\n",
			},
		},
	}
	for _, test := range tests {
		dir := setupTestDir(t)
		defer removeTestDir(t, dir)

		experr, err := tmplExecute(test.ExpectedError, dir, struct {
			Dir, InputFilename, OutputFilename string
		}{dir, test.InputFilename, test.OutputFilename})
		if err != nil {
			t.Fatal(err)
		}
		test.ExpectedError = experr

		input, output, err := getFilehandles(
			path.Join(dir, test.InputFilename),
			path.Join(dir, test.OutputFilename),
			test.Pretty,
		)
		if test.ExpectedError == "" {
			if assert.NoError(t, err) {
				assert.NoError(t, output.Write(map[string]interface{}{
					"output": "was written here",
				}))
			}
		} else {
			if assert.Error(t, err) {
				assert.Equal(t, test.ExpectedError, err.Error())
			}
			assert.Equal(t, nil, input)
			assert.Equal(t, nil, output)
		}

		for filename, content := range test.ExpectedOutput {
			assertFileHasContent(t, path.Join(dir, filename), content)
		}
	}
}

func tmplExecute(pattern, dir string, data interface{}) (string, error) {
	tmpl, err := template.New(dir).Parse(pattern)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	return buf.String(), err

}

func assertFileHasContent(t *testing.T, filename, expected_content string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected_content, string(content))
}

func setupTestDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "djconvert-tests-")
	if err != nil {
		t.Fatal(err)
	}
	// Set up files
	files := map[string]string{
		"input_hello_world.json": `{ "hello": "world" }`,
		"doc_hello_world.json":   `{"events":{"aa582b4df04ba01af5205e702d4d16ed0b2c0705":{"parent":"","handler":"SET","args":{"path":[],"value":{"hello":"world"}}}},"timestamps":["aa582b4df04ba01af5205e702d4d16ed0b2c0705"]}` + "\n",
	}
	for filename, content := range files {
		fullpath := path.Join(dir, filename)
		err = ioutil.WriteFile(fullpath, []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func removeTestDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}
