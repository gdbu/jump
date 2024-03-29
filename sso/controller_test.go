package sso

import (
	"context"
	"os"
	"testing"

	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
)

var testCtx = context.Background()

func TestNew(t *testing.T) {
	var (
		c   *Controller
		err error
	)

	if c, err = testInit(); err != nil {
		t.Fatal(err)
	}
	testTeardown(t, c)
}

func TestController_New(t *testing.T) {
	var (
		c   *Controller
		err error
	)

	if c, err = testInit(); err != nil {
		t.Fatal(err)
	}
	defer testTeardown(t, c)

	var e *Entry
	if e, err = c.New(context.Background(), "user_0"); err != nil {
		t.Fatal(err)
	}

	if e.UserID != "user_0" {
		t.Fatalf("invalid user ID, expected <%s> and received <%s>", "user_0", e.UserID)
	}
}

func TestController_New_replace(t *testing.T) {
	var (
		c   *Controller
		err error
	)

	if c, err = testInit(); err != nil {
		t.Fatal(err)
	}
	defer testTeardown(t, c)

	var e *Entry
	if e, err = c.New(context.Background(), "user_0"); err != nil {
		t.Fatal(err)
	}

	if e.UserID != "user_0" {
		t.Fatalf("invalid user ID, expected <%s> and received <%s>", "user_0", e.UserID)
	}

	if _, err = c.New(context.Background(), "user_0"); err != nil {
		t.Fatal(err)
	}

	if _, err = c.Get(e.ID); err != mojura.ErrEntryNotFound {
		t.Fatalf("invalid error, expected <%v> and received <%v>", mojura.ErrEntryNotFound, err)
	}
}

func TestController_Login(t *testing.T) {
	var (
		c   *Controller
		err error
	)

	if c, err = testInit(); err != nil {
		t.Fatal(err)
	}
	defer testTeardown(t, c)

	var e *Entry
	if e, err = c.New(context.Background(), "user_0"); err != nil {
		t.Fatal(err)
	}

	var userID string
	if userID, err = c.Login(testCtx, e.LoginCode.String()); err != nil {
		t.Fatal(err)
	}

	if userID != e.UserID {
		t.Fatalf("invalid userID, expected <%s> and received <%s>", e.UserID, userID)
	}

	if _, err = c.Get(e.ID); err != mojura.ErrEntryNotFound {
		t.Fatalf("invalid error, expected <%v> and received <%v>", mojura.ErrEntryNotFound, err)
	}
}

func TestController_Login_double_login(t *testing.T) {
	var (
		c   *Controller
		err error
	)

	if c, err = testInit(); err != nil {
		t.Fatal(err)
	}
	defer testTeardown(t, c)

	var e *Entry
	if e, err = c.New(context.Background(), "user_0"); err != nil {
		t.Fatal(err)
	}

	if _, err = c.Login(testCtx, e.LoginCode.String()); err != nil {
		t.Fatal(err)
	}

	if _, err = c.Login(testCtx, e.LoginCode.String()); err != ErrNoCodeMatchFound {
		t.Fatalf("invalid error, expected <%v> and received <%v>", ErrNoCodeMatchFound, err)
	}
}

func testInit() (c *Controller, err error) {
	if err = os.Mkdir("./test_data", 0744); err != nil {
		return
	}

	var opts mojura.Opts
	opts.Dir = "./test_data"
	return New(opts)
}

func testTeardown(t *testing.T, c *Controller) {
	var errs errors.ErrorList
	errs.Push(c.Close())
	errs.Push(os.RemoveAll("./test_data"))
	if err := errs.Err(); err != nil {
		t.Fatalf("error during teardown: %v", err)
	}
}
