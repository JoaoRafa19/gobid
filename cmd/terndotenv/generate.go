package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var generate = &cobra.Command{
	Use:     "gen",
	Aliases: []string{"g"},
	Short:   "Gera os arquivos do sqlc",
	Run: func(cmd *cobra.Command, args []string) {

		cmdArgs := []string{
			"generate",
			"./...",
		}

		execCmd := exec.Command("go", cmdArgs...)

		output, err := execCmd.CombinedOutput()
		if err != nil {
			fmt.Println("Erro ao executar o comando generate:", err)
			fmt.Println("Saída do comando:", string(output))
			os.Exit(1)
		}

		fmt.Println("Comando executado com sucesso!")
		fmt.Println(string(output))
	},
}
