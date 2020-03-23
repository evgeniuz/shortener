package shortener

const digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Base62Encode(n uint64) []byte {
	var a [11]byte // uint64 fits into 11 base62 digits
	i, b := len(a), uint64(62)

	for n >= b {
		i--
		q := n / b
		a[i] = digits[uint(n-q*b)]
		n = q
	}

	i--
	a[i] = digits[uint(n)]

	return a[i:]
}
