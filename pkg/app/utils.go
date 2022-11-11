package app

import (
	"fmt"
	"github.com/google/uuid"
)

func (m *seatingManager) allocateGroup(group *Group) (*Booking, error) {
	table, findErr := m.findTable(group)
	if findErr != nil {
		return nil, fmt.Errorf("table not found for group %v", group.id)
	}

	if err := m.allocateGroupInTable(table, group); err != nil {
		return nil, err
	}

	return &Booking{groupID: group.id, tableID: table.id}, nil
}

func (m *seatingManager) findTable(group *Group) (*Table, error) {
	tableID := m.availableTables.Pickup(group.seats)
	if tableID == uuid.Nil {
		return nil, fmt.Errorf("table not found for group %v with %v seats", group.id, group.seats)
	}

	table, ok := m.tablesMap[tableID]
	if !ok {
		return nil, fmt.Errorf("table id %v not found", tableID)
	}

	return &table, nil
}

func (m *seatingManager) allocateGroupInTable(table *Table, group *Group) error {
	defer m.availableTables.Push(table)

	if err := table.Seat(group.seats); err != nil {
		return fmt.Errorf("could not seat group %v in table %v", group.id, table.id)
	}

	return nil
}

func (m *seatingManager) getTableByGroup(groupID uuid.UUID) *Table {
	for _, booking := range m.booking {
		if booking.groupID == groupID {
			table := m.tablesMap[booking.tableID]
			return &table
		}
	}
	return nil
}

func (m *seatingManager) getBookingByGroupID(id uuid.UUID) *Booking {
	for _, booking := range m.booking {
		if booking.groupID == id {
			return booking
		}
	}
	return nil
}
