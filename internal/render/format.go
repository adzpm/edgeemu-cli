package render

// Output format names accepted by the -f/--format flag.
const (
	FormatList = "list"
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatXML  = "xml"
	FormatCSV  = "csv"
)

// Formats returns all supported output formats in display order.
func Formats() []string {
	return []string{FormatList, FormatJSON, FormatYAML, FormatXML, FormatCSV}
}
