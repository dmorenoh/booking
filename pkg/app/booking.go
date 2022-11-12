package app

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

const maxSeats = 6

type SeatingManager interface {
	Arrives(newGroup *Group)
	Leaves(groupID uuid.UUID) error
	Locate(groupID uuid.UUID) (Table, error)
	GetWaitingGroups() []*Group
	GetBookings() []*Booking
	GetTable(tableID uuid.UUID) (*Table, error)
}

type seatingManager struct {
	sync.RWMutex
	tablesMap map[uuid.UUID]Table
	groupMap  map[uuid.UUID]Group
	// availableTables represents a map where:
	// key: available seats number.
	// value: table IDs having that number of available seats.
	//availableTables map[Seats][]uuid.UUID

	availableTables *AvailableTables
	waitingQueue    []*Group
	booking         []*Booking
}

func NewSeatManager(tables Tables, waitingQueue []*Group) SeatingManager {
	availableSeats := NewAvailableTables(tables)

	return &seatingManager{
		tablesMap:       tables.Map(),
		waitingQueue:    waitingQueue,
		availableTables: availableSeats,
		booking:         make([]*Booking, 0),
		groupMap:        make(map[uuid.UUID]Group),
	}
}

func (m *seatingManager) Arrives(newGroup *Group) {
	// this is for managing concurrency and enforcing requests order.
	m.Lock()
	defer m.Unlock()

	m.groupMap[newGroup.id] = *newGroup

	booking, err := m.allocateGroup(newGroup)
	if err != nil {
		m.waitingQueue = append(m.waitingQueue, newGroup)
		return
	}

	m.booking = append(m.booking, booking)
}

func (m *seatingManager) Leaves(groupID uuid.UUID) error {
	m.Lock()
	defer m.Unlock()

	group, ok := m.groupMap[groupID]
	if !ok {
		return fmt.Errorf("group %v not found", groupID)
	}

	// check if group has table booked if so, release it
	booking := m.getBookingByGroupID(groupID)
	if booking == nil {
		return nil
	}

	table, err := m.GetTable(booking.tableID)
	if err != nil {
		return err
	}

	if releaseErr := table.Release(group.seats); releaseErr != nil {
		return fmt.Errorf("cannot release seats left for %v in table %v", group.id, table.id)
	}

	return nil
}

func (m *seatingManager) Locate(groupID uuid.UUID) (Table, error) {
	group, ok := m.groupMap[groupID]
	if !ok {
		return Table{}, errors.New("group not found")
	}

	booking := m.getBookingByGroupID(group.id)
	if booking == nil {
		return Table{}, nil
	}

	table, ok := m.tablesMap[booking.tableID]
	if !ok {
		return Table{}, errors.New("table not found")
	}

	return table, nil
}

func (m *seatingManager) GetBookings() []*Booking {
	return m.booking
}

func (m *seatingManager) GetWaitingGroups() []*Group {
	return m.waitingQueue
}

func (m *seatingManager) GetTable(tableID uuid.UUID) (*Table, error) {
	table, ok := m.tablesMap[tableID]
	if !ok {
		return nil, errors.New("table not found")
	}

	return &table, nil
}
