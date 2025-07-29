package formy_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/bigelle/formy"
	"github.com/stretchr/testify/assert"
)

func TestWriter_AnyWrites(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := formy.NewWriter(buf)

	err := w.WriteString("string", "text").
		WriteInt("int", 42).
		WriteBool("bool", false).
		WriteFloat64("float64", float64(0.42)).
		WriteFloat32("float32", float32(0.42)).
		WriteFile("file", "file.txt", strings.NewReader("TEST DEEZ NUTS")).
		Close()

	if assert.NoError(t, err) {
		r := multipart.NewReader(buf, w.Boundary())
		for {
			part, err := r.NextPart()
			if err == io.EOF {
				break
			}

			// TODO: use strconv or smth
			switch part.FormName() {
			case "string":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "text", string(buf))
			case "int":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "42", string(buf))
			case "bool":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "false", string(buf))
			case "float64":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "0.42", string(buf))
			case "float32":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "0.42", string(buf))
			case "file.txt":
				buf, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "TEST DEEZ NUTS", string(buf))
				assert.Equal(t, "file.txt", part.FileName())
			}
		}
	}
}
