# Booking tables

All code expected for `SeatingManager` is actually in `booking.go`. The way how this manages table allocation is considering this:
```go
type SeatingManager interface {
	Arrives(newGroup *Group)
	Leaves(groupID uuid.UUID) error
	Locate(groupID uuid.UUID) (Table, error)
}
```

## Arrives operation
It manages group seats allocation considering available seats by table. In order to make it faster, availability is stored in a hash map data 
structure where `key` is available seats number and `value` points to a list of tableIDs that has this number of available seats. Example:
```go
[
 {2:{tableID_1, tableID_2, tableID_3}},
 {3:{tableID_4, tableID_5}}
 ]
```
Above example means that ``tableID_1, tableID_2, tableID_3`` have 2 seats available, `tableID_4, tableID_5` have 3 seats available, and so on.
Then you have:
```go
func (m *seatingManager) Arrives(newGroup *Group) {
	// this is for managing concurrency and enforcing requests order.
	m.Lock()
	defer m.Unlock()

	m.groupMap[newGroup.id] = *newGroup

	// tries to create a booking for this group by assigning a table for it.
	booking, err := m.allocateGroup(newGroup)
	if err != nil {
		// if group could not be allocated, it goes to a waiting queue.
		m.waitingQueue = append(m.waitingQueue, newGroup)
		return
	}

	m.booking = append(m.booking, booking)
}
```

## Leaves operation
It manages groups leaving action. It also implies to release occupied seats-table by this group.
```go
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
```
All test cases are found on ``booking_test.go``

## Caveats
- Solution works but is not scalable enough. Why? It locks the resource for every ``Arrives(...)``. As this function implies multiple writes operation
it could take longer. 
- Error handling is missed.
- Removing queued groups missed.