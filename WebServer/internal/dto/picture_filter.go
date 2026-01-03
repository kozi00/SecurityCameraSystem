// PictureFilters describe user-provided filters to narrow the picture list.
package dto

import "time"

type PictureFilters struct {
	Camera     string
	Object     string
	DateAfter  time.Time
	DateBefore time.Time
	TimeAfter  time.Time
	TimeBefore time.Time
}
