package permissions

import (
	"os"
	"testing"

	"github.com/Hatch1fy/errors"
)

const (
	testErrCannot = errors.Error("group not allowed to perform action they should be able to")
	testErrCan    = errors.Error("group allowed to perform action they should not be able to")
)

const (
	testUser1 = "TEST_USER_1"
	testUser2 = "TEST_USER_2"
	testUser3 = "TEST_USER_3"
)

func TestPermissions(t *testing.T) {
	var (
		p   *Permissions
		err error
	)

	if err = os.MkdirAll("./_testdata", 0744); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./_testdata")

	if p, err = New("./_testdata"); err != nil {
		t.Fatal(err)
	}

	if err = p.SetPermissions("posts", "users", ActionRead); err != nil {
		t.Fatal(err)
	}

	if err = p.SetPermissions("posts", "admins", ActionRead|ActionWrite); err != nil {
		t.Fatal(err)
	}

	if err = p.SetPermissions("posts", "writers", ActionWrite); err != nil {
		t.Fatal(err)
	}

	if err = p.AddGroup(testUser1, "users"); err != nil {
		t.Fatal(err)
	}

	if err = p.AddGroup(testUser2, "admins"); err != nil {
		t.Fatal(err)
	}

	if err = p.AddGroup(testUser3, "writers"); err != nil {
		t.Fatal(err)
	}

	testPerms(p, t)

	if err = p.Close(); err != nil {
		t.Error(err)
	}

	if p, err = New("./_testdata"); err != nil {
		t.Fatal(err)
	}

	testPerms(p, t)

	if err = p.SetPermissions("posts", "writers", ActionDelete); err != nil {
		t.Fatal(err)
	}

	if !p.Can(testUser3, "posts", ActionDelete) {
		t.Fatal(testErrCannot)
	}
}

func testPerms(p *Permissions, t *testing.T) {
	if !p.Can(testUser1, "posts", ActionRead) {
		t.Fatal(testErrCannot)
	}

	if p.Can(testUser1, "posts", ActionWrite) {
		t.Fatal(testErrCan)
	}

	if !p.Can(testUser2, "posts", ActionRead) {
		t.Fatal(testErrCannot)
	}

	if !p.Can(testUser2, "posts", ActionWrite) {
		t.Fatal(testErrCannot)
	}

	if p.Can(testUser3, "posts", ActionRead) {
		t.Fatal(testErrCan)
	}

	if !p.Can(testUser3, "posts", ActionWrite) {
		t.Fatal(testErrCannot)
	}

	if p.Can(testUser3, "posts", ActionDelete) {
		t.Fatal(testErrCan)
	}
}
