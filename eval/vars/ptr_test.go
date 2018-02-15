package vars

import "testing"

func TestPtrVariable(t *testing.T) {
	i := 10
	variable := NewFromPtr(&i)
	if g := variable.Get(); g != "10" {
		t.Errorf(`Getting ptrVariable returns %v, want "10"`, g)
	}
	err := variable.Set("20")
	if err != nil {
		t.Errorf(`Setting ptrVariable with "20" returns error %v`, err)
	}
	if i != 20 {
		t.Errorf(`Setting ptrVariable didn't change underlying value`)
	}
	err = variable.Set("x")
	if err == nil {
		t.Errorf("Setting ptrVariable with incompatible value returns no error")
	}
}
