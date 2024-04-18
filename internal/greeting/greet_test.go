package greeting

import "testing"

func TestGreetOneName(t *testing.T) {
	name := "Espen Askeladd"
	names := []string{name}
	want := "Hello Espen Askeladd"

	got, err := Greet(names)
	if err != nil {
		t.Fatalf("err should be nil, got: %s", err.Error())
	}
	if got != want {
		t.Fatalf("Wanted '%s', got '%s'", want, got)
	}
}

func TestGreetTwoNames(t *testing.T) {
	names := []string{"Per", "Espen Askeladd"}
	want := "Hello Per and Espen Askeladd"

	got, err := Greet(names)
	if err != nil {
		t.Fatalf("err should be nil, got: %s", err.Error())
	}
	if got != want {
		t.Fatalf("Wanted '%s', got '%s'", want, got)
	}
}

func TestGreetThreeNames(t *testing.T) {
	names := []string{"Per", "Pål", "Espen Askeladd"}
	want := "Hello Per, Pål and Espen Askeladd"

	got, err := Greet(names)
	if err != nil {
		t.Fatalf("err should be nil, got: %s", err.Error())
	}
	if got != want {
		t.Fatalf("Wanted '%s', got '%s'", want, got)
	}
}

func TestGreetManyNames(t *testing.T) {
	names := []string{"Hans", "Grete", "Per", "Pål", "Espen Askeladd"}
	want := "Hello Hans, Grete, Per, Pål and Espen Askeladd"

	got, err := Greet(names)
	if err != nil {
		t.Fatalf("err should be nil, got: %s", err.Error())
	}
	if got != want {
		t.Fatalf("Wanted '%s', got '%s'", want, got)
	}
}

func TestGreetNone(t *testing.T) {
	names := []string{}

	got, err := Greet(names)
	if err == nil {
		t.Fatalf("Expected returned error, got nil")
	}
	if got != "" {
		t.Fatalf("Expected empty string as return value with the error, got '%s'", got)
	}
}
