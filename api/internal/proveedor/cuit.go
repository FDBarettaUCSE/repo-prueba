package proveedor

import (
	"regexp"
	"strings"
)

var soloDigitos = regexp.MustCompile(`\D`)

// EsCUITValido implementa el algoritmo de módulo 11 que usa AFIP/ARCA para
// validar el dígito verificador del CUIT. El CUIT tiene 11 dígitos: 2 de tipo,
// 8 de número de documento/orden y 1 verificador. Acepta el CUIT con o sin
// guiones (XX-XXXXXXXX-X).
func EsCUITValido(cuit string) bool {
	cuit = strings.ReplaceAll(cuit, "-", "")

	if len(cuit) != 11 {
		return false
	}

	for _, caracter := range cuit {
		if caracter < '0' || caracter > '9' {
			return false
		}
	}

	multiplicadores := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

	suma := 0

	for i := 0; i < 10; i++ {
		suma += int(cuit[i]-'0') * multiplicadores[i]
	}

	resto := suma % 11
	verificador := 11 - resto

	if verificador == 11 {
		verificador = 0
	}

	if verificador == 10 {
		return false
	}

	return int(cuit[10]-'0') == verificador
}
