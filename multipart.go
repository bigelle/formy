package formy

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type Writer struct {
	mw       *multipart.Writer
	detectCt bool
	firstErr error
}

func (w *Writer) DetectContentType(b bool) {
	w.detectCt = b
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		mw: multipart.NewWriter(w),
		detectCt: true,
	}
}

func (w Writer) Boundary() string {
	return w.mw.Boundary()
}

func (w Writer) FormDataContentType() string {
	return w.mw.FormDataContentType()
}

func (w *Writer) WriteString(fieldname, str string) *Writer {
	if w.firstErr == nil {
		w.firstErr = w.mw.WriteField(fieldname, str)
	}
	return w
}

func (w *Writer) WriteAnyTextField(fieldname string, val any) *Writer {
	if w.firstErr == nil {
		if fieldname == "" {
			w.firstErr = fmt.Errorf("empty field name")
			return w
		}
		if val == nil {
			w.firstErr = fmt.Errorf("empty field value")
			return w
		}

		part, err := w.mw.CreatePart(textFieldHeader(fieldname))
		if err != nil {
			w.firstErr = err
			return w
		}

		if _, err = fmt.Fprint(part, val); err != nil {
			w.firstErr = err
			return w
		}
	}
	return w
}

func (w *Writer) WriteInt(fieldname string, i int) *Writer {
	return w.WriteAnyTextField(fieldname, i)
}

func (w *Writer) WriteBool(fieldname string, b bool) *Writer {
	return w.WriteAnyTextField(fieldname, b)
}

func (w *Writer) WriteFloat32(fieldname string, f float32) *Writer {
	return w.WriteAnyTextField(fieldname, f)
}

func (w *Writer) WriteFloat64(fieldname string, f float64) *Writer {
	return w.WriteAnyTextField(fieldname, f)
}

func (w *Writer) WriteJSON(fieldname string, v any) *Writer {
	if w.firstErr == nil {
		if fieldname == "" {
			w.firstErr = fmt.Errorf("empty field name")
			return w
		}
		if v == nil {
			w.firstErr = fmt.Errorf("empty field value")
			return w
		}

		part, err := w.mw.CreatePart(textFieldHeader(fieldname))
		if err != nil {
			w.firstErr = err
			return w
		}

		enc := json.NewEncoder(part)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err != nil {
			w.firstErr = err
			return w
		}
	}
	return w
}

func (w *Writer) WriteFile(fieldname, filename string, file io.Reader) *Writer {
	if w.firstErr == nil {
		if fieldname == "" {
			w.firstErr = fmt.Errorf("empty field name")
			return w
		}
		if filename == "" {
			w.firstErr = fmt.Errorf("empty file name")
			return w
		}
		if file == nil {
			w.firstErr = fmt.Errorf("empty file reader")
			return w
		}

		part, err := w.mw.CreatePart(fileFieldHeader(fieldname, filename, file, w.detectCt))
		if err != nil {
			w.firstErr = err
			return w
		}

		_, err = io.Copy(part, file)
		if err != nil {
			w.firstErr = err
			return w
		}
	}
	return w
}

func (w *Writer) Close() error {
	if w.firstErr != nil {
		w.mw.Close()
		return w.firstErr
	}
	return w.mw.Close()
}

func textFieldHeader(fieldname string) textproto.MIMEHeader {
	h := textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldname))},
	}
	return h
}

func fileFieldHeader(fieldname, filename string, file io.Reader, detectCt bool) textproto.MIMEHeader {
	h := textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldname), escapeQuotes(filename))},
	}
	if file != nil && detectCt {
		ct, err := mimetype.DetectReader(file)
		if err != nil {
			return h
		}
		h.Set("Content-Type", ct.String())
	}
	return h
}

var quoteReplacer = strings.NewReplacer("\\", "\\\\", `"`, "\\\\")

func escapeQuotes(raw string) string {
	return quoteReplacer.Replace(raw)
}
