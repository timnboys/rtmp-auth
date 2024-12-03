package store

import (
	"github.com/timnboys/rtmp-auth/storage"
)

type Backend interface {
	Read() (*storage.State, error)
	Write(state *storage.State) error
}
