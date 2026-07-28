# wisp

A small LSM-tree key-value store written in Go.

## How it works

- Writes go to an in-memory skip list and a write-ahead log (WAL) for crash recovery.
- When the mutable skip list gets big enough, it's frozen and flushed to disk as an SSTable, while a new mutable skip list takes over writes.
- SSTables are stored as length-prefixed binary records with a sparse index (one entry per ~4096 bytes) for fast lookups.
- Reads check the mutable skip list first, then the frozen skip list, then fall back to searching SSTables newest to oldest.
- On restart, the WAL is replayed to rebuild the in-memory state.

## Status

Actively being built. Current pieces: skip list + WAL + SSTable flush/search, WAL rotation, and crash-safe flushing. Tombstones (proper deletes across SSTables) are next.

## Run

```
go run ./cmd
```

Server starts on `:8080`.
