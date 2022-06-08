package tools

import "github.com/jmoiron/sqlx"

func ReadData(filepath string) error {

	data, err := ReadIn(2, filepath) // 2 dimensions hardcoded
	if err != nil {
		return err
	}

	for pow := 1.0; pow < 3.5; pow += .5 {
		err = MainSolve(data, filepath, pow, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadPGData(db *sqlx.DB, query string) (string, error) {
	return "", nil
}
