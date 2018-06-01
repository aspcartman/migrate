package migrate

import (
	"fmt"
	"strings"
)

type Migration struct {
	ID       int
	Name     string
	Up, Down string
}

func NewMigration(id int, name string, script string) (Migration, error) {
	m := Migration{
		ID:   id,
		Name: name,
	}

	spl := strings.Split(script, "-- DOWN")
	if len(spl) > 0 {
		m.Up = spl[0]
	}
	if len(spl) > 1 {
		m.Down = spl[1]
	}
	if len(spl) > 2 {
		return Migration{}, fmt.Errorf("more than 2 DOWN sections")
	}

	return m, nil
}

// sorting
type migsSlice []Migration

func (s migsSlice) Len() int {
	return len(s)
}

func (s migsSlice) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func (s migsSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
