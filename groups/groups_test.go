package groups

import (
	"fmt"
	"os"
	"testing"

	"github.com/mojura/mojura"
)

const (
	testUser1 = "TEST_USER_1"
	testUser2 = "TEST_USER_2"
	testUser3 = "TEST_USER_3"
)

func TestGroups(t *testing.T) {
	var (
		g   *Groups
		err error
	)

	if err = os.MkdirAll("./test_data", 0744); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./test_data")

	var opts mojura.Opts
	opts.Dir = "./test_data"
	if g, err = New(opts); err != nil {
		t.Fatal(err)
	}

	if _, err = g.AddGroups(testUser1, "users"); err != nil {
		t.Fatal(err)
	}

	if _, err = g.AddGroups(testUser2, "admins"); err != nil {
		t.Fatal(err)
	}

	if _, err = g.AddGroups(testUser3, "writers"); err != nil {
		t.Fatal(err)
	}

	testGroups(g, t)

	if err = g.Close(); err != nil {
		t.Error(err)
	}

	if g, err = New(opts); err != nil {
		t.Fatal(err)
	}

	testGroups(g, t)
}

func testGroups(g *Groups, t *testing.T) {
	var err error
	if err = testGroup(g, testUser1, "users", true); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser1, "admins", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser1, "writers", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser2, "users", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser2, "admins", true); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser2, "writers", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser3, "users", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser3, "admins", false); err != nil {
		t.Fatal(err)
		return
	}

	if err = testGroup(g, testUser3, "writers", true); err != nil {
		t.Fatal(err)
		return
	}
}

func testGroup(g *Groups, userID, group string, shouldMatch bool) (err error) {
	var hasGroup bool
	hasGroup, err = g.HasGroup(userID, group)
	switch {
	case err != nil:
		return
	case hasGroup && !shouldMatch:
		return fmt.Errorf("user <%s> is in group <%s> when not expected to", userID, group)
	case !hasGroup && shouldMatch:
		return fmt.Errorf("user <%s> is not in group <%s> when expected to", userID, group)
	}

	return
}
