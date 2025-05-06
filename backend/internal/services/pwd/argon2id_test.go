package pwd

import "testing"

func TestHashAndSalt(t *testing.T) {
	SetArgon2Configs(64*1024, 1, 2, 16, 32)
	pwd := "plainPassword"

	argon2Hash, err := HashAndSalt(pwd)
	if err != nil {
		t.Errorf("error hashing passowrd: %s", err.Error())

	}

	if argon2Hash == "" {
		t.Errorf("Well this shouldn't be empty")
	}

	t.Logf("argon2Hash: %s", argon2Hash)
}

func TestCheckPassword(t *testing.T) {
	SetArgon2Configs(64*1024, 1, 2, 16, 32)
	pwd := "plainPassword"

	argon2Hash, err := HashAndSalt(pwd)
	if err != nil {
		t.Errorf("error hashing passowrd: %s", err.Error())
	}

	if argon2Hash == "" {
		t.Errorf("Well this shouldn't be empty")
	}

	t.Logf("argon2Hash: %s", argon2Hash)

	if ok, err := CheckPasswordHash("plainPassword", argon2Hash); err != nil {
		t.Errorf("error checking password: %s", err.Error())
	} else if !ok {
		t.Errorf("Incorrect password")
	} else {
		t.Logf("Passowrd is correct")
	}
}
