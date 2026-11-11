package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type PermissionItem struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type PermissionGroup struct {
	Group string           `json:"group"`
	Items []PermissionItem `json:"items"`
}

const statusGroup = "Holatni o'zgartirish"
const itemsGroup = "Buyurtma buyumlari"

var GroupOrder = []string{"Buyurtmalar", "Holatni o'zgartirish", "Buyurtma buyumlari", "Mijozlar", "Ko'rsatkichlar", "Tranzaksiyalar", "Sozlamalar", "Foydalanuvchilar", "Xodimlar", "Telegram guruhlar"}

var HiddenGroups = []string{"Leadlar", "Lead statuslari"}

// GroupNames maps a DB group to the screen group; a group not listed keeps its own name.
var GroupNames = map[string]string{
	"Buyurtma mahsulotlari":         "Buyurtma buyumlari",
	"Mahsulot turlari":              "Sozlamalar",
	"[Sozlama] Buyurtma Statuslari": "Sozlamalar",
	"Statistikalar":                 "Ko'rsatkichlar",
}

var NameOverrides = map[string]string{
	"get_orders_list":            "Ro'yxat",
	"edit_order":                 "Tahrirlash",
	"delete_order":               "O'chirish",
	"attach_courier_to_order":    "Kuryer biriktirish",
	"add_payment":                "To'lov kiritish",
	"delete_transaction":         "To'lovni o'chirish",
	"change_payment_status":      "To'lov holatini o'zgartirish",
	"edit_order_item":            "Tahrirlash",
	"delete_order_item":          "O'chirish",
	"get_clients_list":           "Ro'yxat",
	"edit_client":                "Tahrirlash",
	"delete_client":              "O'chirish",
	"get_users_list":             "Ro'yxat",
	"user_create":                "Qo'shish",
	"edit_user":                  "Tahrirlash",
	"delete_user":                "O'chirish",
	"manage_roles":               "Rollarni boshqarish",
	"get_order_item_type":        "Turlar: ro'yxat",
	"create_order_item_type":     "Turlar: qo'shish",
	"edit_order_item_type":       "Turlar: tahrirlash",
	"delete_order_item_type":     "Turlar: o'chirish",
	"setting_get_order_statuses": "Holatlar: ro'yxat",
	"setting_edit_order_status":  "Holatlar: tahrirlash",
	"get_work_volume_by_day":     "Ish hajmi",
	"get_financial_metrics":      "Moliya",
}

// GroupPermissions turns the flat permission list into screen groups. It does no DB access.
// statuses are the company's own statuses; nil means "no company": status names come from the slug.
func GroupPermissions(perms []Permission, statuses []OrderStatus) []PermissionGroup {
	groups := map[string][]PermissionItem{}
	var seen []string // screen groups in first-seen order
	add := func(group string, item PermissionItem) {
		if _, ok := groups[group]; !ok {
			seen = append(seen, group)
		}
		groups[group] = append(groups[group], item)
	}

	type statusItem struct {
		item  PermissionItem
		order int
		num   int
	}
	var statusItems []statusItem

	for _, p := range perms {
		hidden := false
		for _, g := range HiddenGroups {
			if p.Group == g {
				hidden = true
			}
		}
		if hidden {
			continue
		}
		id, _ := strconv.Atoi(p.ID)
		item := PermissionItem{ID: id, Slug: p.Slug, Name: p.Name}

		if p.Slug == "change_status_to_washed" {
			item.Name = "Yuvilgan"
			add(itemsGroup, item)
			continue
		}
		if p.Slug == "change_status_to_prepare" {
			item.Name = "Tayyor"
			add(itemsGroup, item)
			continue
		}
		if strings.HasPrefix(p.Slug, "change_status_to_") {
			n, err := strconv.Atoi(strings.TrimPrefix(p.Slug, "change_status_to_"))
			if err == nil {
				if len(statuses) == 0 {
					item.Name = fmt.Sprintf("Holat %d", n)
					statusItems = append(statusItems, statusItem{item, n, n})
					continue
				}
				for _, s := range statuses {
					if s.Number == n {
						item.Name = s.Name
						statusItems = append(statusItems, statusItem{item, int(s.Order), n})
						break
					}
				}
				continue // no status with this number: dropped
			}
		}

		if name, ok := NameOverrides[p.Slug]; ok {
			item.Name = name
		} else {
			item.Name = strings.ReplaceAll(item.Name, "Taxrirlash", "Tahrirlash")
		}
		group := p.Group
		if name, ok := GroupNames[group]; ok {
			group = name
		}
		add(group, item)
	}

	sort.SliceStable(statusItems, func(i, j int) bool { return statusItems[i].order < statusItems[j].order })
	for _, s := range statusItems {
		add(statusGroup, s.item)
	}

	var result []PermissionGroup
	used := map[string]bool{}
	for _, g := range GroupOrder {
		if items, ok := groups[g]; ok && len(items) > 0 {
			result = append(result, PermissionGroup{Group: g, Items: items})
			used[g] = true
		}
	}
	for _, g := range seen {
		if !used[g] && len(groups[g]) > 0 {
			result = append(result, PermissionGroup{Group: g, Items: groups[g]})
		}
	}
	return result
}
