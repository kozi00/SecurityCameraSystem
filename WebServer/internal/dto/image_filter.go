// ImageFilters describe user-provided filters to narrow the image list.
package dto

import "time"

type ImageFilters struct {
	Camera     string
	Object     string
	DateAfter  time.Time
	DateBefore time.Time
	TimeAfter  time.Time
	TimeBefore time.Time
}
