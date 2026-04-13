package document

import (
	"unicode"

	"github.com/brazilian-utils/go/cnpj"
	"github.com/brazilian-utils/go/cpf"
)

func Normalize(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if unicode.IsDigit(r) {
			out = append(out, r)
		}
	}
	return string(out)
}

func IsValidCPF(value string) bool  { return cpf.IsValid(Normalize(value)) }
func IsValidCNPJ(value string) bool { return cnpj.IsValid(Normalize(value)) }
