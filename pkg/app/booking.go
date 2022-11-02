package app

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

const maxSeats = 6

type SeatingManager interface {
	Arrives(group *Group)
	GetWaitingGroups() []*Group
	GetBookings() []*Booking
}

type seatingManager struct {
	mux       sync.RWMutex
	tablesMap map[uuid.UUID]Table
	// availableTables represents a map where:
	// key: available seats number.
	// value: table IDs having that number of available seats.
	//availableTables map[Seats][]uuid.UUID
	availableTables AvailableTables
	waitingQueue    []*Group
	booking         []*Booking
}

func NewSeatManager(tables Tables, waitingQueue []*Group) SeatingManager {

	availableSeats := make(map[Seats][]uuid.UUID)
	for _, table := range tables {
		_, ok := availableSeats[table.capacity]
		if !ok {
			availableSeats[table.capacity] = make([]uuid.UUID, 0)
		}
		availableSeats[table.capacity] = append(availableSeats[table.capacity], table.id)
	}

	return &seatingManager{tablesMap: tables.Map(), waitingQueue: waitingQueue, availableTables: availableSeats, booking: make([]*Booking, 0)}

}

func (m *seatingManager) GetBookings() []*Booking {
	return m.booking
}

func (m *seatingManager) GetWaitingGroups() []*Group {
	return m.waitingQueue
}

func (m *seatingManager) Arrives(group *Group) {
	// this is for managing concurrency and enforcing order respecting.
	m.mux.Lock()
	defer m.mux.Unlock()

	booking, err := m.allocateGroup(group)
	if err != nil {
		m.waitingQueue = append(m.waitingQueue, group)
		return
	}

	m.booking = append(m.booking, booking)
}

func (m *seatingManager) allocateGroup(group *Group) (*Booking, error) {
	//table := m.findTable(group)

	tableID := m.availableTables.Find(group.seats)
	if tableID == uuid.Nil {
		return nil, fmt.Errorf("table not found for group %v with %v seats", group.id, group.seats)
	}

	table, ok := m.tablesMap[tableID]
	if !ok {
		return nil, fmt.Errorf("table id %v not found", tableID)
	}

	if err := table.Seat(group.seats); err != nil {
		return nil, fmt.Errorf("could not seat group %v in table %v", group.id, table.id)
	}
	
	return &Booking{groupID: group.id, tableID: table.id}, nil
}

func (m *seatingManager) findTable(group *Group) *Table {
	for seats := group.seats; seats <= maxSeats; seats++ {
		tables, ok := m.availableTables[seats]
		// seats' number not found in tables available list, continue
		if !ok || len(tables) == 0 {
			continue
		}
		tableID := tables[0]
		// remove found table from available table lists
		m.availableTables[seats] = tables[1:]
		table := m.tablesMap[tableID]
		return &table
	}
	return nil
}

func (m *seatingManager) addAsAvailableTable(table *Table) {
	tables, ok := m.availableTables[table.availableSeats]
	switch {
	case ok:
		tables = append(tables, table.id)
	default:
		m.availableTables[table.availableSeats] = []uuid.UUID{table.id}
	}
}
