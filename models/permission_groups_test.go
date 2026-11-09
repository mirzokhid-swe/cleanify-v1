package models

import "testing"

func groupByName(gs []PermissionGroup, name string) *PermissionGroup {
	for i := range gs {
		if gs[i].Group == name {
			return &gs[i]
		}
	}
	return nil
}

func TestGroupPermissionsOrderAndHidden(t *testing.T) {
	perms := []Permission{
		{ID: "1", Slug: "get_users_list", Name: "Ro'yxatni ko'rish", Group: "Foydalanuvchilar"},
		{ID: "2", Slug: "get_leads", Name: "Leadlarni ko'rish", Group: "Leadlar"},
		{ID: "3", Slug: "get_lead_statuses", Name: "x", Group: "Lead statuslari"},
		{ID: "4", Slug: "get_orders_list", Name: "Ro'yxatni ko'rish", Group: "Buyurtmalar"},
		{ID: "5", Slug: "kpi", Name: "KPI", Group: "Statistikalar"},
		{ID: "6", Slug: "get_order_item_type", Name: "Ko'rish", Group: "Mahsulot turlari"},
		{ID: "7", Slug: "setting_get_order_statuses", Name: "Statuslarni ko'rish", Group: "[Sozlama] Buyurtma Statuslari"},
		{ID: "8", Slug: "create_order_item", Name: "Qo'shish", Group: "Buyurtma mahsulotlari"},
		{ID: "9", Slug: "change_status_to_6", Name: "x", Group: "Buyurtmalar"},
		{ID: "10", Slug: "something", Name: "Boshqa", Group: "Yangi guruh"},
	}
	got := GroupPermissions(perms, nil)
	want := []string{"Buyurtmalar", "Holatni o'zgartirish", "Buyurtma buyumlari", "Ko'rsatkichlar", "Sozlamalar", "Foydalanuvchilar", "Yangi guruh"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i, w := range want {
		if got[i].Group != w {
			t.Fatalf("pos %d: got %q want %q", i, got[i].Group, w)
		}
	}
	if s := groupByName(got, "Sozlamalar"); len(s.Items) != 2 || s.Items[0].Slug != "get_order_item_type" {
		t.Fatal(s)
	}
	if groupByName(got, "Leadlar") != nil || groupByName(got, "Lead statuslari") != nil {
		t.Fatal("hidden groups must be dropped")
	}
	if h := groupByName(got, "Holatni o'zgartirish"); h.Items[0].Name != "Holat 6" || h.Items[0].ID != 9 {
		t.Fatal(h)
	}
}

func TestGroupPermissionsStatusNames(t *testing.T) {
	perms := []Permission{
		{ID: "1", Slug: "change_status_to_6", Name: "x", Group: "Buyurtmalar"},
		{ID: "2", Slug: "change_status_to_99", Name: "x", Group: "Buyurtmalar"},
		{ID: "3", Slug: "change_status_to_2", Name: "x", Group: "Buyurtmalar"},
		{ID: "4", Slug: "change_status_to_washed", Name: "x", Group: "Buyurtma mahsulotlari"},
		{ID: "5", Slug: "change_status_to_prepare", Name: "x", Group: "Buyurtma mahsulotlari"},
	}
	statuses := []OrderStatus{
		{Number: 2, Name: "Olingan", Order: 2},
		{Number: 6, Name: "Topshirildi", Order: 1},
	}
	got := GroupPermissions(perms, statuses)
	h := groupByName(got, "Holatni o'zgartirish")
	if h == nil || len(h.Items) != 2 {
		t.Fatal(got)
	}
	if h.Items[0].Name != "Topshirildi" || h.Items[1].Name != "Olingan" {
		t.Fatal("must be ordered by status Order and named by status", h.Items)
	}
	b := groupByName(got, "Buyurtma buyumlari")
	if b == nil || len(b.Items) != 2 || b.Items[0].Name != "Yuvilgan" || b.Items[1].Name != "Tayyor" {
		t.Fatal(b)
	}
}

func TestGroupPermissionsNameOverrides(t *testing.T) {
	perms := []Permission{
		{ID: "1", Slug: "edit_order", Name: "Taxrirlash", Group: "Buyurtmalar"},
		{ID: "2", Slug: "edit_lead_x", Name: "Taxrirlash", Group: "Mijozlar"},
		{ID: "3", Slug: "show_client", Name: "Klient ma'lumotini batafsil ko'rish", Group: "Mijozlar"},
	}
	got := GroupPermissions(perms, nil)
	if n := groupByName(got, "Buyurtmalar").Items[0].Name; n != "Tahrirlash" {
		t.Fatal(n)
	}
	m := groupByName(got, "Mijozlar").Items
	if m[0].Name != "Tahrirlash" || m[1].Name != "Klient ma'lumotini batafsil ko'rish" {
		t.Fatal(m)
	}
}
