package m3u8_test

import (
	"m3u8/m3u8"
	"testing"
)

func CheckType(t *testing.T, p m3u8.Playlist) {
	t.Logf("%T implements Playlist interface OK\n", p)
}

// Create new media playlist.
func TestNewMediaPlaylist(t *testing.T) {
	_, e := m3u8.NewMediaPlaylist(1, 2)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
}
