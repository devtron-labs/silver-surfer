# Table

![Go](https://github.com/tomlazar/table/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/tomlazar/table/branch/main/graph/badge.svg?token=F96DTCC4MC)](undefined)
[![GoDoc](https://godoc.org/github.com/tomlazar/table?status.svg)](http://godoc.org/github.com/tomlazar/table)
[![Go Report Card](https://goreportcard.com/badge/github.com/tomlazar/table)](https://goreportcard.com/report/github.com/tomlazar/table)

Print out tabular data on the command line using the ansi color esacape codes. Support for writing the ouput based on the fields in a struct and for defining and creating the table manully using the underlying object.

Support for colors on windows can be don using [mattn/go-colorable](https://github.com/mattn/go-colorable) to make a `io.Writer` that will work.

## Usage

For creating a table yourself using the struct.
```go
	tab := table.Table{
		Headers: []string{"something", "another"},
		Rows: [][]string{
			{"1", "2"},
			{"3", "4"},
			{"3", "a longer piece of text that should stretch"},
			{"but this one is longer", "shorter now"},
		},
	}
	err := tab.WriteTable(w, nil) // w is any io.Writer
```

With a struct slice 
```go
	data := []struct {
		Name     string `table:"THE NAME"`
		Location string `table:"THE LOCATION"`
	}{
		{"name", "l"},
		{"namgfcxe", "asfdad"},
		{"namr3e", "l134151dsa"},
		{"namear", "lasd2135"},
	}

    err := table.MarshalTo(w, data, nil) // writes to any w = io.Writer
    buf, err := table.Marshal(data, nil) // also supports return the bytes
```

The nil parameter is the configuration for the table, this can be set manually, but if its left as nil the deafult config settings will be used.
```go
type Config struct {
	ShowIndex       bool     // shows the index/row number as the first column
	Color           bool     // use the color codes in the output
	AlternateColors bool     // alternate the colors when writing
	TitleColorCode  string   // the ansi code for the title row
	AltColorCodes   []string // the ansi codes to alternate between
}
```


## Installation 

Go makes this part easy.

```bash
$ go get github.com/mattn/go-colorable
```

## License

MIT

## Author 

Tom Lazar