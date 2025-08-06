package main

import (
	"github.com/spf13/cobra"
)

// migrateCmd representa o comando migrate
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Executa as migrações do banco de dados (tern migrate)",
	Long:  `Executa todas as migrações pendentes no banco de dados. É um wrapper para o comando 'tern migrate'.`,
	Run: func(cmd *cobra.Command, args []string) {
		executeTernCommand("migrate")
	},
}
