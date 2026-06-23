## Backend Engineering Technical Assessment

## Struktur project

```
.
├── main.go                    # entrypoint, menyatukan seluruh service
├── internal/
│   └── booking/               # Section 1 — Race Condition
└── pkg/response/              # standardized response envelope
```

## Cara menjalankan

Membutuhkan Go 1.22+.

```bash
    go run main.go
    # Server berjalan di http://localhost:8000
```

## Mencoba tiap skenario dari Posman/Dll

**Section 1 — Race condition (coba kirim 2 request bersamaan, stok awal = 1):**

```bash
curl -X POST http://localhost:8000/api/v1/tickets/book \
  -H "Content-Type: application/json" \
  -d '{"ticket_id":"VIP-1","user_id":"user-A"}' &
curl -X POST http://localhost:8000/api/v1/tickets/book \
  -H "Content-Type: application/json" \
  -d '{"ticket_id":"VIP-1","user_id":"user-B"}' &
wait
# Hanya salah satu yang mengembalikan 201, yang lain 409 "tiket sudah habis terjual"
```

## Cara testing

```bash
# Semua test
go test ./... -v

# Test spesifik per skenario
go test ./internal/booking/...   -v   # Section 1
```
