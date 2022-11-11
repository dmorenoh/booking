package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSeatingManager_Arrives(t *testing.T) {
	t.Run("Given a some available tablesMap", func(t *testing.T) {
		waitingQueue := make([]*Group, 0)
		tableFor4, table4Err := NewTable(uint8(4))
		require.NoError(t, table4Err)

		tableFor2, table2Err := NewTable(uint8(2))
		require.NoError(t, table2Err)

		tableFor3, table3Err := NewTable(3)
		require.NoError(t, table3Err)

		tables := []*Table{tableFor4, tableFor2, tableFor3}
		seatManager := NewSeatManager(tables, waitingQueue)

		t.Run("When a group arrives and no enough seats for them", func(t *testing.T) {
			group := NewGroup(5)
			seatManager.Arrives(group)

			t.Run("Then group goes to waiting queue", func(t *testing.T) {
				waitingGroups := seatManager.GetWaitingGroups()
				bookings := seatManager.GetBookings()
				assert.Len(t, waitingGroups, 1)
				assert.Equal(t, waitingGroups[0], group)
				assert.Len(t, bookings, 0)

			})
		})

		t.Run("When a group arrives and enough seats for them", func(t *testing.T) {
			newGroup := NewGroup(2)
			seatManager.Arrives(newGroup)

			t.Run("Then group goes to waiting queue", func(t *testing.T) {
				waitingGroups := seatManager.GetWaitingGroups()
				assert.Len(t, waitingGroups, 1)

				bookings := seatManager.GetBookings()
				assert.Len(t, bookings, 1)
				assert.Equal(t, bookings[0].groupID, newGroup.id)
				assert.Equal(t, bookings[0].tableID, tableFor2.id)
			})
		})

		t.Run("When a group arrives and enough seats for them", func(t *testing.T) {
			groupFor4People := NewGroup(4)
			seatManager.Arrives(groupFor4People)

			t.Run("Then group goes to waiting queue", func(t *testing.T) {
				waitingGroups := seatManager.GetWaitingGroups()
				assert.Len(t, waitingGroups, 1)

				bookings := seatManager.GetBookings()
				assert.Len(t, bookings, 2)
				var storedBooking *Booking
				for _, booking := range bookings {
					if booking.groupID == groupFor4People.id {
						storedBooking = booking
					}
				}
				assert.NotNil(t, storedBooking)
				assert.Equal(t, storedBooking.groupID, groupFor4People.id)
				assert.Equal(t, storedBooking.tableID, tableFor4.id)
			})

			t.Run("When locating group", func(t *testing.T) {
				tableFound, err := seatManager.Locate(groupFor4People.id)
				require.NoError(t, err)

				t.Run("Then return table assigned for that group", func(t *testing.T) {
					assert.Equal(t, tableFound.id, tableFor4.id)
				})
			})

			t.Run("When group leaves", func(t *testing.T) {
				err := seatManager.Leaves(groupFor4People.id)
				require.NoError(t, err)

				t.Run("Then return table assigned for that group", func(t *testing.T) {
					tableFound, err2 := seatManager.Locate(groupFor4People.id)
					assert.NoError(t, err2)
					assert.Equal(t, tableFound.availableSeats, tableFound.capacity)
				})
			})
		})

	})
}
