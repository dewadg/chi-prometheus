# Chi Prometheus

Prometheus RED metrics middleware for Chi

## Usage

```go
r := chi.NewRouter()
r.Use(chiprometheus.New(chiprometheus.Config{
	ServiceName:  "my-service",
	ServiceLabel: "service",
	MetricPrefix: "http"
}))
```
