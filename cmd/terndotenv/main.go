package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os/exec"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	cmd := exec.Command(
		"tern",
		"migrate",
		"--migrations",
		"./internal/store/pgstore/migrations",
		"--config",
		"./internal/store/pgstore/migrations/tern.conf",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Command execution error:", err)
		fmt.Println(string(output))
		panic(err)
	}

	fmt.Println("Command execution success:", string(output))

}
