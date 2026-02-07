package csvpp

// Export unexported functions for testing.
var (
	ParseColumnHeader             = parseColumnHeader
	ParseColumnHeaderWithDepth    = parseColumnHeaderWithDepth
	ParseHeaderRecordWithMaxDepth = parseHeaderRecordWithMaxDepth
	ParseName                     = parseName
	ParseArrayDelimiter           = parseArrayDelimiter
	IsFieldChar                   = isFieldChar
	SplitByRune                   = splitByRune
	SplitByDelimiter              = splitByDelimiter
	FormatColumnHeader            = formatColumnHeader
	FormatComponentList           = formatComponentList
	ExtractTagName                = extractTagName
	CachedTypeInfo                = cachedTypeInfo
)

// ReaderLine returns the current line number of the reader for testing.
func ReaderLine(r *Reader) int {
	return r.line
}
