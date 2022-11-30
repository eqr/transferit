package cmd

import (
	"fmt"
	"github.com/eqr/eqr-auth/config"
	"github.com/eqr/eqr-auth/service"
	"log"
	"net/rpc"

	"github.com/spf13/cobra"
)

var ConfigPath string

type contextKey string

var ConfigPathKey contextKey = "configPath"

var UserManagerCmd = &cobra.Command{
	Use:   "users",
	Short: "User management for GoStream",
	Long:  `GoStream`,
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Context().Value(ConfigPathKey)
		cfg := config.InitConfig(path.(string))

		var users []service.ListUsersResult

		if pingInternalService(cfg) {
			// service is online, use internal rpc
			client, err := getRpcClient(cfg)
			if err != nil {
				log.Fatalf(err.Error())
			}
			defer client.Close()

			var request service.ListUsersRequest
			response := new(service.ListUsersResponse)

			err = client.Call("ListUsersHandler.Execute", request, &response)
			if err != nil {
				log.Fatalf("error doing rpc call: %v", err.Error())
				return
			}

			users = response.Users

		} else {
			log.Fatal("cannot ping internal service")
		}

		if len(users) == 0 {
			fmt.Println("no users exist")
			return
		}

		fmt.Println("list of existing users:")
		for _, user := range users {
			fmt.Println(user)
		}
	},
}

var AddUsersCmd = &cobra.Command{
	Use:   "add",
	Short: "add user to the system",
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Context().Value(ConfigPathKey)
		cfg := config.InitConfig(path.(string))

		if pingInternalService(cfg) {
			// service is online, use internal rpc
			client, err := getRpcClient(cfg)
			if err != nil {
				log.Fatalf(err.Error())
				return
			}
			defer client.Close()

			request := service.CreateUserRequest{
				Login:    Login,
				Password: Password,
			}
			response := new(service.Response)

			err = client.Call("CreateUserHandler.Execute", request, &response)
			if err != nil {
				log.Fatalf("error doing rpc call: %v", err.Error())
				return
			}
		} else {
			log.Fatal("cannot ping internal service")
		}
	},
}

var DeleteUsersCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete user from the system",
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Context().Value(ConfigPathKey)
		cfg := config.InitConfig(path.(string))

		if pingInternalService(cfg) {
			// service is online, use internal rpc
			client, err := getRpcClient(cfg)
			if err != nil {
				log.Fatalf(err.Error())
				return
			}
			defer client.Close()

			request := service.DeleteUserRequest{
				Id: Id,
			}
			response := new(service.Response)

			err = client.Call("DeleteUserHandler.Execute", request, &response)
			if err != nil {
				log.Fatalf("error doing rpc call: %v", err.Error())
				return
			}
		} else {
			log.Fatalf("cannot ping internal service")
		}
	},
}

var Login string
var Password string
var Id uint64

func BuildUserManager() error {
	AddUsersCmd.Flags().StringVarP(&Login, "login", "l", "", "Login of the user to create")
	if err := AddUsersCmd.MarkFlagRequired("login"); err != nil {
		return fmt.Errorf("cannot mark flag 'login' required: %w", err)
	}

	AddUsersCmd.Flags().StringVarP(&Password, "password", "p", "", "Password of the user to create")
	if err := AddUsersCmd.MarkFlagRequired("password"); err != nil {
		return fmt.Errorf("cannot mark flag 'password' required: %w", err)
	}

	DeleteUsersCmd.Flags().Uint64VarP(&Id, "id", "i", 0, "Id of the user to delete")
	if err := DeleteUsersCmd.MarkFlagRequired("id"); err != nil {
		return fmt.Errorf("cannot mark flag 'id' required: %w", err)
	}

	UserManagerCmd.AddCommand(AddUsersCmd)
	UserManagerCmd.AddCommand(DeleteUsersCmd)
	return nil
}

func getRpcClient(cfg *config.Config) (*rpc.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.InternalPort)
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		err = fmt.Errorf("error connecting internal service %s: %s", addr, err.Error())
		return nil, err
	}
	return client, nil
}

func pingInternalService(cfg *config.Config) bool {
	client, err := getRpcClient(cfg)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer client.Close()

	var request service.PingRequest
	response := new(service.Response)

	err = client.Call("PingHandler.Execute", request, &response)
	if err != nil {
		log.Printf("error pinging internal service: %s", err.Error())
		return false
	}

	return true
}
