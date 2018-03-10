package regrev

func AllCharacters() []byte {
	result := []byte{}
	result = append(result, AlphaUpper()...)
	result = append(result, AlphaLower()...)
	result = append(result, Digits()...)
	return result
}

func AlphaUpper() []byte {
	return []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
}

func AlphaLower() []byte {
	return []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
}

func Digits() []byte {
	return []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
}

func Whitespace() []byte {
	return []byte{' ', '\t', '\r', '\n', '\f', '\v'}
}
