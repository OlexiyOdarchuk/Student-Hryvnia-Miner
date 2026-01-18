package backend

import "errors"

func ChangePassword(oldPass, newPass string) error {
	if oldPass != sessionPassword {
		return errors.New("Старий пароль невірний")
	}

	err := SaveStorage(newPass, CurrentStorage)
	if err != nil {
		return err
	}

	sessionPassword = newPass
	return nil
}
