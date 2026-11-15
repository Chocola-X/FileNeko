package settings

import "testing"

func TestFail2BanValidation(t *testing.T) {
	valid := DefaultFail2Ban()
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, p := range []Fail2Ban{
		{true, 0, 5, 5}, {true, 101, 5, 5}, {true, 5, 0, 5}, {true, 5, 1441, 5}, {true, 5, 5, 0}, {true, 5, 5, 10081},
	} {
		if p.Validate() == nil {
			t.Fatalf("invalid policy accepted: %+v", p)
		}
	}
}
