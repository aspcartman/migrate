package migrate

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

func GetMigrationsFromFolder(path string) ([]Migration, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var mgs migsSlice
	for _, file := range files {
		id, name, err := migrationName(file.Name())
		if err != nil {
			return nil, fmt.Errorf("wrong file name %s: %s", file.Name(), err.Error())
		}

		data, err := ioutil.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		mg, err := NewMigration(id, name, string(data))
		if err != nil {
			return nil, err
		}

		mgs = append(mgs, mg)
	}

	return mgs, nil
}

func migrationName(filename string) (int, string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("len(parts) != 2: %s", parts)
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}

	if strings.HasSuffix(parts[1], ".sql") {
		parts[1] = parts[1][:len(parts[1])-4]
	}

	return int(id), parts[1], nil
}
