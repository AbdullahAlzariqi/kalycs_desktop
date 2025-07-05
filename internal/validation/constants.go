package validation

// Project validation constants
const (
	MaxProjectNameLength        = 25
	MaxProjectDescriptionLength = 200
	MinProjectNameLength        = 1
)

// Rule validation constants
const (
	MaxRuleNameLength = 25
	MinRuleNameLength = 1
	MaxRuleTextLength = 64
	MaxRuleTextsItems = 20
)

// Common validation constants
const (
	UUIDLength      = 36
	UUIDHyphenCount = 4
)

// Valid rule types
var ValidRuleTypes = []string{
	"starts_with",
	"contains",
	"ends_with",
	"extension",
	"regex",
}
