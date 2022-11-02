package app

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Table struct {
	id             uuid.UUID
	capacity       Seats
	availableSeats Seats
}

func NewTable(capacity uint8) (*Table, error) {
	seats := Seats(capacity)
	if err := seats.Validate(); err != nil {
		return nil, err
	}
	return &Table{id: uuid.New(), capacity: seats, availableSeats: seats}, nil
}

func (t *Table) Seat(people Seats) error {
	if people > t.availableSeats {
		return fmt.Errorf("no space in table %v", t.id)
	}
	t.availableSeats = t.availableSeats - people
	return nil
}

type Group struct {
	id    uuid.UUID
	seats Seats
}

func NewGroup(size uint8) *Group {
	return &Group{id: uuid.New(), seats: Seats(size)}
}

type Booking struct {
	groupID uuid.UUID
	tableID uuid.UUID
}

type Tables []*Table

// Map return tables list as map where key is ID table and value is correspondent table item.
func (t Tables) Map() map[uuid.UUID]Table {
	tablesMap := make(map[uuid.UUID]Table)

	for _, table := range t {
		tablesMap[table.id] = Table{id: table.id, capacity: table.capacity, availableSeats: table.availableSeats}
	}

	return tablesMap
}

// AvailableSeats represents a map where:
// key: available seats number.
// value: table IDs having that number of available seats.
type AvailableSeats map[int][]uuid.UUID

type Seats uint8

func (s Seats) Validate() error {
	if s > 6 {
		return fmt.Errorf("invalid seats number %v", s)
	}
	return nil
}

type TableIDStack struct {
	mutex  sync.RWMutex
	tables []uuid.UUID
}

func (t *TableIDStack) Push(tableID uuid.UUID) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.tables == nil {
		t.tables = []uuid.UUID{}
	}

	t.tables = append(t.tables, tableID)
}

func (t *TableIDStack) Pop() uuid.UUID {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(t.tables) == 0 {
		return uuid.Nil
	}

	item := t.tables[len(t.tables)-1]
	t.tables = t.tables[0 : len(t.tables)-1]
	return item
}

type AvailableTables map[Seats][]uuid.UUID

func (a AvailableTables) Find(desiredSeats Seats) uuid.UUID {
	for seats := desiredSeats; seats <= maxSeats; seats++ {
		tables, ok := a[seats]
		// seats' number not found in tables available list, continue
		if !ok || len(tables) == 0 {
			continue
		}
		tableID := tables[0]
		// remove found table from available table lists
		a[seats] = tables[1:]

		return tableID
	}
	return uuid.Nil
}
