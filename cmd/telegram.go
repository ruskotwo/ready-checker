package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ruskotwo/ready-checker/cmd/factory"
	"github.com/spf13/cobra"
)

var telegramCmd = &cobra.Command{
	Use:   "telegram",
	Short: "Just telegram bot for simple test",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start telegram bot")

		bot, cleaner1, err := factory.InitTelegramBot()
		defer func() {
			if cleaner1 != nil {
				cleaner1()
			}
		}()
		if err != nil {
			log.Printf("error initialize bot. %v", err)
			return
		}

		bot.Start()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	},
}

func init() {
	rootCmd.AddCommand(telegramCmd)
}
