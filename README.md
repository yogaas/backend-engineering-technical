## Backend Engineering Technical Assessment

## Struktur project

```
.
├── main.go                    # entrypoint, menyatukan seluruh service
├── internal/
│   ├── booking/               # Section 1 — Race Condition
│   ├── ingestion/             # Section 2 — High Traffic Processing
│   ├── external/              # Section 3 — External API Integration
│   ├── webhook/               # Section 4 — Duplicate Request
│   └── datasync/              # Section 5 — Data Synchronization
├── pkg/response/              # standardized response envelope
└── docs/flow-diagram.md       # diagram flow gabungan (Mermaid)
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

**Section 2 — High traffic (kirim banyak transaksi):**

```bash
curl -X POST http://localhost:8000/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-A","amount":150000}'
# Response 202 Accepted -> id transaksi
curl http://localhost:8000/api/v1/transactions/ING-1/status
```

**Section 3 — External API integration (simulasi kirim ke accounting service):**

```bash
curl -X POST http://localhost:8000/api/v1/transactions/TX-1/send-to-accounting
# Akan retry otomatis jika gagal (2 percobaan pertama disimulasikan gagal)
```

**Section 4 — Duplicate webhook (kirim payload sama 2x):**

```bash
curl -X POST http://localhost:8000/api/v1/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"idempotency_key":"idem-001","transaction_id":"TX-1","amount":150000,"status":"PAID"}'
curl -X POST http://localhost:8000/api/v1/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"idempotency_key":"idem-001","transaction_id":"TX-1","amount":150000,"status":"PAID"}'
# Request kedua dikembalikan sebagai "duplicate, diabaikan" tanpa double-insert
```

**Section 5 — Data sync (kirim update out-of-order):**

```bash
curl -X POST http://localhost:8000/api/v1/sync/ticket-availability \
  -H "Content-Type: application/json" \
  -d '{"ticket_id":"VIP-1","quantity":2,"version":2}'
curl -X POST http://localhost:8000/api/v1/sync/ticket-availability \
  -H "Content-Type: application/json" \
  -d '{"ticket_id":"VIP-1","quantity":5,"version":1}'
# Update kedua (version lebih lama) diabaikan, quantity tetap 2
```

## Cara testing

```bash
# Semua test
go test ./... -v

# Test spesifik per skenario
go test ./internal/booking/...   -v   # Section 1
go test ./internal/ingestion/... -v   # Section 2
go test ./internal/external/...  -v   # Section 3
go test ./internal/webhook/...   -v   # Section 4
go test ./internal/datasync/...  -v   # Section 5
```
