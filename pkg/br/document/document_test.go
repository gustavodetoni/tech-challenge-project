package document

import "testing"

func TestNormalize(t *testing.T) {
	got := Normalize("  529.982.247-25 ")
	if got != "52998224725" {
		t.Fatalf("expected %q, got %q", "52998224725", got)
	}
}

func TestIsValidCPF(t *testing.T) {
	if !IsValidCPF("529.982.247-25") {
		t.Fatal("expected CPF to be valid")
	}
	if IsValidCPF("111.111.111-11") {
		t.Fatal("expected CPF to be invalid")
	}
}

func TestIsValidCNPJ(t *testing.T) {
	if !IsValidCNPJ("04.252.011/0001-10") {
		t.Fatal("expected CNPJ to be valid")
	}
	if IsValidCNPJ("00.000.000/0000-00") {
		t.Fatal("expected CNPJ to be invalid")
	}
}

