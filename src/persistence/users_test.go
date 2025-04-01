package persistence_test

import (
	"testing"

	"github.com/coopstools-homebrew/I-am-zuul/src/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTable(t *testing.T) {
	userTable := persistence.NewUserTable(testDB)

	t.Run("Test adding a user", func(t *testing.T) {
		user := persistence.UserInfo{
			ID:        1,
			LoginName: "Imma_number_one",
			AvatarURL: "https://github.com/test1.png",
			Email:     "test@example.com",
		}

		err := userTable.UpdateUser(&user)
		require.NoError(t, err, "Failed to update user")
	})

	t.Run("Test getting a user", func(t *testing.T) {
		user, err := userTable.GetUserByID(1)
		require.NoError(t, err, "Failed to get user")
		require.Equal(t, user.ID, int32(1))
		require.Equal(t, user.LoginName, "Imma_number_one")
		require.Equal(t, user.AvatarURL, "https://github.com/test1.png")
		require.Equal(t, user.Email, "test@example.com")
	})

	t.Run("Test getting all users", func(t *testing.T) {
		user2 := persistence.UserInfo{
			ID:        2,
			LoginName: "Imma_number_two",
			AvatarURL: "https://github.com/test2.png",
			Email:     "test2@example.com",
		}
		err := userTable.UpdateUser(&user2)
		require.NoError(t, err, "Failed to update user")

		users, err := userTable.GetAllUsers()
		require.NoError(t, err, "Failed to get all users")
		assert.Equal(t, len(users), 2)

		assert.Equal(t, users[0].LoginName, "Imma_number_one")
		assert.Equal(t, users[1].LoginName, "Imma_number_two")
	})
}
