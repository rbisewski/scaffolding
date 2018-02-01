# scaffolding

Convert rosewood tables into plain-text CSV files.

## Building

To build this program, ensure that golang and GNU Make is installed, and
use the below command:

`make`

Currently compilation works best on POSIX compatible machines that implement
the `date` command. Other architectures will build, but will have incomplete
version information.

## Usage

An example usage of the `scaffolding` program is as follows:

```
./scaffolding -tables "conditions-table,screening-table" -indir /path/to/tables -outdir /path/to/write/output
```

Where `tables` is a CSV delimited list of table names to search for an
attempt to read, `indir` is the path to where the input plain-text files
are stored and `outdir` is the location to write the output file. By default
the output file name is `rosewood.csv`. The file itself can then be opened
with software like LibreOffice.

To import them as LibreOffice Writer tables, do the following:

1) Open the file.

2) Select the individual set of tabe lines.

3) Table->Convert->Text To Table

4) Choose the comma delimiter.

5) The imported plain-text has now been converted to a table.

Consider running the program with the `--help` flag for additional
information regarding these flags and what options are available.

## Testing

To run the current test suite of this program, type the following command:

`make test`

If all of the tests pass and are listed as ok, then the IO functions of this
program work as expected.

## TODO

* Consider adding an option to create ODT files with autopopulated tables.
* Add flag to give end-user the option to create either CSV or ODT files.
