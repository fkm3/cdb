package cdb

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type rec struct {
	key    string
	values []string
}

var records = []rec{
	{"one", []string{"1"}},
	{"two", []string{"2", "22"}},
	{"three", []string{"3", "33", "333"}},
}

var data []byte // set by init()

func TestCdb(t *testing.T) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create temp file: %s", err)
	}

	defer os.Remove(tmp.Name())

	// Test Make
	err = Make(tmp, bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Make failed: %s", err)
	}

	// Test reading records
	c, err := Open(tmp.Name())
	if err != nil {
		t.Fatalf("Error opening %s: %s", tmp.Name(), err)
	}
	for _, rec := range records {
		key := []byte(rec.key)
		values := rec.values

		v, err := c.Bytes(key)
		if err != nil {
			t.Fatalf("Record read failed: %s", err)
		}

		if !bytes.Equal(v, []byte(values[0])) {
			t.Fatal("Incorrect value returned")
		}

		iter := c.Iterate(key)
		for _, value := range values {
			sr, err := iter.NextReader()
			if err != nil {
				t.Fatalf("Record read failed: %s", err)
			}

			data := make([]byte, sr.Size())
			_, err = sr.Read(data)
			if err != nil {
				t.Fatalf("Record read failed: %s", err)
			}

			if !bytes.Equal(data, []byte(value)) {
				t.Fatal("value mismatch")
			}
		}
		// Read all values, so should get EOF
		_, err = iter.NextReader()
		if err != io.EOF {
			t.Fatalf("Expected EOF, got %s", err)
		}
	}

	// Test Dump
	if _, err = tmp.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	err = Dump(buf, tmp)
	if err != nil {
		t.Fatalf("Dump failed: %s", err)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Fatalf("Dump round-trip failed")
	}
}

func init() {
	b := bytes.NewBuffer(nil)
	for _, rec := range records {
		key := rec.key
		for _, value := range rec.values {
			b.WriteString(fmt.Sprintf("+%d,%d:%s->%s\n", len(key), len(value), key, value))
		}
	}
	b.WriteByte('\n')
	data = b.Bytes()
}
