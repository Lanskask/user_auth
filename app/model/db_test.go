package model

import (
	"strconv"
	"testing"
)

func AfterEach(t *testing.T) {
	testDB, err := NewDB(NewDefaultDBConfig())

	if err != nil {
		t.Errorf("Error creating DB: %s", err)
	}

	if err := testDB.TruncateTable(); err != nil {
		t.Errorf("Error truncating table: %s", err)
		return
	}
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		want    IDB
		wantErr bool
	}{
		{
			name:    "Simple test",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDB(NewDefaultDBConfig())

			if (err != nil) != tt.wantErr {
				t.Errorf("NewDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}

func TestDB_CreateUser(t *testing.T) {
	defer AfterEach(t)

	tests := []struct {
		name      string
		inputUser *User
		wantErr   bool
	}{
		{
			name:      "First simple test",
			inputUser: &User{Name: "Peter", Email: "peter@gmail.com", Password: "asfafasf"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		testDB, err := NewDB(NewDefaultDBConfig())

		if (err != nil) != tt.wantErr {
			t.Errorf("Err creating DB: error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		t.Run(tt.name, func(t *testing.T) {
			gotID, err := testDB.CreateUser(tt.inputUser)
			if err != nil {
				t.Errorf("Err creating a new User: error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := testDB.DeleteUser(gotID); err != nil {
				t.Errorf("Err deleting created user: %s", err)
			}
		})

	}
}

func TestDB_GetUser(t *testing.T) {
	defer AfterEach(t)

	tests := []struct {
		name      string
		inputUser *User
		wantErr   bool
	}{
		{
			name:      "First simple test",
			inputUser: &User{Name: "Peter", Email: "peter@gmail.com", Password: "asfafasf"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		testDB, err := NewDB(NewDefaultDBConfig())

		if err != nil {
			t.Errorf("Err creating DB: error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		t.Run(tt.name, func(t *testing.T) {
			createdUserID, err := testDB.CreateUser(tt.inputUser)
			if err != nil {
				t.Errorf("Err creating a new User: error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotUser, err := testDB.GetUser(createdUserID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !CompareUsers(tt.inputUser, &gotUser) {
				t.Errorf("GetUser() createdUserID = %v, expectedUser %v", createdUserID, gotUser)
			}

			if err := testDB.DeleteUser(strconv.Itoa(int(gotUser.ID))); err != nil {
				t.Errorf("Err deleting created user: %s", err)
			}
		})
	}
}

func TestDB_DeleteUser(t *testing.T) {
	defer AfterEach(t)

	tests := []struct {
		name    string
		inputId string
		wantErr bool
	}{
		{
			name:    "First simple test",
			inputId: "2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		testDB, err := NewDB(NewDefaultDBConfig())

		if (err != nil) != tt.wantErr {
			t.Errorf("Err creating DB: error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := testDB.DeleteUser(tt.inputId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func CompareUsers(a, b *User) bool {
	return a.Name == b.Name && a.Email == b.Email && a.Password == b.Password
}
