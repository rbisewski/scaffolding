package main

const usageMessage = `
Convert rosewood tables into ISO standard ODT files or CSV files.

Usage: identify_conditions
       -csv
       -tables <comma,separated,list,of,tables>
       -indir  <path_to_input_directory>
       -outdir <path_to_output_directory>

Arguments:
	h, help       Prints this usage message
  	version       Prints the current program version and build info
	csv           Prints the given rosewood tables as plain-text CSVs (default is ODT)
	tables        Comma separated list of tables; e.g. "table-w-conditions,table-wo-screening"
	outdir        Output location; e.g. /path/to/output/directory

	Description:
		The ODT values created by this program can be read by Libreoffice or
		imported into other software. Word tends to complain about the file
		being non-standard, however, in practice it will import it without
		problems. Afterwards the end-user may save it via Word into other
		formats should they need.

		The CSV values created by this program can be imported and used as tables
		in LibreOffice or other software. A sample output file looks like:

		name-of-table-w-conditions
		variable, ci of cases, ci of controls
		condition_1, ci_cases_value, ci_of_controls
		condition_2, ci_cases_value, ci_of_controls

		name-of-table-wo-screening
		variable, ci of cases, ci of controls
		col_1, ci_cases_value, ci_of_controls
		col_2, ci_cases_value, ci_of_controls`
