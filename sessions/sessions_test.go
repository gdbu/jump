package sessions

import (
	"os"
	"testing"
)

const (
	testUser1 = "TEST_USER_1"
	testUser2 = "TEST_USER_2"
	testUser3 = "TEST_USER_3"
)

func TestSessions(t *testing.T) {
	var (
		s   *Sessions
		err error
	)

	if err = os.MkdirAll("./test_data", 0744); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./test_data")

	if s, err = New("./test_data", nil); err != nil {
		t.Fatal(err)
	}

	var tu1k, tu1t string
	if tu1k, tu1t, err = s.New(testUser1); err != nil {
		t.Fatal(err)
	}

	var tu2k, tu2t string
	if tu2k, tu2t, err = s.New(testUser2); err != nil {
		t.Fatal(err)
	}

	var tu3k, tu3t string
	if tu3k, tu3t, err = s.New(testUser3); err != nil {
		t.Fatal(err)
	}

	var mu string
	if mu, err = s.Get(tu1k, tu1t); err != nil {
		t.Fatal(err)
	} else if mu != testUser1 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser1, mu)
	}

	if mu, err = s.Get(tu2k, tu2t); err != nil {
		t.Fatal(err)
	} else if mu != testUser2 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser2, mu)
	}

	if mu, err = s.Get(tu3k, tu3t); err != nil {
		t.Fatal(err)
	} else if mu != testUser3 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser3, mu)
	}

	if err = s.Close(); err != nil {
		t.Fatal(err)
	}

	// Re-open sessions from snapshot
	if s, err = New("./test_data", nil); err != nil {
		t.Fatal(err)
	}

	// Make sure the values still match

	if mu, err = s.Get(tu1k, tu1t); err != nil {
		t.Fatal(err)
	} else if mu != testUser1 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser1, mu)
	}

	if mu, err = s.Get(tu2k, tu2t); err != nil {
		t.Fatal(err)
	} else if mu != testUser2 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser2, mu)
	}

	if mu, err = s.Get(tu3k, tu3t); err != nil {
		t.Fatal(err)
	} else if mu != testUser3 {
		t.Fatalf("invalid user match, expected %s and received %s", testUser3, mu)
	}
}
