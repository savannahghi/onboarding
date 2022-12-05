module github.com/savannahghi/onboarding

go 1.18

require (
	cloud.google.com/go/firestore v1.6.1
	cloud.google.com/go/pubsub v1.23.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/99designs/gqlgen v0.17.10
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/brianvoe/gofakeit/v5 v5.11.2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/google/uuid v1.3.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/imroc/req v0.3.2
	github.com/savannahghi/converterandformatter v0.0.11
	github.com/savannahghi/enumutils v0.0.3
	github.com/savannahghi/errorcodeutil v0.0.6
	github.com/savannahghi/feedlib v0.0.6
	github.com/savannahghi/firebasetools v0.0.17
	github.com/savannahghi/interserviceclient v0.0.18
	github.com/savannahghi/profileutils v0.0.27
	github.com/savannahghi/pubsubtools v0.0.3
	github.com/savannahghi/scalarutils v0.0.4
	github.com/savannahghi/serverutils v0.0.6
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/vektah/gqlparser/v2 v2.4.5
	go.opencensus.io v0.23.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.21.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858
)

require (
	cloud.google.com/go v0.102.1 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/errorreporting v0.2.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/logging v1.5.0 // indirect
	cloud.google.com/go/monitoring v1.5.0 // indirect
	cloud.google.com/go/profiler v0.3.0 // indirect
	cloud.google.com/go/storage v1.23.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.13 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/aws/aws-sdk-go v1.44.45 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/pprof v0.0.0-20220608213341-c488b8fa1db3 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/lithammer/shortuuid v3.0.0+incompatible // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/prometheus v0.36.2 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opentelemetry.io/contrib v1.7.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.7.0 // indirect
	go.opentelemetry.io/otel/metric v0.30.0 // indirect
	go.opentelemetry.io/otel/sdk v1.7.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/oauth2 v0.0.0-20220628200809-02e64fa58f26 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220627191245-f75cf1eec38b // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.11 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/api v0.86.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220628213854-d9e0b6570c03 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
