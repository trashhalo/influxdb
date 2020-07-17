package tenant_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	influxdb "github.com/influxdata/influxdb/servicesv2"
	"github.com/influxdata/influxdb/servicesv2/api"
	"github.com/influxdata/influxdb/servicesv2/tenant"
	itesting "github.com/influxdata/influxdb/servicesv2/testing"
	"go.uber.org/zap/zaptest"
)

//  pushdowngrouprule
//  pushdownreadtagkeysrule
//  registerphysicalrules
func initHttpOrgService(f itesting.OrganizationFields, t *testing.T) (influxdb.OrganizationService, string, func()) {
	t.Helper()

	s, stCloser, err := tenant.NewTestBoltStore(t)
	if err != nil {
		t.Fatal(err)
	}

	storage := tenant.NewStore(s)
	svc := tenant.NewService(storage)
	ctx := context.Background()
	for _, o := range f.Organizations {
		if err := svc.CreateOrganization(ctx, o); err != nil {
			t.Fatalf("failed to populate organizations, Err: %v", err)
		}
	}

	handler := tenant.NewHTTPOrgHandler(zaptest.NewLogger(t), svc)
	r := chi.NewRouter()
	r.Mount(handler.Prefix(), handler)
	server := httptest.NewServer(r)
	httpClient, err := api.NewHTTPClient(server.URL, "", false)
	if err != nil {
		t.Fatal(err)
	}

	orgClient := tenant.OrgClientService{
		Client: httpClient,
	}

	return &orgClient, "http_tenant", func() {
		server.Close()
		stCloser()
	}
}

func TestOrgService(t *testing.T) {
	itesting.OrganizationService(initHttpOrgService, t)
}