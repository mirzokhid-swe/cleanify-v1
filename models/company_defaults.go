package models

// Yangi korxona bilan birga yaratiladigan standartlar (POST /company).
// Spec: cleanify-docs/design/pencil/README.md, 13-bo'lim.

type DefaultStatus struct {
	Number      int
	Order       int
	Name        string
	Slug        string
	Color       string // order_statuses.color: matn ko'rinishidagi {"r":..,"g":..,"b":..}
	Description string
}

// Omborda raqami 8: ruxsat faqat change_status_to_8 uchun bor.
var DefaultOrderStatuses = []DefaultStatus{
	{1, 1, "Olish kerak", "change_status_to_1", `{"r":255,"g":0,"b":0}`, "Bu status mijoz tomonidan yangi zayavka tushganini anglatadi"},
	{2, 2, "Olingan", "change_status_to_2", `{"r":90,"g":198,"b":255}`, "Bu status buyurtma mijozdan kuryer tomonidan olinganini anglatadi."},
	{3, 3, "Yuvilgan", "change_status_to_3", `{"r":33,"g":57,"b":210}`, "Bu status buyurtma yuvilganini bildiradi"},
	{4, 4, "Tayyor", "change_status_to_4", `{"r":101,"g":135,"b":42}`, "Bu status buyurtma tayyorligini anglatadi"},
	{5, 5, "Oborish kerak", "change_status_to_5", `{"r":160,"g":5,"b":240}`, "Buyurtma oborishga tayyor"},
	{6, 6, "Topshirildi", "change_status_to_6", `{"r":107,"g":114,"b":128}`, "Buyurtma topshirildi"},
	{8, 7, "Omborda", "change_status_to_8", `{"r":224,"g":138,"b":0}`, "Buyurtma omborga olindi"},
}

type DefaultItemType struct {
	Name  string
	Unit  string
	Price float64
}

var DefaultItemTypes = []DefaultItemType{
	{"12 000", "area", 12000},
	{"13 000", "area", 13000},
	{"14 000", "area", 14000},
	{"15 000", "area", 15000},
	{"Odeyalo 50", "piece", 50000},
	{"Odeyalo 60", "piece", 60000},
	{"Ko'rpa", "length", 25000},
	{"Parda", "weight", 30000},
}

// DefaultRole: Slugs bo'sh bo'lsa (Admin) rol scope = 'company' bo'lgan hamma ruxsatni oladi.
type DefaultRole struct {
	Name  string
	Slugs []string
}

var DefaultRoles = []DefaultRole{
	{"Admin", nil},
	{"Menejer", []string{
		"get_orders_list", "edit_order", "delete_order", "attach_courier_to_order",
		"add_payment", "delete_transaction", "change_payment_status",
		"edit_order_item", "delete_order_item",
		"get_work_volume_by_day", "get_financial_metrics", "get_status_metrics",
		"manage_all_transactions",
		"get_clients_list", "show_client", "edit_client", "delete_client",
		"change_status_to_1", "change_status_to_2", "change_status_to_3", "change_status_to_4",
		"change_status_to_5", "change_status_to_6", "change_status_to_8", "change_status_to_99",
		"change_status_to_washed", "change_status_to_prepare",
	}},
	{"Kuryer", []string{
		"get_orders_list", "attach_courier_to_order", "edit_order", "add_payment",
		"change_status_to_1", "change_status_to_2", "change_status_to_5", "change_status_to_6",
	}},
	{"Yuvuvchi", []string{
		"get_orders_list", "change_status_to_3", "change_status_to_4",
		"change_status_to_washed", "change_status_to_prepare", "edit_order_item",
	}},
}

type DefaultPaymentPurpose struct {
	Name string
	Slug string
	Type int // 1 kirim, 2 chiqim
}

var DefaultPaymentPurposes = []DefaultPaymentPurpose{
	{"buyurtma", "orders", 1},
	{"Kassa", "kassa", 1},
	{"Oylik", "salary", 2},
}
