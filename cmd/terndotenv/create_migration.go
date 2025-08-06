package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var createMigration = &cobra.Command{
	Use:     "create [opções]",
	Aliases: []string{"c"},
	Short:   "Cria uma nova migration ",
	Long:    `Cria uma nova migration de acordo com as configurações do tern.conf na pasta migrations`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Erro ao carregar o arquivo .env:", err)
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Println("Espera um nome de migração.")
			os.Exit(1)
		}

		migrationName := args[0]
		if migrationName == "" {
			fmt.Println("Nome da migração invalido")
			os.Exit(1)
		}

		fmt.Println("migrationName:", migrationName)

		cmdArgs := []string{
			"new",
			migrationName,
			"-m",
			migrationsPath,
		}

		fmt.Printf("Executando: tern %v\n", cmdArgs)

		execCmd := exec.Command("tern", cmdArgs...)

		output, err := execCmd.CombinedOutput()
		if err != nil {
			fmt.Println("Erro ao executar o comando tern:", err)
			fmt.Println("Saída do comando:", string(output))
			os.Exit(1)
		}

		fmt.Println("Comando executado com sucesso!")
		fmt.Println(string(output))

	},
}
