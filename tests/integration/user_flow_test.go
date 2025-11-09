package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRegistration tests the user registration flow
func TestUserRegistration(t *testing.T) {
	WaitForAllServices(t)

	timestamp := GenerateTimestamp()
	email := fmt.Sprintf("newuser%d@example.com", timestamp)

	registerPayload := map[string]string{
		"email":    email,
		"password": "SecurePass@123",
		"name":     "New User",
		"phone":    "+5511988887777",
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", registerPayload, "")
	defer resp.Body.Close()

	// Verify status code
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Verify response structure
	apiResp := AssertResponseSuccess(t, resp, http.StatusCreated)

	// Parse user data
	var userData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &userData)
	require.NoError(t, err)

	// Verify user fields
	assert.NotEmpty(t, userData["id"], "User ID should be present")
	assert.Equal(t, email, userData["email"], "Email should match")
	assert.Equal(t, "New User", userData["name"], "Name should match")
	assert.Equal(t, "+5511988887777", userData["phone"], "Phone should match")
	assert.NotEmpty(t, userData["created_at"], "Created at should be present")
	assert.NotEmpty(t, userData["updated_at"], "Updated at should be present")

	// Verify password is NOT returned
	_, hasPassword := userData["password"]
	assert.False(t, hasPassword, "Password should not be in response")
}

// TestUserRegistrationDuplicateEmail tests duplicate email validation
func TestUserRegistrationDuplicateEmail(t *testing.T) {
	WaitForAllServices(t)

	// Create first user
	user := CreateTestUser(t)

	// Try to register with same email
	registerPayload := map[string]string{
		"email":    user.Email,
		"password": "AnotherPass@123",
		"name":     "Another User",
		"phone":    "+5511977776666",
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", registerPayload, "")
	defer resp.Body.Close()

	// Should return conflict error
	AssertResponseError(t, resp, http.StatusConflict)
}

// TestUserRegistrationInvalidData tests validation of registration data
func TestUserRegistrationInvalidData(t *testing.T) {
	WaitForAllServices(t)

	testCases := []struct {
		name    string
		payload map[string]string
		status  int
	}{
		{
			name: "missing email",
			payload: map[string]string{
				"password": "Pass@123",
				"name":     "Test User",
				"phone":    "+5511999999999",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			payload: map[string]string{
				"email":    "invalid-email",
				"password": "Pass@123",
				"name":     "Test User",
				"phone":    "+5511999999999",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "weak password",
			payload: map[string]string{
				"email":    fmt.Sprintf("test%d@example.com", GenerateTimestamp()),
				"password": "123",
				"name":     "Test User",
				"phone":    "+5511999999999",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "missing name",
			payload: map[string]string{
				"email":    fmt.Sprintf("test%d@example.com", GenerateTimestamp()),
				"password": "Pass@123",
				"phone":    "+5511999999999",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", tc.payload, "")
			defer resp.Body.Close()

			AssertResponseError(t, resp, tc.status)
		})
	}
}

// TestUserLogin tests successful login
func TestUserLogin(t *testing.T) {
	WaitForAllServices(t)

	// Create user
	user := CreateTestUser(t)

	// Login with correct credentials
	loginPayload := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/login", loginPayload, "")
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	// Parse token data
	var loginData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &loginData)
	require.NoError(t, err)

	// Verify token is present
	token, ok := loginData["token"].(string)
	assert.True(t, ok, "Token should be present")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Verify user data
	assert.Equal(t, user.Email, loginData["email"])
	assert.Equal(t, user.Name, loginData["name"])
}

// TestUserLoginInvalidCredentials tests login with invalid credentials
func TestUserLoginInvalidCredentials(t *testing.T) {
	WaitForAllServices(t)

	// Create user
	user := CreateTestUser(t)

	testCases := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "wrong password",
			email:    user.Email,
			password: "WrongPassword@123",
		},
		{
			name:     "non-existent user",
			email:    "nonexistent@example.com",
			password: "AnyPassword@123",
		},
		{
			name:     "empty password",
			email:    user.Email,
			password: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loginPayload := map[string]string{
				"email":    tc.email,
				"password": tc.password,
			}

			resp := DoRequest(t, "POST", GatewayURL+"/api/v1/login", loginPayload, "")
			defer resp.Body.Close()

			AssertResponseError(t, resp, http.StatusUnauthorized)
		})
	}
}

// TestUserProfile tests fetching user profile
func TestUserProfile(t *testing.T) {
	WaitForAllServices(t)

	// Create user
	user := CreateTestUser(t)

	// Get profile
	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, user.Token)
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	// Parse profile data
	var profileData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &profileData)
	require.NoError(t, err)

	// Verify profile fields
	assert.Equal(t, user.Email, profileData["email"])
	assert.Equal(t, user.Name, profileData["name"])
	assert.Equal(t, user.Phone, profileData["phone"])
	assert.NotEmpty(t, profileData["id"])
	assert.NotEmpty(t, profileData["created_at"])

	// Verify password is NOT returned
	_, hasPassword := profileData["password"]
	assert.False(t, hasPassword, "Password should not be in profile response")
}

// TestUserProfileWithoutAuth tests profile access without authentication
func TestUserProfileWithoutAuth(t *testing.T) {
	WaitForAllServices(t)

	// Try to get profile without token
	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, "")
	defer resp.Body.Close()

	// Should return unauthorized
	AssertResponseError(t, resp, http.StatusUnauthorized)
}

// TestUserProfileWithInvalidToken tests profile access with invalid token
func TestUserProfileWithInvalidToken(t *testing.T) {
	WaitForAllServices(t)

	// Try with invalid token
	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, "invalid.token.here")
	defer resp.Body.Close()

	// Should return unauthorized
	AssertResponseError(t, resp, http.StatusUnauthorized)
}

// TestUpdateUserProfile tests updating user profile
func TestUpdateUserProfile(t *testing.T) {
	WaitForAllServices(t)

	// Create user
	user := CreateTestUser(t)

	// Update profile
	updatePayload := map[string]string{
		"name":  "Updated Name",
		"phone": "+5511966665555",
	}

	resp := DoRequest(t, "PUT", GatewayURL+"/api/v1/profile", updatePayload, user.Token)
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	// Parse updated data
	var updatedData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &updatedData)
	require.NoError(t, err)

	// Verify updated fields
	assert.Equal(t, "Updated Name", updatedData["name"])
	assert.Equal(t, "+5511966665555", updatedData["phone"])
	assert.Equal(t, user.Email, updatedData["email"], "Email should not change")

	// Verify by fetching profile again
	resp2 := DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, user.Token)
	defer resp2.Body.Close()

	apiResp2 := AssertResponseSuccess(t, resp2, http.StatusOK)

	var profileData map[string]interface{}
	err = json.Unmarshal(apiResp2.Data, &profileData)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", profileData["name"])
	assert.Equal(t, "+5511966665555", profileData["phone"])
}

// TestUpdateUserProfileWithoutAuth tests update without authentication
func TestUpdateUserProfileWithoutAuth(t *testing.T) {
	WaitForAllServices(t)

	updatePayload := map[string]string{
		"name": "Hacker Name",
	}

	resp := DoRequest(t, "PUT", GatewayURL+"/api/v1/profile", updatePayload, "")
	defer resp.Body.Close()

	// Should return unauthorized
	AssertResponseError(t, resp, http.StatusUnauthorized)
}

// TestCompleteUserFlow tests the complete user journey
func TestCompleteUserFlow(t *testing.T) {
	WaitForAllServices(t)

	timestamp := GenerateTimestamp()
	email := fmt.Sprintf("fullflow%d@example.com", timestamp)
	password := "FullFlow@123"

	// Step 1: Register
	t.Log("Step 1: Registering new user")
	registerPayload := map[string]string{
		"email":    email,
		"password": password,
		"name":     "Full Flow User",
		"phone":    "+5511955554444",
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", registerPayload, "")
	AssertResponseSuccess(t, resp, http.StatusCreated)
	resp.Body.Close()

	// Step 2: Login
	t.Log("Step 2: Logging in")
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp = DoRequest(t, "POST", GatewayURL+"/api/v1/login", loginPayload, "")
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)
	token := ParseToken(t, apiResp.Data)
	resp.Body.Close()

	// Step 3: Get profile
	t.Log("Step 3: Fetching profile")
	resp = DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	var profileData map[string]interface{}
	json.Unmarshal(apiResp.Data, &profileData)
	assert.Equal(t, email, profileData["email"])
	resp.Body.Close()

	// Step 4: Update profile
	t.Log("Step 4: Updating profile")
	updatePayload := map[string]string{
		"name":  "Updated Flow User",
		"phone": "+5511944443333",
	}

	resp = DoRequest(t, "PUT", GatewayURL+"/api/v1/profile", updatePayload, token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	var updatedData map[string]interface{}
	json.Unmarshal(apiResp.Data, &updatedData)
	assert.Equal(t, "Updated Flow User", updatedData["name"])
	assert.Equal(t, "+5511944443333", updatedData["phone"])
	resp.Body.Close()

	// Step 5: Verify changes persisted
	t.Log("Step 5: Verifying changes persisted")
	resp = DoRequest(t, "GET", GatewayURL+"/api/v1/profile", nil, token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	json.Unmarshal(apiResp.Data, &profileData)
	assert.Equal(t, "Updated Flow User", profileData["name"])
	assert.Equal(t, "+5511944443333", profileData["phone"])
	resp.Body.Close()

	t.Log("Complete user flow test passed!")
}
