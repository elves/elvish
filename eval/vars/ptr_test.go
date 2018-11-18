package vars

import "testing"

func TestFromPtr_UsesUnderlyingValue(t *testing.T) {
	i := 10
	variable := FromPtr(&i)
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

func TestFromInit_AllowsAnyType(t *testing.T) {
	variable := FromInit(10)
	err := variable.Set("x")
	if err != nil {
		t.Errorf("Failed to set variable created with FromInit: %v", err)
	}
}
