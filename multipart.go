package formy

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// Writer is a wrapper around [multipart.Writer].
type Writer struct {
	mw       *multipart.Writer
	detectCt bool
	firstErr error
}

// NewWriter is a wrapper around [multipart.NewWriter] which is auto-detecting content type by default
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		mw:       multipart.NewWriter(w),
		detectCt: true,
	}
}

// DetectContentType used to turn on/off content type detection
func (w *Writer) DetectContentType(b bool) {
	w.detectCt = b
}

// Boundary is a wrapper around [multipart.Writer.Boundary]
func (w Writer) Boundary() string {
	return w.mw.Boundary()
}

// FormDataContentType is a wrapper around [multipart.Writer.FormDataContentType]
func (w Writer) FormDataContentType() string {
	return w.mw.FormDataContentType()
}

// WriteString is a wrapper around [multipart.Writer.WriteField]
func (w *Writer) WriteString(fieldname, str string) *Writer {
	if w.firstErr == nil {
		w.firstErr = w.mw.WriteField(fieldname, str)
	}
	return w
}

// WriteOptionalString is a wrapper around [multipart.Writer.WriteField]
// that writes the string only if str is not empty
func (w *Writer) WriteOptionalString(fieldname string, str string) *Writer {
	if str != "" {
		return w.WriteString(fieldname, str)
	}
	return w
}

// WriteAnyTextField is equivalent to creating a part and writing val using [fmt.Fprint]
// with the part as writer and val as value
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

// WriteInt creates a part with the given fieldname and writes i as is.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteInt(fieldname string, i int) *Writer {
	return w.WriteAnyTextField(fieldname, i)
}

// WriteOptionalInt creates a part with the given fieldname and writes i if it's not zero.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteOptionalInt(fieldname string, i int) *Writer {
	if i != 0 {
		return w.WriteAnyTextField(fieldname, i)
	}
	return w
}

// WriteBool creates a part with the given fieldname and writes b as is.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteBool(fieldname string, b bool) *Writer {
	return w.WriteAnyTextField(fieldname, b)
}

// WriteOptionalBool creates a part with the given fieldname and writes b if it's true.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteOptionalBool(fieldname string, b bool) *Writer {
	if b == true {
		return w.WriteAnyTextField(fieldname, b)
	}
	return w
}

// WriteFloat32 creates a part with the given fieldname and writes f as is.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteFloat32(fieldname string, f float32) *Writer {
	return w.WriteAnyTextField(fieldname, f)
}

// WriteOptionalFloat32 creates a part with the given fieldname and writes f if it's not zero.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteOptionalFloat32(fieldname string, f float32) *Writer {
	if f != 0 {
		return w.WriteAnyTextField(fieldname, f)
	}
	return w
}

// WriteFloat64 creates a part with the given fieldname and writes f as is.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteFloat64(fieldname string, f float64) *Writer {
	return w.WriteAnyTextField(fieldname, f)
}

// WriteFloat64 creates a part with the given fieldname and writes f if it's not zero.
// It is a wrapper around [Writer.WriteAnyTextField]
func (w *Writer) WriteOptionalFloat64(fieldname string, f float64) *Writer {
	if f != 0 {
		return w.WriteAnyTextField(fieldname, f)
	}
	return w
}

// WriteJSON creates a part with the given fieldname and writes v as JSON encoded value.
// V can't be nil
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

// TODO: add optional arrays, maps and slices that will check for default and nil value
// and write it as JSON

// WriteJSON creates a part with the given fieldname and writes v as JSON encoded value,
// or does nothing if it's nil, and does not return an error
func (w *Writer) WriteOptionalJSON(fieldname string, v any) *Writer {
	if w.firstErr == nil {
		if fieldname == "" {
			w.firstErr = fmt.Errorf("empty field name")
			return w
		}
		if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
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

// WriteFile creates a part with the given fieldname and filename and writes the file into the part.
// If w.detectCt is true, it will read the first 3072 bytes
// and automatically set the "Content-Type" header to the most suitable MIME type.
// Otherwise, or if the detection failed, "application/octet-stream" will be used instead
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

		var (
			err error
			buf []byte
		)

		// reading it to both detect content type and write it to the part
		buf, err = io.ReadAll(file)
		if err != nil {
			w.firstErr = err
			return w
		}

		var h textproto.MIMEHeader
		if w.detectCt {
			h = fileFieldHeader(fieldname, filename, buf)
		} else {
			h = fileFieldHeader(fieldname, filename, nil)
		}
		part, err := w.mw.CreatePart(h)
		if err != nil {
			w.firstErr = err
			return w
		}

		_, err = part.Write(buf)
		if err != nil {
			w.firstErr = err
			return w
		}
	}
	return w
}

// Close returns the first error occurred while writing any fields,
// or the result of [multipart.Writer.Close]
func (w *Writer) Close() error {
	if w.firstErr != nil {
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

func fileFieldHeader(fieldname, filename string, buf []byte) textproto.MIMEHeader {
	h := textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldname), escapeQuotes(filename))},
	}
	if buf != nil {
		ct := mimetype.Detect(buf)
		h.Set("Content-Type", ct.String())
	} else {
		h.Set("Content-Type", "application/octet-stream")
	}
	return h
}

var quoteReplacer = strings.NewReplacer("\\", "\\\\", `"`, "\\\\")

func escapeQuotes(raw string) string {
	return quoteReplacer.Replace(raw)
}
