package bed

import "time"

type Data struct {
	Entries []*Entry `json:"entries"`
}

type Entry struct {
	When time.Time `json:"when"`
}
