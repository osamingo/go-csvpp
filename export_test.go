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
)
