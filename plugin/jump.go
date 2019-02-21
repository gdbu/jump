package main

import (
	"log"

	"gitlab.com/itsMontoya/permissions"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/jump"
	"github.com/missionMeteora/journaler"
)

const (
	// ErrResourceIDIsEmpty is returned when resource id is expected but not found withing a permissions hook
	ErrResourceIDIsEmpty = errors.Error("resourceID is empty")
	// ErrInvalidSetUserArguments is returned when an invalid number of set user arguments are provided
	ErrInvalidSetUserArguments = errors.Error("invalid set user arguments, expecting no or one argument (redirectOnFail, optional)")
	// ErrInvalidCheckPermissionsArguments is returned when an invalid number of check permissions arguments are provided
	ErrInvalidCheckPermissionsArguments = errors.Error("invalid check permissions arguments, expecting two arguments (resource name and parameter key)")
	// ErrInvalidGrantPermissionsArguments is returned when an invalid number of grant permissions arguments are provided
	ErrInvalidGrantPermissionsArguments = errors.Error("invalid check permissions arguments, expecting three arguments (resource name, user actions, admin actions)")
)

const (
	permR   = permissions.ActionRead
	permRW  = permissions.ActionRead | permissions.ActionWrite
	permRWD = permissions.ActionRead | permissions.ActionWrite | permissions.ActionDelete
	permWD  = permissions.ActionWrite | permissions.ActionDelete
)

var p plugin

// Init will initialize a plugin
func init() {
	var err error
	dir := "./data"
	p.out = journaler.New("Auth")
	if p.jump, err = jump.New(dir); err != nil {
		log.Fatalf("error initializing jump: %v", err)
	}

	if err = p.seed(); err != nil {
		log.Fatalf("error seeding users: %v", err)
	}
}

// Backend will return the plugin's backend
func Backend() *jump.Jump {
	return p.jump
}

type plugin struct {
	out  *journaler.Journaler
	jump *jump.Jump
}

func (p *plugin) seed() (err error) {
	var apiKey string
	if _, err = p.jump.Users().Get("00000000"); err == nil {
		return
	}

	if err = p.jump.SetPermission("users", "", "users", permissions.ActionNone, permRWD); err != nil {
		return
	}

	if _, apiKey, err = p.jump.CreateUser("admin", "admin", "users", "admins"); err != nil {
		return
	}

	p.out.Success("Successfully created admin with api key of: %s", apiKey)
	return

}
