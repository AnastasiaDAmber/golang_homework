//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	t.Run("empty data", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(""), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	t.Run("invalid JSON line", func(t *testing.T) {
		badData := `{"Email":"user@ok.com"}
invalid_json_line
{"Email":"user@ok.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(badData), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"ok.com": 2}, result)
	})

	t.Run("email without @", func(t *testing.T) {
		badEmail := `{"Email":"justtext.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(badEmail), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	t.Run("mixed case domain", func(t *testing.T) {
		mixed := `{"Email":"user@Example.CoM"}`
		result, err := GetDomainStat(bytes.NewBufferString(mixed), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"example.com": 1}, result)
	})

	t.Run("reader error", func(t *testing.T) {
		reader := &errorReader{}
		result, err := GetDomainStat(reader, "com")
		require.Error(t, err)
		require.Nil(t, result)
	})
}

// errorReader — вспомогательная структура для имитации ошибки чтения
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")

}
