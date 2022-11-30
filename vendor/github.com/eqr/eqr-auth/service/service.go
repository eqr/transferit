package service

import (
	"fmt"
	"net/rpc"

	"github.com/eqr/eqr-auth/auth"
)

type InternalServer struct {
}

type Response struct {
	Message string
}

type CreateUserRequest struct {
	Login    string
	Password string
}

type DeleteUserRequest struct {
	Id uint64
}

type ListUsersRequest struct {
}

type PingRequest struct {
}

type ListUsersResult struct {
	Login string
	Id    uint64
}

func (r ListUsersResult) String() string {
	return fmt.Sprintf("Id: %d, Login: %s", r.Id, r.Login)
}

type ListUsersResponse struct {
	Users []ListUsersResult
}

type ListUsersHandler struct {
	Service auth.LoginService
}

type CreateUserHandler struct {
	Service auth.LoginService
}

type DeleteUserHandler struct {
	Service auth.LoginService
}

type PingHandler struct {
}

func SetupRpc(srv auth.LoginService) error {
	if err := rpc.Register(&PingHandler{}); err != nil {
		return fmt.Errorf("cannot register Ping handler: %w", err)
	}

	if err := rpc.Register(&ListUsersHandler{Service: srv}); err != nil {
		return fmt.Errorf("cannot register ListUsers handler: %w", err)
	}

	if err := rpc.Register(&CreateUserHandler{Service: srv}); err != nil {
		return fmt.Errorf("cannot register CreateUser handler: %w", err)
	}

	if err := rpc.Register(&DeleteUserHandler{Service: srv}); err != nil {
		return fmt.Errorf("cannot register CreateUser handler: %w", err)
	}

	return nil
}

func (h *ListUsersHandler) Execute(req ListUsersRequest, res *ListUsersResponse) error {
	users, err := h.Service.ListUsers()
	if err != nil {
		return err
	}

	res.Users = make([]ListUsersResult, 0, len(users))
	for _, u := range users {
		respUser := ListUsersResult{Id: u.Id, Login: u.Login}
		res.Users = append(res.Users, respUser)
	}

	return nil
}

func (h *PingHandler) Execute(req PingRequest, res *Response) error {
	res.Message = "ping succeeded"
	return nil
}

func (h *CreateUserHandler) Execute(req CreateUserRequest, res *Response) error {
	id, err := h.Service.CreateUser(req.Login, req.Password)
	if err != nil {
		res.Message = err.Error()
		return err
	} else {
		res.Message = fmt.Sprintf("created user %d - %v", id, req.Login)
	}

	return nil
}

func (h *DeleteUserHandler) Execute(req DeleteUserRequest, res *Response) error {
	err := h.Service.DeleteUser(req.Id)
	if err != nil {
		res.Message = err.Error()
		return err
	} else {
		res.Message = fmt.Sprintf("deleted user: %d", req.Id)
	}

	return nil
}
