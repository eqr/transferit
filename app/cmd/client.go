package cmd

import (
	"log"

	"github.com/eqr/transferit/app/client"
	"github.com/spf13/cobra"
)

var TransferCmd = &cobra.Command{
	Use:   "file",
	Short: "uploads and downloads",
	Long:  `file management`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// command to upload file
var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "uploads a file",
	Long:  `uploads a file`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			log.Fatal("no filename provided")
		}

		fileName := args[0]

		cl, err := client.Connect("localhost:8083")
		if err != nil {
			log.Fatalf("cannot connect: %v", err.Error())
		}

		err = cl.Upload(fileName)
		if err != nil {
			log.Fatalf("error uploading file %s: %v", fileName, err.Error())
		}
	},
}

func BuildFileManager() {
	TransferCmd.AddCommand(UploadCmd)
}
