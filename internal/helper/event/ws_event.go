package event

import "github.com/HieuNguyenVV/book-hive/internal/server/model"

func WSEventFormat(eventName model.WSEvent, message []byte) []byte {
	payload := append([]byte(eventName), message...)
	return append(payload, '\n')
}
