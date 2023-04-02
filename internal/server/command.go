package server

// commandLength is the length of the command
const commandLength = 12

// commandToBytes converts a string command to a byte array
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

// bytesToCommand converts a byte array to a string command
func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return string(command)
}

// extractCommand extracts the command from the payload
func extractCommand(request []byte) []byte {
	return request[:commandLength]
}
