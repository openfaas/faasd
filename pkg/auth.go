package pkg

import (
	"github.com/sethvargo/go-password/password"
)

func MakeBasicAuthFiles(wd string) error {

	dirErr := EnsureWorkingDir(wd)
	if dirErr != nil {
		return dirErr
	}

	pwdFile := wd + "/basic-auth-password"
	authPassword, err := password.Generate(63, 10, 0, false, true)

	if err != nil {
		return err
	}

	err = MakeFile(pwdFile, authPassword)
	if err != nil {
		return err
	}

	userFile := wd + "/basic-auth-user"
	err = MakeFile(userFile, "admin")
	if err != nil {
		return err
	}

	return nil
}
