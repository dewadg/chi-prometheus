package chi_prometheus

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_middleware_Handle(t *testing.T) {
	type fields struct {
		serviceName      string
		requestCounter   *Mockcounter
		latencyHistogram *Mockhistogram
		counter          *MockCounter
		histogram        *MockObserver
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name    string
		fields  fields
		request func(url string) *http.Request
		before  func(f fields)
		assert  func(f fields)
	}{
		{
			name: "success: record correct metrics",
			fields: fields{
				serviceName:      "test",
				requestCounter:   NewMockcounter(ctrl),
				latencyHistogram: NewMockhistogram(ctrl),
				counter:          NewMockCounter(ctrl),
				histogram:        NewMockObserver(ctrl),
			},
			request: func(url string) *http.Request {
				req, _ := http.NewRequest(http.MethodGet, url, nil)
				return req
			},
			before: func(f fields) {
				f.requestCounter.EXPECT().
					WithLabelValues("test", "200", "GET", "/").
					Return(f.counter)

				f.counter.EXPECT().Inc().Times(1)

				f.latencyHistogram.EXPECT().
					WithLabelValues("test", "200", "GET", "/").
					Return(f.histogram)

				f.histogram.EXPECT().Observe(gomock.Any()).Times(1)
			},
			assert: func(f fields) {

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &middleware{
				serviceName:      tt.fields.serviceName,
				requestCounter:   tt.fields.requestCounter,
				latencyHistogram: tt.fields.latencyHistogram,
			}

			r := chi.NewRouter()
			r.Use(m.Handle)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, "ok")
			})

			s := httptest.NewServer(r)
			defer s.Close()

			req := tt.request(s.URL)

			tt.before(tt.fields)
			s.Client().Do(req)
			tt.assert(tt.fields)
		})
	}
}
