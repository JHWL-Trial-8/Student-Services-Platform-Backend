package db

import (
	"testing"

	"student-services-platform-backend/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UsersRepoTestSuite struct {
	testutils.TestSuite
}

func (suite *UsersRepoTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()
}

func (suite *UsersRepoTestSuite) TestGetUserByEmail_Success() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("repo@example.com", "Repo User", RoleStudent)

	// Act
	user, err := GetUserByEmail(suite.DB, testUser.Email)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, user.ID)
	assert.Equal(suite.T(), testUser.Email, user.Email)
	assert.Equal(suite.T(), testUser.Name, user.Name)
	assert.Equal(suite.T(), testUser.Role, user.Role)
}

func (suite *UsersRepoTestSuite) TestGetUserByEmail_UserNotFound() {
	// Act
	user, err := GetUserByEmail(suite.DB, "nonexistent@example.com")

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
}

func (suite *UsersRepoTestSuite) TestCreateUser_Success() {
	// Arrange
	user := &User{
		Email:        "newuser@example.com",
		Name:         "New User",
		Role:         RoleStudent,
		Phone:        nil,
		Dept:         nil,
		IsActive:     true,
		AllowEmail:   true,
		PasswordHash: "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPkU2Cslq",
	}

	// Act
	err := CreateUser(suite.DB, user)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotZero(suite.T(), user.ID)
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)

	// Verify user was actually created in the database
	var dbUser User
	err = suite.DB.Where("email = ?", user.Email).First(&dbUser).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Name, dbUser.Name)
	assert.Equal(suite.T(), user.Role, dbUser.Role)
}

func (suite *UsersRepoTestSuite) TestGetUserByID_Success() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("getbyid@example.com", "Get By ID User", RoleStudent)

	// Act
	user, err := GetUserByID(suite.DB, testUser.ID)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, user.ID)
	assert.Equal(suite.T(), testUser.Email, user.Email)
	assert.Equal(suite.T(), testUser.Name, user.Name)
	assert.Equal(suite.T(), testUser.Role, user.Role)
}

func (suite *UsersRepoTestSuite) TestGetUserByID_UserNotFound() {
	// Act
	user, err := GetUserByID(suite.DB, 999999)

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
}

func (suite *UsersRepoTestSuite) TestExistsOtherUserWithEmail_Exists() {
	// Arrange - create two users
	user1 := suite.CreateTestUser("exists1@example.com", "Exists User 1", RoleStudent)
	user2 := suite.CreateTestUser("exists2@example.com", "Exists User 2", RoleStudent)

	// Act - check if user2's email exists for a different user (user1)
	exists, err := ExistsOtherUserWithEmail(suite.DB, user2.Email, user1.ID)

	// Assert
	require.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UsersRepoTestSuite) TestExistsOtherUserWithEmail_NotExists() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("notexists@example.com", "Not Exists User", RoleStudent)

	// Act - check if user's email exists for the same user (should return false)
	exists, err := ExistsOtherUserWithEmail(suite.DB, testUser.Email, testUser.ID)

	// Assert
	require.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *UsersRepoTestSuite) TestExistsOtherUserWithEmail_NonexistentEmail() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("nonexistent@example.com", "Nonexistent Email User", RoleStudent)

	// Act - check if a nonexistent email exists
	exists, err := ExistsOtherUserWithEmail(suite.DB, "completelynonexistent@example.com", testUser.ID)

	// Assert
	require.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *UsersRepoTestSuite) TestUpdateUser_Success() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("update@example.com", "Update User", RoleStudent)
	
	// Update user details
	testUser.Name = "Updated Name"
	testUser.Email = "updated@example.com"
	newPhone := "9876543210"
	testUser.Phone = &newPhone
	newDept := "Updated Department"
	testUser.Dept = &newDept
	testUser.AllowEmail = false

	// Act
	err := UpdateUser(suite.DB, testUser)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), testUser.UpdatedAt, testUser.CreatedAt) // UpdatedAt should have changed

	// Verify user was actually updated in the database
	var dbUser User
	err = suite.DB.First(&dbUser, testUser.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", dbUser.Name)
	assert.Equal(suite.T(), "updated@example.com", dbUser.Email)
	assert.Equal(suite.T(), &newPhone, dbUser.Phone)
	assert.Equal(suite.T(), &newDept, dbUser.Dept)
	assert.Equal(suite.T(), false, dbUser.AllowEmail)
}

func (suite *UsersRepoTestSuite) TestUpdateUser_ClearFields() {
	// Arrange - create a user with phone and dept
	testUser := suite.CreateTestUser("clear@example.com", "Clear User", RoleStudent)
	
	// Set phone and dept in database
	phone := "1234567890"
	dept := "Computer Science"
	suite.DB.Model(&testUser).Updates(map[string]interface{}{
		"phone": &phone,
		"dept":  &dept,
	})
	
	// Reload user to get updated values
	suite.DB.First(&testUser, testUser.ID)
	
	// Clear phone and dept
	testUser.Phone = nil
	testUser.Dept = nil

	// Act
	err := UpdateUser(suite.DB, testUser)

	// Assert
	require.NoError(suite.T(), err)

	// Verify fields were actually cleared in the database
	var dbUser User
	err = suite.DB.First(&dbUser, testUser.ID).Error
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), dbUser.Phone)
	assert.Nil(suite.T(), dbUser.Dept)
}

func TestUsersRepo(t *testing.T) {
	testutils.RunTestSuite(t, &UsersRepoTestSuite{})
}