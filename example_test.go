package polar_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-polar"
)

//nolint:testableexamples // requires a real access token, compile check only
func Example() {
	auth := polar.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"Authorization",
				"Bearer "+os.Getenv("POLAR_ACCESS_TOKEN"),
			)

			return nil
		},
	)

	client, err := polar.NewClientWithResponses(
		"https://api.polar.sh",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ProductsListWithResponse(
		context.Background(),
		&polar.ProductsListParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
