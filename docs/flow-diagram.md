# Combined Flow Diagram — Ticket Booking System

Diagram ini menggabungkan kelima skenario (Section 1–5) menjadi satu flow
end-to-end: dari user book tiket, sistem menangani lonjakan traffic,
mengirim data ke accounting service pihak ketiga, menerima webhook payment,
dan menyinkronkan ketersediaan tiket ke sistem lain.

```mermaid
flowchart TD
    U[User / Client] -->|POST /tickets/book| A[Booking Service]

    subgraph S1["Section 1: Race Condition"]
        A --> A1{Lock ticket_id<br/>cek stok > 0?}
        A1 -->|Ya, atomik kurangi stok| A2[Simpan transaksi]
        A1 -->|Tidak| A3[Response: Tiket Habis - 409]
    end

    A2 -->|enqueue| B[Ingestion Queue<br/>buffered channel]

    subgraph S2["Section 2: High Traffic Processing"]
        B --> B1{Queue penuh?}
        B1 -->|Ya| B2[Response: 503 Service Unavailable<br/>backpressure]
        B1 -->|Tidak| B3[Worker Pool<br/>N goroutine]
        B3 --> B4[Simpan ke Database]
        B4 -->|gagal setelah retry| B5[Dead Letter Queue]
        B4 -->|sukses| C[Trigger pengiriman ke Accounting]
    end

    subgraph S3["Section 3: External API Integration"]
        C --> C1[Simpan ke Outbox<br/>status=pending]
        C1 --> C2{Circuit Breaker<br/>Allow?}
        C2 -->|Open| C3[Skip, tunggu dispatcher berikutnya]
        C2 -->|Closed/HalfOpen| C4[Call Accounting API]
        C4 -->|500 / timeout| C5[Retry exponential backoff]
        C5 -->|max attempt tercapai| C1
        C4 -->|Sukses| C6[Outbox status=sent]
        D[Background Dispatcher<br/>periodic scan] -.->|pending entries| C2
    end

    E[Accounting Service<br/>Pihak Ketiga] -->|webhook async| F[Webhook Receiver]

    subgraph S4["Section 4: Duplicate Webhook Request"]
        F --> F1{idempotency_key<br/>sudah ada?}
        F1 -->|Ya - duplicate retry| F2[Return response sukses lama<br/>tanpa proses ulang]
        F1 -->|Tidak - baru| F3[Lock + Simpan ke<br/>transaction_payment]
    end

    G[Ticket System] -->|update availability<br/>ticket_id, qty, version| H[Sistem Tujuan]

    subgraph S5["Section 5: Data Synchronization"]
        H --> H1{incoming.version ><br/>current.version?}
        H1 -->|Ya| H2[Apply update]
        H1 -->|Tidak - stale/out-of-order| H3[Abaikan update]
    end

    C6 -.->|notify| E
```

## Ringkasan alur per section

1. **Section 1 (Booking):** setiap request booking melewati critical
   section yang dilindungi lock per `ticket_id`, menjamin hanya satu
   pemenang saat stok = 1.
2. **Section 2 (Ingestion):** request masuk ke queue, di-ack cepat, lalu
   diproses asinkron oleh worker pool; backpressure mencegah overload.
3. **Section 3 (External):** transaksi sukses dicatat di outbox sebelum
   dikirim, dilindungi circuit breaker + retry, dispatcher background
   menjamin pengiriman akhirnya berhasil (at-least-once).
4. **Section 4 (Webhook):** idempotency key mencegah payload duplikat
   tersimpan dua kali walau diterima bersamaan.
5. **Section 5 (Sync):** version/sequence number memastikan update yang
   stale (out-of-order) tidak menimpa data yang lebih baru.
