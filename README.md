# scaffolding

Convert rosewood tables into ISO standard ODT files.

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

Where `tables` is a comma delimited list of table names to search for an
attempt to read, `indir` is the path to where the input plain-text files
are stored and `outdir` is the location to write the output file. By default
the output file name is `rosewood.odt`. The file itself can then be opened
with software like LibreOffice.

This program can also convert Rosewood tables into CSV using the `-csv` flag,
in the event that the end-users wishes to have plain-text or wants to manually
generate the output tables. In that scenario, follow the below steps...

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

* Consider figuring out a way to read-in rosewood commands, assuming this
  is needed at some point in the future.
* Check if the ODT format generated by this code can be read by MS Office
  / Libreoffice on platforms other than Linux.
* Test that tables with 3+ columns still have correct styling.
* Add the ability to combine lone single-cell rows to the full length of
  table.
* Since ODT tables use letter sequencing, figure out how to make tables
  with 26+ columns.
* Add the ability to convert rosewood footnotes like `^a` to `^z` to
  superscripted numerals + ODT Footnotes for the purpose of generating more
  complex tables from the rosewood plaintext.
