package users

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.
func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		user        *NewUser
		expectError bool
	}{
		{
			"Basic case",
			&NewUser{
				"mail@newuser.com",
				"password",
				"password",
				"firstname",
				"lastname",
			},
			false,
		},
		{
			"Email badly formatted",
			&NewUser{
				"@@@newuser.com",
				"password",
				"password",
				"firstname",
				"lastname",
			},
			true,
		},
		{
			"Non matching password confirmation",
			&NewUser{
				"mail@newuser.com",
				"password",
				"password2",
				"firstname",
				"lastname",
			},
			true,
		},
		{
			"Password less than 6 characters",
			&NewUser{
				"mail@newuser.com",
				"a",
				"a",
				"firstname",
				"lastname",
			},
			true,
		},
		{
			"Empty user",
			&NewUser{},
			true,
		},
	}

	for _, c := range cases {
		anyErr := c.user.Validate()
		if anyErr != nil && !c.expectError {
			t.Errorf("Unexpected error occurred. Got \"%v\" for test [%s]", anyErr, c.name)
		}

		if anyErr == nil && c.expectError {
			t.Errorf("Expected error but received none for test [%s]", c.name)
		}
	}
}

func TestSetPassword(t *testing.T) {

	cases := []struct {
		name     string
		password string
	}{
		{
			"Test hashing works correctly",
			"password",
		},
		{
			"Test hasing a crazy string",
			"adsfjsayu83oi147103985thrwuelijkfdo(*@&^#*(U",
		},
		{
			"Test empty string",
			"",
		},
		{
			"Test super long space string",
			"                                                                                                       ",
		},
	}

	for _, c := range cases {
		user := &User{}
		anyErr := user.SetPassword(c.password)

		if anyErr != nil {
			t.Errorf("Unexpected error occured. Got \"%v\" for test [%s]", anyErr, c.name)
		}

		if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(c.password)); err != nil {
			t.Errorf("Password and hash matching failed for test [%s]", c.name)
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name        string
		user        *NewUser
		expectError bool
	}{
		{
			"Working case",
			&NewUser{
				"asdf@asdf.com",
				"password",
				"password",
				"firstname",
				"lastname",
			},
			false,
		},
		{
			"Case expected to fail validation",
			&NewUser{
				"asdf@asdf.com",
				"password",
				"passwordconffail",
				"firstname",
				"lastname",
			},
			true,
		},
	}

	for _, c := range cases {
		user, err := c.user.ToUser()
		if !c.expectError {
			if err != nil {
				t.Errorf("Unexpected error occured for test [%s]. Received \"%v\"", c.name, err)
			}
			if c.user.FirstName != user.FirstName {
				t.Errorf("First name does not match for test [%s]. Expected \"%s\" but got \"%s\"", c.name, c.user.FirstName, user.FirstName)
			}

			if c.user.LastName != user.LastName {
				t.Errorf("Last name does not match for test [%s]. Expected \"%s\" but got \"%s\"", c.name, c.user.LastName, user.LastName)
			}

			if c.user.Email != user.Email {
				t.Errorf("Email does not match for test [%s]. Expected \"%s\" but got \"%s\"", c.name, c.user.Email, user.Email)
			}
		} else {
			if err == nil {
				t.Errorf("Expected error validation error but got %v", err)
			}
		}
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		name     string
		user     *User
		expected string
	}{
		{
			"Working case",
			&User{
				FirstName: "ABC",
				LastName:  "CDE",
			},
			"ABC CDE",
		},
		{
			"No first name",
			&User{
				LastName: "CDE",
			},
			"CDE",
		},
		{
			"No last name",
			&User{
				FirstName: "ABC",
			},
			"ABC",
		},
		{
			"No first or last name",
			&User{},
			"",
		},
	}

	for _, c := range cases {
		if c.user.FullName() != c.expected {
			t.Errorf("Failure on [%s], expected [%s] but got [%s]", c.name, c.expected, c.user.FullName())
		}
	}
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name     string
		password string
	}{
		{
			"Base case",
			"password",
		},
		{
			"Empty password (never happens)",
			"",
		},
		{
			"Long password random string",
			"asdjkfhslajkhr2uyio3y41o93yr@*^$&@$*^YFBDS",
		},
	}

	for _, c := range cases {
		user := &User{}
		if err := user.SetPassword(c.password); err != nil {
			t.Errorf("Unepected error occured on test [%s], got [%v]", c.name, err)
		}

		if err := user.Authenticate(c.password); err != nil {
			t.Errorf("Unepected error occured on test [%s], got [%v]", c.name, err)
		}

		if err := user.Authenticate(c.password + "randomstufftoaddtopassword"); err == nil {
			t.Errorf("Expected error, but got [%v]", err)
		}
	}
}

func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name         string
		firstName    string
		lastName     string
		newFirstName string
		newLastName  string
	}{
		{
			"Base case",
			"name",
			"Lastname",
			"namenew",
			"namenewlastname",
		},
		{
			"Empty last name",
			"firstname",
			"lastname",
			"newfirstname",
			"",
		},
		{
			"Empty first name",
			"firstname",
			"lastname",
			"",
			"newlastname",
		},
		{
			"Empty new name",
			"firstname",
			"lastname",
			"",
			"",
		},
	}

	for _, c := range cases {
		user := &User{
			FirstName: c.firstName,
			LastName:  c.lastName,
		}

		if err := user.ApplyUpdates(&Updates{FirstName: c.newFirstName, LastName: c.newLastName}); err != nil {
			t.Errorf("Unexpected error occurred in case [%s], got [%v]", c.name, err)
		}

		if c.newFirstName != "" && user.FirstName != c.newFirstName {
			t.Errorf("First name does not match new first name, expected [%s] but got [%s]", c.newFirstName, user.FirstName)
		}

		if c.newFirstName == "" && user.FirstName != c.firstName {
			t.Errorf("First name does not match expected output. Expected [%s], got [%s]", c.firstName, user.FirstName)
		}

		if c.newLastName != "" && user.LastName != c.newLastName {
			t.Errorf("Last name does not match new last name, expected [%s] but got [%s]", c.newLastName, user.LastName)
		}

		if c.newLastName == "" && user.LastName != c.lastName {
			t.Errorf("Last name does not match expected output. Expected [%s], got [%s]", c.lastName, user.LastName)
		}

	}
}
