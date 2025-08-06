package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// Constantes para os caminhos, facilitando a manutenção
const (
	migrationsPath = "./internal/store/pgstore/migrations"
	configPath     = "./internal/store/pgstore/migrations/tern.conf"
)

var rootCmd = &cobra.Command{
	Use:   "dbtool",
	Short: "Uma ferramenta de CLI para gerenciar migrações de banco de dados com tern.",
	Long: `dbtool é uma aplicação de linha de comando que serve como um wrapper
para o 'tern', facilitando a execução de migrações, rollbacks e verificação de status.`,
}

// Execute adiciona todos os comandos filhos ao comando raiz e define os sinalizadores apropriadamente.
// É chamado por main.main(). Só precisa acontecer uma vez para o rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
}

func executeTernCommand(action string, args ...string) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Erro ao carregar o arquivo .env:", err)
		os.Exit(1)
	}

	cmdArgs := []string{
		action,
		"--migrations",
		migrationsPath,
		"--config",
		configPath,
	}
	cmdArgs = append(cmdArgs, args...)

	fmt.Printf("Executando: tern %v\n", cmdArgs)

	cmd := exec.Command("tern", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Erro ao executar o comando tern:", err)
		fmt.Println("Saída do comando:", string(output))
		os.Exit(1)
	}

	fmt.Println("Comando executado com sucesso!")
	fmt.Println(string(output))
}
