package datasync

import "testing"

func TestApplyUpdate_OutOfOrder(t *testing.T) {
	svc := NewService()
	
	applied3 := svc.ApplyUpdate(Availability{TicketID: "VIP-1", Quantity: 7, Version: 3})
	if !applied3 {
		t.Fatalf("(version 3) to be applied")
	}

	applied1 := svc.ApplyUpdate(Availability{TicketID: "VIP-1", Quantity: 2, Version: 2})
	if applied1 {
		t.Fatalf("expected first-arriving update (version 2) to be applied")
	}

	applied2 := svc.ApplyUpdate(Availability{TicketID: "VIP-1", Quantity: 5, Version: 1})
	if applied2 {
		t.Fatalf("expected stale update (version 1) to be rejected")
	}

	final, _ := svc.Get("VIP-1")
	if final.Quantity != 7 {
		t.Fatalf("expected final quantity to remain 2 (latest version), got %d", final.Quantity)
	}
	
	t.Logf("Final availability for VIP-1: %+v", final)
}
