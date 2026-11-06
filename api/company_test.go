package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakeCompanies struct {
	repo.CompanyStorageI
	settings  map[string]json.RawMessage
	blocked   bool
	blockedBy int64
	upserts   int
}

func (f *fakeCompanies) GetShow(id string) (models.CompanyShow, error) {
	if id != companyA {
		return models.CompanyShow{}, sql.ErrNoRows
	}
	return models.CompanyShow{
		CompanyListItem: models.CompanyListItem{ID: id, Name: "A", UsersCount: 3, IsBlocked: f.blocked},
		Settings:        models.MergeSettings(f.settings),
	}, nil
}

func (f *fakeCompanies) UpsertSetting(companyID, key string, value json.RawMessage, userID int64) error {
	if f.settings == nil {
		f.settings = map[string]json.RawMessage{}
	}
	f.settings[key] = value
	f.upserts++
	return nil
}

func (f *fakeCompanies) SetBlocked(companyID string, blocked bool, userID int64) error {
	f.blocked, f.blockedBy = blocked, userID
	return nil
}

func (f *fakeCompanies) GetList() ([]models.CompanyListItem, error) {
	return []models.CompanyListItem{{ID: companyA, Name: "A"}}, nil
}

type companyStorage struct {
	fakeStorage
	companies *fakeCompanies
}

func (s companyStorage) Company() repo.CompanyStorageI { return s.companies }

func newCompanyRouter(superAdmin bool) (*gin.Engine, *fakeCompanies) {
	gin.SetMode(gin.TestMode)
	companies := &fakeCompanies{}
	h := handlers.Handler{Stg: companyStorage{companies: companies}}
	r := gin.New()
	setUser := func(c *gin.Context) { c.Set("user", models.User{ID: 8, IsSuperAdmin: superAdmin}) }
	companyRoutes(r.Group("/company"), &h, setUser)
	return r, companies
}

func TestCompanyRoutesNeedSuperAdmin(t *testing.T) {
	r, _ := newCompanyRouter(false)
	for _, p := range []struct{ method, path string }{
		{"GET", "/company"}, {"GET", "/company/" + companyA},
		{"PATCH", "/company/" + companyA + "/settings"}, {"POST", "/company/" + companyA + "/block"},
	} {
		w := do(r, p.method, p.path, `{"key":"kpi_enabled","value":true}`)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s %s: code=%d", p.method, p.path, w.Code)
		}
	}
}

func TestUpdateCompanySettingValidates(t *testing.T) {
	r, companies := newCompanyRouter(true)
	for _, body := range []string{
		`{"key":"payment_tolerance","value":-5}`,
		`{"key":"volume_mode","value":"x"}`,
		`{"key":"kpi_enabled","value":"true"}`,
		`{"key":"nope","value":1}`,
		`{"key":"kpi_enabled"}`,
	} {
		w := do(r, "PATCH", "/company/"+companyA+"/settings", body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: code=%d body=%s", body, w.Code, w.Body.String())
		}
	}
	if companies.upserts != 0 {
		t.Fatal("invalid values must not be stored")
	}
	w := do(r, "PATCH", "/company/"+companyA+"/settings", `{"key":"kpi_enabled","value":true}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"kpi_enabled":true`) || !strings.Contains(w.Body.String(), `"volume_mode":"type"`) {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	w = do(r, "PATCH", "/company/"+companyB+"/settings", `{"key":"kpi_enabled","value":true}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown company: code=%d", w.Code)
	}
}

func TestBlockAndUnblockCompany(t *testing.T) {
	r, companies := newCompanyRouter(true)
	w := do(r, "POST", "/company/"+companyA+"/block", "")
	if w.Code != http.StatusOK || !companies.blocked || companies.blockedBy != 8 || !strings.Contains(w.Body.String(), `"is_blocked":true`) {
		t.Fatalf("block: code=%d blocked=%v by=%d body=%s", w.Code, companies.blocked, companies.blockedBy, w.Body.String())
	}
	w = do(r, "POST", "/company/"+companyA+"/unblock", "")
	if w.Code != http.StatusOK || companies.blocked {
		t.Fatalf("unblock: code=%d blocked=%v", w.Code, companies.blocked)
	}
	w = httptest.NewRecorder()
	if w = do(r, "GET", "/company", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"users_count"`) {
		t.Fatalf("list: code=%d body=%s", w.Code, w.Body.String())
	}
}
