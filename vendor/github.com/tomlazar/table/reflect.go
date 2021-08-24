package table

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// Marshal the slince into a table format using reflection
func Marshal(arr interface{}, c *Config) ([]byte, error) {
	var buf bytes.Buffer
	err := MarshalTo(&buf, arr, c)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalTo writes the reflected table into the passed in io.Writer
func MarshalTo(w io.Writer, arr interface{}, c *Config) error {
	tab, err := parse(arr)
	if err != nil {
		return err
	}

	return tab.WriteTable(w, c)
}

// parse is the main method for refletion right now
func parse(arr interface{}) (*Table, error) {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("arr must be a slice")
	}

	if v.Len() < 1 {
		return nil, errors.New("arr must have at least one element")
	}

	head := v.Index(0)

	tab := Table{
		Headers: make([]string, head.NumField()),
		Rows:    make([][]string, v.Len()),
	}
	for i := 0; i < head.NumField(); i++ {
		f := head.Type().Field(i)
		n, ok := f.Tag.Lookup("table")
		if !ok {
			n = f.Name
		}

		tab.Headers[i] = n
	}

	for i := 0; i < v.Len(); i++ {
		ref := v.Index(i)

		tab.Rows[i] = make([]string, ref.NumField())
		for f := 0; f < ref.NumField(); f++ {
			tab.Rows[i][f] = fmt.Sprintf("%v", ref.Field(f).Interface())
		}
	}

	return &tab, nil
}
