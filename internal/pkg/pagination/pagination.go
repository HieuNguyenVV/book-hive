package pagination

import (
	"encoding/base64"
	"fmt"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
)

const DefaultPaginationLimit int = 10
const DefaultPaginationOffset int = 0
const MaxPaginationLimit int = 100

func GetNextToken(limit, offset *int, total int) string {
	next := ""
	if total == 0 {
		return next
	}

	normalizedLimit, normalizedOffset := normalizeLimitOffset(limit, offset)

	nextOffset := normalizedOffset + normalizedLimit
	if hasNextOffset(nextOffset, total) {
		next = getOffsetToken(normalizedLimit, nextOffset)
	}

	return base64.URLEncoding.EncodeToString([]byte(next))
}

func GetPrevToken(limit, offset *int, total int) string {
	prev := ""
	if total == 0 {
		return prev
	}

	normalizedLimit, normalizedOffset := normalizeLimitOffset(limit, offset)
	prevOffset := normalizedOffset - normalizedLimit
	if hasPrevOffset(prevOffset) {
		prev = getOffsetToken(normalizedLimit, prevOffset)
	}

	return base64.URLEncoding.EncodeToString([]byte(prev))
}

func normalizeLimitOffset(limit, offset *int) (int, int) {
	normalizedLimit := DefaultPaginationLimit
	normalizedOffset := DefaultPaginationOffset

	if limit != nil && *limit > 0 {
		normalizedLimit = *limit
	}
	if offset != nil && *offset >= 0 {
		normalizedOffset = *offset
	}
	return normalizedLimit, normalizedOffset
}

func hasNextOffset(nextOffset, total int) bool {
	return nextOffset < total
}

func hasPrevOffset(prevOffset int) bool {
	return prevOffset >= 0
}

func getOffsetToken(limit, offset int) string {
	return fmt.Sprintf("?limit=%d&offset=%d", limit, offset)
}

func IsValidParams(limit, offset *int) bool {
	if limit == nil || offset == nil {
		return true
	}

	if (limit == nil && offset != nil) || (limit != nil && offset == nil) {
		return false
	}
	if *limit <= 0 || *offset < 0 {
		return false
	}

	return true
}

func GetLimitOffsetFromToken(token string, limit, offset *int) error {
	decodedBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return errors.ErrInvalidValue.Wrap(err).Reform("invalid token")
	}

	_, err = fmt.Sscanf(string(decodedBytes), "?limit=%d&offset=%d", limit, offset)
	if err != nil {
		return errors.ErrInvalidValue.Wrap(err).Reform("get limit offset error")
	}
	if !IsValidParams(limit, offset) {
		return errors.ErrInvalidValue.Wrap(err).Reform("invalid pagination parameters")
	}

	return nil
}
