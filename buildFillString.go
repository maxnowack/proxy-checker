package main

func buildFillString(length int) string {
	str := ""
	for i := 0; i < length; i++ {
		str = str + "0"
	}
	return str
}
