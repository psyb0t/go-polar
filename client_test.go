package polar_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/psyb0t/go-polar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testToken     = "EXAMPLE-DO-NOT-USE"
	testProductID = "00000000-0000-4000-8000-000000000001"
	testProduct   = `{
		"id":"00000000-0000-4000-8000-000000000001",
		"created_at":"2026-10-17T00:00:00Z","modified_at":null,
		"trial_interval":null,"trial_interval_count":null,
		"name":"Example product","description":null,"visibility":"public",
		"recurring_interval":null,"recurring_interval_count":null,
		"meter_interval":null,"meter_interval_count":null,
		"is_recurring":false,"is_archived":false,
		"organization_id":"00000000-0000-4000-8000-000000000002",
		"metadata":{},"is_deletable":true,"prices":[],"benefits":[],
		"medias":[],"attached_custom_fields":[]
	}`
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *polar.ClientWithResponses {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	auth := polar.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+testToken)

			return nil
		},
	)
	client, err := polar.NewClientWithResponses(srv.URL, auth)
	require.NoError(t, err)

	return client
}

func TestProductsListPagination(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/products/", r.URL.Path)
		assert.Equal(t, "Bearer "+testToken, r.Header.Get("Authorization"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		assert.Equal(t, []string{testProductID}, r.URL.Query()["id"])
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		assert.NoError(t, err)
		assert.Contains(t, []int{1, 2}, page)

		w.Header().Set("Content-Type", "application/json")

		body := map[string]any{
			"items":      []json.RawMessage{json.RawMessage(testProduct)},
			"pagination": map[string]int{"total_count": 2, "max_page": 2},
		}
		assert.NoError(t, json.NewEncoder(w).Encode(body))
	})

	const pageCount = 2

	limit := 1
	ids := []string{testProductID}

	for page := 1; page <= pageCount; page++ {
		res, err := client.ProductsListWithResponse(
			context.Background(),
			&polar.ProductsListParams{Page: &page, Limit: &limit, Id: &ids},
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode())
		require.NotNil(t, res.JSON200)
		require.Len(t, res.JSON200.Items, 1)
		assert.Equal(t, testProductID, res.JSON200.Items[0].Id)
		assert.Nil(t, res.JSON200.Items[0].Description)
		assert.Equal(t, pageCount, res.JSON200.Pagination.MaxPage)
		assert.Equal(t, pageCount, res.JSON200.Pagination.TotalCount)
	}
}

func TestProductsGet(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/products/"+testProductID, r.URL.Path)
		assert.Equal(t, "Bearer "+testToken, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(testProduct))
		assert.NoError(t, err)
	})
	res, err := client.ProductsGetWithResponse(
		context.Background(),
		testProductID,
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Equal(t, testProductID, res.JSON200.Id)
	assert.Equal(t, "Example product", res.JSON200.Name)
	assert.False(t, res.JSON200.IsRecurring)
}

func TestProductsGetNotFound(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(
			`{"error":"ResourceNotFound","detail":"Product not found."}`,
		))
		assert.NoError(t, err)
	})
	res, err := client.ProductsGetWithResponse(
		context.Background(),
		testProductID,
	)
	require.NoError(t, err)
	assert.Nil(t, res.JSON200)
	require.NotNil(t, res.JSON404)
	assert.Equal(t, "Product not found.", res.JSON404.Detail)
}

func TestProductsGetMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"id":`))
		assert.NoError(t, err)
	})
	_, err := client.ProductsGetWithResponse(
		context.Background(),
		testProductID,
	)
	require.Error(t, err)
}

func TestCustomerTaxIDTuple(t *testing.T) {
	t.Parallel()

	var customer polar.CustomerIndividual

	err := json.Unmarshal([]byte(`{"tax_id":["EXAMPLE","us_ein"]}`), &customer)
	require.NoError(t, err)
	require.NotNil(t, customer.TaxId)
	assert.Equal(t, []any{"EXAMPLE", "us_ein"}, *customer.TaxId)
}

func TestProductsListMultipleIDs(t *testing.T) {
	t.Parallel()

	ids := []string{testProductID, "00000000-0000-4000-8000-000000000002"}
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, ids, r.URL.Query()["id"])
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(
			`{"items":[],"pagination":{"total_count":0,"max_page":1}}`,
		))
		assert.NoError(t, err)
	})
	res, err := client.ProductsListWithResponse(
		context.Background(),
		&polar.ProductsListParams{Id: &ids},
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Empty(t, res.JSON200.Items)
}

func TestProductsGetUnauthorized(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		token string
	}{
		{name: "missing"},
		{name: "malformed", token: "EXAMPLE-DO-NOT-USE"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, tc.token, r.Header.Get("Authorization"))
					w.WriteHeader(http.StatusUnauthorized)
				},
			))
			t.Cleanup(server.Close)
			client, err := polar.NewClientWithResponses(server.URL)
			require.NoError(t, err)
			res, err := client.ProductsGetWithResponse(
				context.Background(),
				testProductID,
				func(_ context.Context, req *http.Request) error {
					req.Header.Set("Authorization", tc.token)

					return nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, res.StatusCode())
			assert.Nil(t, res.JSON200)
		})
	}
}

func TestMeterQuantityFloat64(t *testing.T) {
	t.Parallel()

	var meter polar.MeterQuantity

	err := json.Unmarshal([]byte(`{"quantity":1.23456789012345}`), &meter)
	require.NoError(t, err)
	assert.InDelta(t, 1.23456789012345, meter.Quantity, 0)
}

func TestOrdersGenerateInvoiceRawResponse(t *testing.T) {
	t.Parallel()

	const responseBody = `{"queued":true,"request":{"id":"EXAMPLE"}}`

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/orders/"+testProductID+"/invoice", r.URL.Path)
		assert.Equal(t, "Bearer "+testToken, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	})
	res, err := client.OrdersGenerateInvoiceWithResponse(
		context.Background(),
		testProductID,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, res.StatusCode())
	require.NotNil(t, res.JSON202)
	assert.Equal(t, responseBody, string(*res.JSON202))
}
