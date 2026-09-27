package proveedor

import (
	"regexp"
	"strconv"
)

var soloDigitos = regexp.MustCompile(`\D`)

// EsCUITValido implementa el algoritmo de módulo 11 que usa AFIP/ARCA para
// validar el dígito verificador del CUIT. El CUIT tiene 11 dígitos: 2 de tipo,
// 8 de número de documento/orden y 1 verificador. Acepta el CUIT con o sin
// guiones (XX-XXXXXXXX-X).
func EsCUITValido(cuit string) bool {
	digitos := soloDigitos.ReplaceAllString(cuit, "")
	if len(digitos) != 11 {
		return false
	}

	multiplicadores := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

	suma := 0
	for i := 0; i < 10; i++ {
		n, err := strconv.Atoi(string(digitos[i]))
		if err != nil {
			return false
		}
		suma += n * multiplicadores[i]
	}

	resto := suma % 11
	verificadorEsperado := 11 - resto
	if verificadorEsperado == 11 {
		verificadorEsperado = 0
	}
	if verificadorEsperado == 10 {
		// Ningún CUIT real da 10 como resultado de este cálculo.
		return false
	}

	ultimoDigito, err := strconv.Atoi(string(digitos[10]))
	if err != nil {
		return false
	}
	return ultimoDigito == verificadorEsperado
}
