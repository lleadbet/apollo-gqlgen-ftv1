package trace

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
)

type (
	Tracer struct {
		ClientName  string
		Version     string
		Hostname    string
		TreeBuilder *TreeBuilder
		ShouldTrace bool
	}

	TracingExtension struct {
		Report
	}
)

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
	graphql.OperationInterceptor
	graphql.Transport
} = &Tracer{}

func (Tracer) ExtensionName() string {
	return "ApolloFederatedTracingV1"
}

func (Tracer) Validate(graphql.ExecutableSchema) error {
	return nil
}

func (Tracer) Supports(r *http.Request) bool {
	return true
}

func (t *Tracer) Do(w http.ResponseWriter, r *http.Request, exec graphql.GraphExecutor) {
	if r.Header.Get("apollo-federation-include-trace") == "ftv1" {
		t.ShouldTrace = true
	}
}

func (t *Tracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	if !t.ShouldTrace {
		return next(ctx)
	}

	t.TreeBuilder = NewTreeBuilder()
	t.TreeBuilder.StartTimer()

	return next(ctx)
}
func (t *Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	if !t.ShouldTrace {
		return next(ctx)
	}

	t.TreeBuilder.WillResolveField(ctx)

	return next(ctx)
}

func (t *Tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	if !t.ShouldTrace {
		return next(ctx)
	}

	defer func() {
		rc := graphql.GetOperationContext(ctx)

		t.TreeBuilder.StopTimer()

		td := &TracingExtension{
			Report: Report{
				Header: &ReportHeader{
					Hostname:           t.ClientName,
					AgentVersion:       t.Version,
					ExecutableSchemaId: GetMD5Hash(rc.RawQuery),
				},
				TracesPerQuery: map[string]*TracesAndStats{
					rc.RawQuery: {
						Trace: []*Trace{t.TreeBuilder.Trace},
					},
				},
			},
		}

		graphql.RegisterExtension(ctx, "ftv1", td)
	}()

	resp := next(ctx)
	return resp
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
