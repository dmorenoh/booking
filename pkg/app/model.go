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

func (t *Table) Release(people Seats) error {
	if t.capacity > t.availableSeats+people {
		return fmt.Errorf("error releasing %v for table %v", people, t.id)
	}
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

type FreeSeats struct {
	sync.RWMutex
	freeSeats map[Seats][]uuid.UUID
}

func (f *FreeSeats) Pickup(desiredSeats Seats) uuid.UUID {
	f.Lock()
	defer f.Unlock()

	for seats := desiredSeats; seats <= maxSeats; seats++ {
		tables, ok := f.freeSeats[seats]
		// seats' number not found in tables available list, continue
		if !ok || len(tables) == 0 {
			continue
		}
		tableID := tables[0]
		// remove found table from available table lists
		f.freeSeats[seats] = tables[1:]

		return tableID
	}

	return uuid.Nil
}

func (f *FreeSeats) Remove(table Table) {
	f.Lock()
	defer f.Unlock()
	tableIDs := f.freeSeats[table.availableSeats]
	for i, tableId := range tableIDs {
		if tableId != table.id {
			f.freeSeats[table.availableSeats] = append(tableIDs[:i], tableIDs[i+1:]...)
			return
		}
	}
}

type AvailableTables struct {
	sync.RWMutex
	seatsMap map[Seats][]uuid.UUID
}

//type AvailableTables map[Seats][]uuid.UUID

func (a *AvailableTables) Pickup(desiredSeats Seats) uuid.UUID {
	a.RLock()
	defer a.RUnlock()

	for seats := desiredSeats; seats <= maxSeats; seats++ {
		tables, ok := a.seatsMap[seats]
		// seats' number not found in tables available list, continue
		if !ok || len(tables) == 0 {
			continue
		}
		tableID := tables[0]
		// remove found table from available table lists
		a.seatsMap[seats] = tables[1:]

		return tableID
	}
	return uuid.Nil
}

func (a *AvailableTables) Push(table *Table) {
	a.Lock()
	defer a.Unlock()

	_, ok := a.seatsMap[table.availableSeats]
	if !ok {
		a.seatsMap[table.availableSeats] = make([]uuid.UUID, 0)
	}

	a.seatsMap[table.availableSeats] = append(a.seatsMap[table.availableSeats], table.id)
}

func (a *AvailableTables) Remove(table *Table) {
	a.Lock()
	defer a.Unlock()

	uuids := a.seatsMap[table.availableSeats]
	for i, u := range uuids {
		if u == table.id {
			uuids = uuids[i+1:]
		}
	}
}

// NewAvailableTables create a k,v map where k is number of free seats and v is list of tables with that amount of free seats
func NewAvailableTables(tables []*Table) *AvailableTables {
	availableSeatsMap := make(map[Seats][]uuid.UUID)
	for _, table := range tables {
		_, ok := availableSeatsMap[table.capacity]
		if !ok {
			availableSeatsMap[table.capacity] = make([]uuid.UUID, 0)
		}
		availableSeatsMap[table.capacity] = append(availableSeatsMap[table.capacity], table.id)
	}

	return &AvailableTables{seatsMap: availableSeatsMap}
}
