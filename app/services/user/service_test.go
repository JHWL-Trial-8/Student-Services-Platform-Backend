package user

import (
	"testing"

	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	testutils.TestSuite
	service *Service
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()
	
	// Create user service
	suite.service = NewService(suite.DB)
}

func (suite *UserServiceTestSuite) TestGetUserInfomationByID_Success() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("getuser@example.com", "Get User", db.RoleStudent)
	
	// Act
	user, err := suite.service.GetUserInfomationByID(string(testUser.ID))

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, uint(user.Id))
	assert.Equal(suite.T(), testUser.Email, user.Email)
	assert.Equal(suite.T(), testUser.Name, user.Name)
	assert.Equal(suite.T(), db.Role(user.Role), testUser.Role)
	assert.Equal(suite.T(), testUser.Phone, user.Phone)
	assert.Equal(suite.T(), testUser.Dept, user.Dept)
	assert.Equal(suite.T(), testUser.IsActive, user.IsActive)
	assert.Equal(suite.T(), testUser.AllowEmail, user.AllowEmail)
}

func (suite *UserServiceTestSuite) TestGetUserInfomationByID_UserNotFound() {
	// Act
	user, err := suite.service.GetUserInfomationByID("999999")

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserServiceTestSuite) TestGetUserInfomationByID_InvalidID() {
	// Act
	user, err := suite.service.GetUserInfomationByID("invalid-id")

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_Success() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("update@example.com", "Update User", db.RoleStudent)
	
	newEmail := "updated@example.com"
	newName := "Updated Name"
	newPhone := "9876543210"
	newDept := "Updated Department"
	newAllowEmail := false

	updateFields := UpdateFields{
		Email:      &newEmail,
		Name:       &newName,
		Phone:      &newPhone,
		Dept:       &newDept,
		AllowEmail: &newAllowEmail,
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(testUser.ID, updateFields)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, uint(user.Id))
	assert.Equal(suite.T(), newEmail, user.Email)
	assert.Equal(suite.T(), newName, user.Name)
	assert.Equal(suite.T(), &newPhone, user.Phone)
	assert.Equal(suite.T(), &newDept, user.Dept)
	assert.Equal(suite.T(), newAllowEmail, user.AllowEmail)

	// Verify user was actually updated in the database
	var dbUser db.User
	err = suite.DB.First(&dbUser, testUser.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), newEmail, dbUser.Email)
	assert.Equal(suite.T(), newName, dbUser.Name)
	assert.Equal(suite.T(), &newPhone, dbUser.Phone)
	assert.Equal(suite.T(), &newDept, dbUser.Dept)
	assert.Equal(suite.T(), newAllowEmail, dbUser.AllowEmail)
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_PartialUpdate() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("partial@example.com", "Partial User", db.RoleStudent)
	
	newName := "Partially Updated Name"
	updateFields := UpdateFields{
		Name: &newName, // Only update name
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(testUser.ID, updateFields)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, uint(user.Id))
	assert.Equal(suite.T(), testUser.Email, user.Email) // Email should remain unchanged
	assert.Equal(suite.T(), newName, user.Name)         // Name should be updated
	assert.Equal(suite.T(), testUser.Phone, user.Phone) // Phone should remain unchanged
	assert.Equal(suite.T(), testUser.Dept, user.Dept)   // Dept should remain unchanged
	assert.Equal(suite.T(), testUser.AllowEmail, user.AllowEmail) // AllowEmail should remain unchanged
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_ClearPhone() {
	// Arrange - create a user with phone
	testUser := suite.CreateTestUser("clearphone@example.com", "Clear Phone User", db.RoleStudent)
	
	// Set phone in database
	phone := "1234567890"
	suite.DB.Model(&testUser).Update("phone", &phone)
	
	emptyPhone := ""
	updateFields := UpdateFields{
		Email: &testUser.Email, // Email is required
		Phone: &emptyPhone,     // Clear phone
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(testUser.ID, updateFields)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, uint(user.Id))
	assert.Nil(suite.T(), user.Phone) // Phone should be nil (cleared)

	// Verify phone was actually cleared in the database
	var dbUser db.User
	err = suite.DB.First(&dbUser, testUser.ID).Error
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), dbUser.Phone)
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_ClearDept() {
	// Arrange - create a user with dept
	testUser := suite.CreateTestUser("cleardept@example.com", "Clear Dept User", db.RoleStudent)
	
	// Set dept in database
	dept := "Computer Science"
	suite.DB.Model(&testUser).Update("dept", &dept)
	
	emptyDept := ""
	updateFields := UpdateFields{
		Email: &testUser.Email, // Email is required
		Dept:  &emptyDept,      // Clear dept
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(testUser.ID, updateFields)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), testUser.ID, uint(user.Id))
	assert.Nil(suite.T(), user.Dept) // Dept should be nil (cleared)

	// Verify dept was actually cleared in the database
	var dbUser db.User
	err = suite.DB.First(&dbUser, testUser.ID).Error
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), dbUser.Dept)
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_EmptyEmail() {
	// Arrange - create a user
	testUser := suite.CreateTestUser("emptyemail@example.com", "Empty Email User", db.RoleStudent)
	
	emptyEmail := ""
	updateFields := UpdateFields{
		Email: &emptyEmail, // Empty email
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(testUser.ID, updateFields)

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "email 为空")
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_EmailTaken() {
	// Arrange - create two users
	user1 := suite.CreateTestUser("user1@example.com", "User 1", db.RoleStudent)
	user2 := suite.CreateTestUser("user2@example.com", "User 2", db.RoleStudent)
	
	// Try to update user1's email to user2's email
	updateFields := UpdateFields{
		Email: &user2.Email, // Email already taken by user2
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(user1.ID, updateFields)

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), &ErrEmailTaken{}, err)
	assert.Contains(suite.T(), err.Error(), user2.Email)
}

func (suite *UserServiceTestSuite) TestUpdateUserInfomationByID_UserNotFound() {
	// Arrange
	email := "nonexistent@example.com"
	updateFields := UpdateFields{
		Email: &email,
	}

	// Act
	user, err := suite.service.UpdateUserInfomationByID(999999, updateFields)

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func TestUserService(t *testing.T) {
	testutils.RunTestSuite(t, &UserServiceTestSuite{})
}