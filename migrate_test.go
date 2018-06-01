package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	if err := execute("docker", "run", "-d", "--rm", "-p", "7733:5432", "--name", "migration_testing_postgres", "postgres:10"); err != nil {
		os.Exit(-1)
	}
	retCode := m.Run()
	if err := execute("docker", "stop", "migration_testing_postgres"); err != nil {
		os.Exit(-1)
	}
	os.Exit(retCode)
}

func TestSimple(t *testing.T) {
	if err := execute("go", "run", "cmd/main.go", "--addr", "localhost:7733", "tests/simple"); err != nil {
		t.Error(err)
	}
}

func execute(cmd string, args ...interface{}) error {
	var argStr []string
	for _, a := range args {
		argStr = append(argStr, fmt.Sprint(a))
	}
	c := exec.Command(cmd, argStr...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
