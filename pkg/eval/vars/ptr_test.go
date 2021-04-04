package vars

import "testing"

func TestFromPtr(t *testing.T) {
	i := 10
	variable := FromPtr(&i)
	if g := variable.Get(); g != 10 {
		t.Errorf(`Get -> %v, want 10`, g)
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

func TestFromInit(t *testing.T) {
	v := FromInit(true)
	if val := v.Get(); val != true {
		t.Errorf("Get returned %v, want true", val)
	}
	if err := v.Set("233"); err != nil {
		t.Errorf("Set errors: %v", err)
	}
	if val := v.Get(); val != "233" {
		t.Errorf(`Get returns %v, want "233"`, val)
	}
}
