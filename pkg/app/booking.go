package pkg

import "errors"

const maxSeats = 6

type Table struct {
	id             int
	size           int
	availableSeats int
}

func (t *Table) Seat(people int) error {
	if people > t.availableSeats {
		return errors.New("no space")
	}
	t.size = t.availableSeats - people
	return nil
}

type Group struct {
	id      int
	size    int
}

type Booking struct {
	groupID int
	tableID int
}

type SeatingManager struct {
	tables          map[int]Table
	availableTables map[int][]Table
	waitingQueue    []Group
	booking         []*Booking
}

func (m SeatingManager) Arrives(group *Group) {
	booking, err := m.allocateGroup(group)
	if err != nil {
		m.waitingQueue = append(m.waitingQueue, *group)
		return
	}
	m.booking = append(m.booking, booking)
}

func (m SeatingManager) allocateGroup(group *Group) (*Booking, error) {
	for i := group.size; i <= maxSeats; i++ {
		tables, ok := m.availableTables[group.size]
		if ok && len(tables) > 0 {
			tbl := tables[0]
			if err := tbl.Seat(group.size); err != nil {
				return nil, errors.New("could not allocate in this table")
			}
			m.availableTables[i] = tables[1:]
			avlTables, okT := m.availableTables[tbl.availableSeats]
			switch {
			case okT:
				avlTables = append(avlTables, tbl)
			default:
				m.availableTables[tbl.availableSeats] = []Table{tbl}
			}
			return &Booking{groupID: group.id, tableID: tbl.id}, nil
		}
	}
	return nil, errors.New("could not allocate group")
}
