package importjob

import "time"

// MaxImportRows is the maximum number of rows allowed in a single import
const MaxImportRows = 10000

// DefaultTimezone is the default timezone for parsing times in imports
// JST (Japan Standard Time) = UTC+9
var DefaultTimezone = time.FixedZone("JST", 9*60*60)
